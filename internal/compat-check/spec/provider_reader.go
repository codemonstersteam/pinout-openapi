package spec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/compat-check/config"
)

// ErrProviderUnreachable is raised when the provider master spec cannot be obtained over
// HTTP(S): a non-2xx response, or a failure before any response (network/DNS/connection
// refused). error.code PROVIDER_UNREACHABLE, exit 3 (contracts.md "Error model"). Applies only
// to the URL branch (source.Kind() == config.ProviderSourceHTTP) — a missing/unreadable
// **file** source is a config/exit-2 concern (ErrSpecUnreadable), per the frozen
// api-specification/exit-codes.md note ("A provider file that is missing/unreadable surfaces
// as SPEC_UNREADABLE (exit 2)") and config.schema.json's provider.spec_path description.
var ErrProviderUnreachable = errors.New("provider spec unreachable")

// ErrProviderTimeout is raised when the HTTP(S) GET of spec_url does not complete within
// settings.timeout. error.code PROVIDER_TIMEOUT, exit 3 (contracts.md "Error model").
var ErrProviderTimeout = errors.New("provider spec fetch timeout")

// ErrProviderParse is raised when the fetched/read provider document does not parse or does not
// validate as OpenAPI 3.x. error.code PROVIDER_PARSE_ERROR, exit 3 (contracts.md "Error model").
var ErrProviderParse = errors.New("provider spec parse error")

// maxProviderSpecBytes is the payload budget (http-io skill) for the provider HTTP response: an
// explicit, logged rejection instead of an unbounded read of an external body. A spec this large
// is not a legitimate OpenAPI document for this tool's use case; the fetch is aborted rather than
// silently truncated (http-io "no silent caps").
const maxProviderSpecBytes = 20 << 20 // 20 MiB

// ProviderSpecReader is the slice's one outbound-HTTP node (module-tree.md, contracts.md). It
// resolves the provider master spec from exactly one origin, source.Kind(): an unauthenticated
// HTTP(S) GET of source.URL(), bounded by timeout, or a local read of source.Path() — then
// honest-reuses kin-openapi to parse + resolve $refs. Pipe — not unit-tested; proven by
// component scenarios CS6 (unreachable) / CS7 (timeout) / CS8 (unparseable). Never mutates the
// master spec (read-only, GET/read-only file access).
//
// Load budget (http-io skill): this reader issues exactly one unauthenticated GET per Load call
// — no fan-out, no invented retries — well inside any provider rate limit, so pacing rung 1
// (sequential, no extra pause beyond the bounding timeout) applies trivially; a 429 or transient
// failure surfaces as ErrProviderUnreachable rather than a silent retry loop, per the "no generic
// HTTP error, no unbounded retry" rule. Payload budget: the response body is capped at
// maxProviderSpecBytes.
type ProviderSpecReader struct {
	timeout time.Duration
}

// NewProviderSpecReader constructs a reader bounded by timeout. The caller (head/wiring) supplies
// the run's settings.timeout (config.Settings.Timeout(), range 1..600s, config.schema.json) — the
// HTTP GET branch is bounded by it explicitly (http-io skill: "HTTP timeout explicit").
func NewProviderSpecReader(timeout time.Duration) ProviderSpecReader {
	return ProviderSpecReader{timeout: timeout}
}

// Load resolves the provider master spec per source.Kind() and honest-reuses kin-openapi to
// parse it, resolving local/external $refs (IsExternalRefsAllowed).
func (r ProviderSpecReader) Load(source config.ProviderSource) (Spec, error) {
	var (
		data     []byte
		location *url.URL
		err      error
	)

	switch source.Kind() {
	case config.ProviderSourceHTTP:
		data, location, err = r.fetch(source.URL())
	default:
		data, location, err = r.read(source.Path())
	}
	if err != nil {
		return Spec{}, err
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromDataWithPath(data, location)
	if err != nil {
		return Spec{}, fmt.Errorf("%w: %s: %s", ErrProviderParse, location, err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		return Spec{}, fmt.Errorf("%w: %s: %s", ErrProviderParse, location, err)
	}

	return newSpec(doc), nil
}

// fetch performs the single unauthenticated GET (http-io load budget: N=1, no retries), bounded
// by r.timeout, and returns the raw body plus the resolved request URL (used as the $ref-relative
// location by the kin-openapi loader — the same role LoadFromURI's location plays).
func (r ProviderSpecReader) fetch(specURL string) ([]byte, *url.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %s", ErrProviderUnreachable, specURL, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, fmt.Errorf("%w: %s: %s", ErrProviderTimeout, specURL, err)
		}
		return nil, nil, fmt.Errorf("%w: %s: %s", ErrProviderUnreachable, specURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, nil, fmt.Errorf("%w: %s: HTTP %d", ErrProviderUnreachable, specURL, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxProviderSpecBytes+1))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, fmt.Errorf("%w: %s: %s", ErrProviderTimeout, specURL, err)
		}
		return nil, nil, fmt.Errorf("%w: %s: %s", ErrProviderUnreachable, specURL, err)
	}
	if len(body) > maxProviderSpecBytes {
		return nil, nil, fmt.Errorf("%w: %s: response exceeds %d byte payload budget", ErrProviderUnreachable, specURL, maxProviderSpecBytes)
	}

	return body, req.URL, nil
}

// read loads the provider spec from a local file. Missing/unreadable -> ErrSpecUnreadable (the
// same sentinel ConsumerSpecReader.Load uses) -> SPEC_UNREADABLE, exit 2 — NOT
// ErrProviderUnreachable/exit 3. This is mandated by the frozen contract: exit-codes.md ("A
// provider file that is missing/unreadable surfaces as SPEC_UNREADABLE (exit 2)") and
// config.schema.json's provider.spec_path description ("Missing/unreadable -> SPEC_UNREADABLE
// (exit 2); unparseable -> PROVIDER_PARSE_ERROR (exit 3)"). A provider file that reads but does
// not parse/validate still falls through to ErrProviderParse below (PROVIDER_PARSE_ERROR,
// exit 3), matching that same sentence.
func (r ProviderSpecReader) read(specPath string) ([]byte, *url.URL, error) {
	data, err := os.ReadFile(specPath)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %s", ErrSpecUnreadable, specPath, err)
	}
	return data, &url.URL{Path: filepath.ToSlash(specPath)}, nil
}

// Package provider — slice-01-validate: how the provider OpenAPI spec is acquired and
// parsed (ticket 09, io: http, contracts.md §SpecLoader.Load, ADR-0001). One SpecLoader
// unifies spec_path (local file) XOR spec_url (HTTP GET) behind one kin-openapi-backed
// object — file-vs-url is an internal strategy, not two objects. Pure I/O pipe: no
// domain logic (that's compare/) and NOT unit-tested (its four failure branches are
// component scenarios 3-6 — module-tree.md).
package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/validate/domain"
)

// providerTokenEnv — env var carrying the Bearer token for a private spec_url
// (FRD "Auth to a private spec_url"; contracts.md §SpecLoader.Load). Optional: an
// empty/unset value means no Authorization header (public spec_url). Never read from
// config/git — secret-from-env only (code-style).
const providerTokenEnv = "PINOUT_PROVIDER_TOKEN"

// SpecLoader acquires + parses the provider OpenAPI spec: p.SpecPath (local file) XOR
// p.SpecURL (HTTP GET), then $ref-resolves via kin-openapi. Deps are encapsulated
// inside (kin-openapi loader + *http.Client + timeout + bearer token) — the head sees
// only Load(p) (contracts.md §SpecLoader.Load). Load budget: one GET per invocation
// (N=1, not a fan-out) — no pacing rung needed (http-io). Payload budget: the whole
// provider spec is the acquisition target itself, no whitelist to trim.
type SpecLoader struct {
	// timeout bounds the HTTP branch only (settings.timeout, seconds > 0 — guaranteed
	// in-range by NewConfig). The file branch has no network bound.
	timeout time.Duration
}

// NewSpecLoader constructs the production SpecLoader bounded by timeoutSeconds
// (cfg.Settings.Timeout, wired by the caller after NewConfig validates it > 0).
func NewSpecLoader(timeoutSeconds int) SpecLoader {
	return SpecLoader{timeout: time.Duration(timeoutSeconds) * time.Second}
}

// Load acquires and parses the provider spec described by p. Antecedent: p valid —
// exactly one of SpecPath/SpecURL set (guaranteed by NewConfig). Consequent: Ok
// ProviderSpec (parsed *openapi3.T, $ref resolved). Fail: ErrFileNotFound (spec_path
// missing), ErrParse (bad OpenAPI/YAML/JSON), ErrHTTP (spec_url unreachable/non-2xx),
// ErrTimeout (spec_url fetch > timeout) — contracts.md §SpecLoader.Load.
func (l SpecLoader) Load(p domain.ProviderConfig) (domain.ProviderSpec, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	var doc *openapi3.T
	var err error

	switch {
	case p.SpecPath != "":
		doc, err = loadFromFile(loader, p.SpecPath)
	case p.SpecURL != "":
		doc, err = l.loadFromURL(loader, p.SpecURL)
	default:
		// Unreachable given a valid antecedent (NewConfig guarantees exactly-one
		// source) — no domain sentinel fits an antecedent breach here, so surface
		// it as a parse-shaped configuration error rather than invent a new code.
		return domain.ProviderSpec{}, fmt.Errorf("%w: provider config has neither spec_path nor spec_url", domain.ErrParse)
	}
	if err != nil {
		return domain.ProviderSpec{}, err
	}
	return domain.ProviderSpec{Doc: doc}, nil
}

// loadFromFile reads+parses a local spec_path via kin-openapi. A missing file maps to
// ErrFileNotFound; any other read/parse failure maps to ErrParse.
func loadFromFile(loader *openapi3.Loader, path string) (*openapi3.T, error) {
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s: %v", domain.ErrFileNotFound, path, err)
		}
		return nil, fmt.Errorf("%w: %s: %v", domain.ErrParse, path, err)
	}
	return doc, nil
}

// loadFromURL fetches spec_url (bounded by l.timeout, optional Bearer token from env)
// and parses the fetched bytes via kin-openapi. HTTP failures (unreachable/non-2xx)
// map to ErrHTTP; exceeding l.timeout maps to ErrTimeout; a fetched-but-unparseable
// body maps to ErrParse.
func (l SpecLoader) loadFromURL(loader *openapi3.Loader, specURL string) (*openapi3.T, error) {
	data, err := fetchSpec(specURL, l.timeout)
	if err != nil {
		return nil, err
	}
	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", domain.ErrParse, specURL, err)
	}
	return doc, nil
}

// fetchSpec performs the bounded HTTP GET (Authorization: Bearer $PINOUT_PROVIDER_TOKEN
// if set), classifying the failure per http-io: a client-side timeout maps to
// ErrTimeout, everything else (dial/connection/non-2xx) maps to ErrHTTP.
//
// The bound is enforced by a request-scoped context (context.WithTimeout), NOT by
// http.Client{Timeout}: when the client-level timeout fires mid-flight it surfaces as
// a connection-level error whose Timeout() reports false, so it was mis-classified as
// ErrHTTP. A context deadline unwraps to context.DeadlineExceeded, which errors.Is
// detects deterministically (the net.Error Timeout() check is kept as a fallback for
// transport-level deadlines, e.g. a dial timeout).
func fetchSpec(specURL string, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", domain.ErrHTTP, specURL, err)
	}
	if token := os.Getenv(providerTokenEnv); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		var netErr interface{ Timeout() bool }
		if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			return nil, fmt.Errorf("%w: %s: %v", domain.ErrTimeout, specURL, err)
		}
		return nil, fmt.Errorf("%w: %s: %v", domain.ErrHTTP, specURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: %s: status %d", domain.ErrHTTP, specURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", domain.ErrHTTP, specURL, err)
	}
	return body, nil
}

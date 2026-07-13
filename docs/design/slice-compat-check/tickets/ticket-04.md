---
id: 04
type: module
slice: slice-compat-check
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-compat-check/contracts.md, docs/design/slice-compat-check/module-tree.md]
outputs: [internal/compat-check/spec/consumer_reader.go, internal/compat-check/spec/provider_reader.go, internal/compat-check/spec/spec.go]
io: http
skills: [http-io]
---

### TICKET S-compat-check.04 — `spec/`: load + `$ref`-resolve OpenAPI 3.x (consumer file, provider file|HTTP)

**io:** `http` (header value — this package's ONE outbound-HTTP node, `ProviderSpecReader`, is what routes
the extra skill; `ConsumerSpecReader` is plain file I/O with no dedicated remote sub-skill per
`module-tree.md`)   →  skills: `http-io`

**Context (only this package — `internal/compat-check/spec/`):**
- **`consumer_reader.go` — `ConsumerSpecReader.Load(path string) -> Result<Spec, Error>`.** Deps: —.
  Read the consumer file + `kin-openapi` load (**honest reuse** — do not reimplement OpenAPI parsing/
  validation), local+external `$ref` resolved, `allOf` flattened. I/O pipe — not unit-tested.
  Failures: `ErrSpecUnreadable` → `SPEC_UNREADABLE` (exit 2, CS4) · `ErrSpecParse` → `SPEC_PARSE_ERROR`
  (exit 2, CS5).
- **`provider_reader.go` — `ProviderSpecReader.Load(source config.ProviderSource) -> Result<Spec, Error>`.**
  Deps: —. `source.Kind ∈ {http, file}` (from `config.NewProviderSource`, ticket 03). **HTTP(S) GET
  `spec_url`, unauthenticated, bounded by `settings.timeout` — OR read `spec_path`** — then `kin-openapi`
  load. Never mutates the master spec (read-only). Apply the `http-io` skill's load/payload-budget
  discipline to the HTTP branch (single unauthenticated GET per invocation, response bounded by
  `settings.timeout`, no retries invented beyond what `http-io` prescribes). Failures: `ErrProviderUnreachable`
  → `PROVIDER_UNREACHABLE` (exit 3, CS6, non-2xx/unreachable) · `ErrProviderTimeout` →
  `PROVIDER_TIMEOUT` (exit 3, CS7) · `ErrProviderParse` → `PROVIDER_PARSE_ERROR` (exit 3, CS8).
- **`spec.go` — `Spec`** (wraps the `kin-openapi` model; `HasOperation(path, method)`, `Operation(path,
  method)` query methods). Types + queries only, not unit-tested (thin wrapper over honest-reused parser).

**Unit tests:** none — this whole package is I/O pipes + a thin query wrapper (proven by component
scenarios CS4/CS5/CS6/CS7/CS8 via the fixer, not this ticket).

**Dependencies:** imports `internal/compat-check/config` (ticket 03) for the `config.ProviderSource` type
(`provider_reader.go`'s input). Does not import `compare/`, `report/`, `cli/`, or the package-root files.

**Subagent instruction:** implement `spec.go` → `consumer_reader.go` → `provider_reader.go` (apply
`http-io` budget rules to the GET) → run build/vet. No test file to write (no unit tests in this package).
Do not touch `config/`, `compare/`, `report/`, `cli/`, or the package-root files.

**Verify:** `go build ./internal/compat-check/spec/... && go vet ./internal/compat-check/spec/...`

**Acceptance:** package builds and vets clean; no unit tests expected (I/O pipe package — component
scenarios CS4–CS8 prove it later, via the fixer, not this ticket's deliverable).

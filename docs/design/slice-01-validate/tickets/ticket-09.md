---
id: 09
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/adr/0001-specloader-unifies-file-and-http.md, .agent/planner/frd.md]
outputs: [internal/validate/provider/loader.go]
io: http
skills: [http-io]
---

### TICKET 09 — slice-01-validate/SpecLoader.Load: acquire+parse provider spec (I/O, kin-openapi)

**io:** http → skills: **`http-io`** (outbound metered fetch of `spec_url`; timeout & payload budgets,
provider spec as the frozen machine contract, real-protocol stub for component tests — ADR-0001). Pure
I/O pipe, **not unit-tested**.

**Context (only this module):**
- contract (`contracts.md` §SpecLoader.Load): `Load(p: ProviderConfig) -> Result[ProviderSpec, Error]`.
  Acquire the provider OpenAPI — `spec_path` = local file **XOR** `spec_url` = HTTP GET with
  `Authorization: Bearer $PINOUT_PROVIDER_TOKEN` (from **env**, optional; secret never in config/git),
  bounded by `settings.timeout` — then parse + resolve `$ref` via **`kin-openapi`**. `spec_path` XOR
  `spec_url` is an **internal strategy** (one loader, not split — ADR-0001). No domain logic (that is
  `compare`). Deps = — (kin-openapi loader + `*http.Client` + timeout + token encapsulated inside).
  - antecedent: `ProviderConfig` valid (exactly-one source — guaranteed by `NewConfig`).
  - consequent: Ok `ProviderSpec` (parsed `*openapi3.T`, `$ref` resolved). Fail `ErrFileNotFound` →
    `FILE_NOT_FOUND`/3; `ErrParse` → `PARSE_ERROR`/3; `ErrHTTP` (unreachable/non-2xx) → `HTTP_ERROR`/3;
    `ErrTimeout` (fetch > timeout) → `TIMEOUT_ERROR`/3.
- unit tests: **none** (I/O pipe — its four failure branches are component scenarios).
- component scenario(s) to green: 3 (`FILE_NOT_FOUND`), 4 (`PARSE_ERROR`), 5 (`HTTP_ERROR`),
  6 (`TIMEOUT_ERROR`) — greened later; scenarios 5/6 drive the real-protocol HTTP stub.

**Dependencies:** `ProviderConfig`/`ProviderSpec` types + sentinels (ticket 03); `kin-openapi` (add to go.mod).

**Subagent instruction:** apply `http-io` (compute load/payload budgets, use the provider spec as the
frozen contract) → implement `SpecLoader.Load` in `internal/validate/provider/loader.go` →
`go build`/`vet` → done. Touch no other module; do not drive component scenarios GREEN.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `provider` builds and vets clean; no units (I/O pipe). Component GREEN is @fagan's step.

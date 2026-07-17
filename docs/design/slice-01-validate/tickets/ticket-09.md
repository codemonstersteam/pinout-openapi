---
id: 09
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/adr/0001-specloader-unifies-file-and-http.md, docs/design/slice-01-validate/adr/0005-late-specloader-construction-via-factory.md, .agent/planner/frd.md]
outputs: [internal/validate/provider/loader.go]
io: http
skills: [http-io]
---

### TICKET 09 — slice-01-validate/SpecLoader.Load: acquire+parse provider spec (I/O, kin-openapi)

> **REWORK (ADR-0005) — edit the EXISTING `internal/validate/provider/loader.go`.** Neither
> `NewSpecLoader(timeoutSeconds int)` nor `Load(ProviderConfig) -> Result[ProviderSpec, Error]`
> changes signature — the code here is **unchanged**. What moves is the **call site** of the
> constructor (out of `register.go`, into the head via a factory — tickets 17 & 16). This ticket only
> re-affirms the loader contract under the new construction lifecycle; if `loader.go` already matches
> the contract below, it stays as-is (verify build/vet).

**io:** http → skills: **`http-io`** (outbound metered fetch of `spec_url`; timeout & payload budgets,
provider spec as the frozen machine contract, real-protocol stub for component tests — ADR-0001). Pure
I/O pipe, **not unit-tested**.

**Context (only this module):**
- contract (`contracts.md` §SpecLoader.Load): `Load(p: ProviderConfig) -> Result[ProviderSpec, Error]`
  **(signature unchanged)**. Acquire the provider OpenAPI — `spec_path` = local file **XOR** `spec_url`
  = HTTP GET with `Authorization: Bearer $PINOUT_PROVIDER_TOKEN` (from **env**, optional; secret never
  in config/git), bounded by the loader's `timeout` — then parse + resolve `$ref` via **`kin-openapi`**.
  `spec_path` XOR `spec_url` is an **internal strategy** (one loader, not split — ADR-0001). No domain
  logic (that is `compare`).
- **construction lifecycle (ADR-0005 — the ONLY delta):** `provider.NewSpecLoader(timeoutSeconds int)`
  keeps its signature. It is **no longer called eagerly in `register.go`** with a hardcoded
  `defaultProviderTimeoutSeconds = 30`. Instead `register.go` passes a **factory closure**
  `func(s Settings) SpecLoader { return provider.NewSpecLoader(s.Timeout) }` into `Deps.BuildSpecLoader`
  (ticket 17), and the head calls that factory **after `NewConfig`** so the loader's `timeout` is the
  real `cfg.Settings.Timeout` (ticket 16). **This module's code does not know or care** — it just gets
  a real `timeoutSeconds` at construction. Consequence: with the true `settings.timeout`, the HTTP
  branch can now genuinely emit `ErrTimeout` (component scenario 6).
  - antecedent: `ProviderConfig` valid (exactly-one source — guaranteed by `NewConfig`); loader built
    with `timeoutSeconds = cfg.Settings.Timeout` (> 0).
  - consequent: Ok `ProviderSpec` (parsed `*openapi3.T`, `$ref` resolved). Fail `ErrFileNotFound` →
    `FILE_NOT_FOUND`/3; `ErrParse` → `PARSE_ERROR`/3; `ErrHTTP` (unreachable/non-2xx) → `HTTP_ERROR`/3;
    `ErrTimeout` (fetch > timeout) → `TIMEOUT_ERROR`/3.
- unit tests: **none** (I/O pipe — its four failure branches are component scenarios).
- component scenario(s) to green: 3 (`FILE_NOT_FOUND`), 4 (`PARSE_ERROR`), 5 (`HTTP_ERROR`),
  6 (`TIMEOUT_ERROR`) — greened later; scenarios 5/6 drive the real-protocol HTTP stub. Scenario 6 is
  now realizable because the loader receives the real `settings.timeout` (ADR-0005).

**Dependencies:** `ProviderConfig`/`ProviderSpec` types + sentinels (ticket 03); `kin-openapi` (already
in go.mod from the prior implementation).

**Subagent instruction:** apply `http-io` (confirm load/payload budgets, provider spec as the frozen
contract) → keep `NewSpecLoader(timeoutSeconds int)` + `SpecLoader.Load` in
`internal/validate/provider/loader.go` conformant to the contract above (no signature change) →
`go build`/`vet` → done. Do **not** move the call site here (that is tickets 16/17). Touch no other
module; do not drive component scenarios GREEN.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `provider` builds and vets clean; `NewSpecLoader(int)`/`Load(...)` signatures
intact; no units (I/O pipe). Component GREEN is @fagan's step.

---
id: 10
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md]
outputs: [internal/validate/provider/operation.go, internal/validate/provider/operation_test.go]
io: none
skills: []
---

### TICKET 10 — slice-01-validate/DeriveProviderOperation: navigate spec → {requires, provides} (R1)

**io:** none → skills: [] (pure logic — **unit-tested by formula**).

**Context (only this module):**
- contract (`contracts.md` §DeriveProviderOperation):
  `DeriveProviderOperation(spec: ProviderSpec, ref: OperationRef) -> Result[ProviderOperation, NotPresent]`.
  Navigate `paths[path][method]`; derive `{ requires, provides }` over **body + parameters
  (path/query/header)** — required request fields/params (contravariant) and response body props
  (covariant). Implements **R1** (operation exists). Deps = —.
  - antecedent: `ProviderSpec` parsed.
  - consequent: Ok `ProviderOperation{ Requires, Provides }`; `NotPresent` signals R1 →
    `OP_NOT_IN_PROVIDER` (a **verdict**, exit 1 — not an error), recorded by `CompareOperation`.
- **unit tests: 2** (`module-tree.md` formula) — 1 happy + 1 branch: operation absent in provider
  (→ R1 `OP_NOT_IN_PROVIDER`).
- component scenario(s) to green: none (R1 verdict is UNIT-covered here — ADR-0002).

**Dependencies:** `ProviderSpec`/`ProviderOperation`/`OperationRef` types (ticket 03); `kin-openapi`.

**Subagent instruction:** write the 2 unit tests → implement `DeriveProviderOperation` in
`internal/validate/provider/operation.go` → run units → green → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/... && go test ./internal/validate/provider/...`.

**Acceptance:** package `provider` builds/vets clean; 2 unit tests green.

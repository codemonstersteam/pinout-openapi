---
id: 12
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md]
outputs: [internal/validate/compare/comparison.go, internal/validate/compare/comparison_test.go]
io: none
skills: []
---

### TICKET 12 — slice-01-validate/NewComparison: unite the trio into one domain entity

**io:** none → skills: [] (uniting constructor — **unit-tested by formula**).

**Context (only this module):**
- contract (`contracts.md` §NewComparison):
  `NewComparison(cfg: Config, consumed: ConsumedContract, spec: ProviderSpec) -> Result[Comparison, Error]`.
  Sanctioned uniting constructor (2+ entities → one node): build
  `Comparison{ ScopedOps, Consumed, Spec, Provenance }` — the scope (`cfg.operations`) intersected with
  the consumed per-op data and the provider spec. Deps = —.
  - antecedent: all three inputs valid.
  - consequent: Ok one `Comparison`. No failing antecedent (inputs already valid by construction).
- **unit tests: 1** (`module-tree.md` formula) — 1 happy (uniting constructor over already-valid inputs).
- component scenario(s) to green: none.

**Dependencies:** `Config`/`ConsumedContract`/`ProviderSpec`/`Comparison`/`Provenance` types (ticket 03).

**Subagent instruction:** write the 1 happy unit test → implement `NewComparison` in
`internal/validate/compare/comparison.go` → run unit → green → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/... && go test ./internal/validate/compare/...`.

**Acceptance:** package `compare` builds/vets clean; 1 unit test green.

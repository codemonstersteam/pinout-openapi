---
id: 13
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03, 10, 11]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md]
outputs: [internal/validate/compare/compare.go, internal/validate/compare/compare_test.go]
io: none
skills: []
---

### TICKET 13 — slice-01-validate/CompareContracts: fold R1–R4 over scoped operations

**io:** none → skills: [] (pure fold logic — **unit-tested by formula**). The **loop lives here**, not
in the head.

**Context (only this module):**
- contract (`contracts.md` §CompareContracts): `CompareContracts(c: Comparison) -> ComparisonOutcome`.
  For each scoped operation call `DeriveProviderOperation` (ticket 10) + `CompareOperation` (ticket 11),
  `concatMap` the violations; compute `uncovered_operations[]` (provider ops outside `cfg.operations`,
  informational); carry `Provenance` through. Deps = —. No failure path.
  - antecedent: valid `Comparison`.
  - consequent: `ComparisonOutcome{ Violations, UncoveredOps, Provenance }` (`compatible ⇔ Violations == []`).
- **unit tests: 3** (`module-tree.md` formula) — 1 happy + 2 branches: multi-operation fold
  accumulation, `uncovered_operations[]` detection.
- component scenario(s) to green: none directly (verdict codes are UNIT — ADR-0002); its ∅-violations
  path underpins component scenario 1 (happy), greened later.

**Dependencies:** `DeriveProviderOperation` (ticket 10) + `CompareOperation` (ticket 11);
`Comparison`/`ComparisonOutcome`/`Violation` types (ticket 03).

**Subagent instruction:** write the 3 unit tests → implement `CompareContracts` in
`internal/validate/compare/compare.go` → run units → green → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/... && go test ./internal/validate/compare/...`.

**Acceptance:** package `compare` builds/vets clean; 3 unit tests green.

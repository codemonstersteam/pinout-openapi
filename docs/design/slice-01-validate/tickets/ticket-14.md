---
id: 14
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md, api-specification/report.schema.json]
outputs: [internal/validate/report/build.go, internal/validate/report/build_test.go]
io: none
skills: []
---

### TICKET 14 — slice-01-validate/FoldReport: outcome → Report DTO

**io:** none → skills: [] (pure logic — **unit-tested by formula**).

**Context (only this module):**
- contract (`contracts.md` §FoldReport): `FoldReport(o: ComparisonOutcome) -> Report`.
  Shape the frozen `report.schema.json` DTO — `schema_version="1.0"`, `compatible = Violations == []`,
  `errors[] = Violations`, top-level `provenance` echo, informational `uncovered_operations[]`. Deps = —.
  No failure path.
  - antecedent: valid `ComparisonOutcome`.
  - consequent: schema-valid `Report`.
- **unit tests: 3** (`module-tree.md` formula) — 1 happy + 2 branches: non-empty `errors` ⇒
  `compatible=false`, `uncovered_operations[]` populated.
- component scenario(s) to green: none directly; feeds component scenario 1 (happy report), greened later.

**Dependencies:** `ComparisonOutcome`/`Report`/`Provenance`/`Violation` types (ticket 03).

**Subagent instruction:** write the 3 unit tests → implement `FoldReport` in
`internal/validate/report/build.go` → run units → green → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/... && go test ./internal/validate/report/...`.

**Acceptance:** package `report` builds/vets clean; 3 unit tests green.

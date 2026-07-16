---
id: 11
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md, docs/design/slice-01-validate/adr/0002-verdict-is-unit-not-component.md]
outputs: [internal/validate/compare/rules.go, internal/validate/compare/rules_test.go]
io: none
skills: []
---

### TICKET 11 — slice-01-validate/CompareOperation (+typesMatch): R2/R3/R4 per operation

**io:** none → skills: [] (pure logic core — **unit-tested by formula**). Ported verbatim from
`sandbox/check.mjs::compareOp` (not reinvented).

**Context (only this module):**
- contract (`contracts.md` §CompareOperation):
  `CompareOperation(consumedOp: ConsumedOperation, providerOp: ProviderOperation|NotPresent) -> []Violation`.
  Over **body + parameters**: **R1** absent ⇒ `OP_NOT_IN_PROVIDER`; **R2** `requires(provider) ⊆
  sends(consumer)` else `MISSING_REQUIRED_REQUEST_FIELD`; **R3** `reads(consumer) ⊆ provides(provider)`
  else `READS_FIELD_NOT_PROVIDED`; **R4** shared-field types match (`typesMatch`) else `TYPE_MISMATCH`.
  Deps = —. A violation is a **value**, not an error — no failure path.
  - antecedent: a paired `{ consumedOp, providerOp }`.
  - consequent: `[]Violation` (possibly empty ⇒ compatible for this op).
- **unit tests: 6** (`module-tree.md` formula) — 1 happy + 5 branches: R2 required **body** field not
  sent, R2 required **param** (path/query/header) not sent, R3 read field not provided, R4 type
  mismatch (request), R4 type mismatch (response). (`body` vs `param` equivalence is unit-level, lesson D1.)
- component scenario(s) to green: none — the four verdict codes (exit 1) are UNIT boundaries here,
  **not** component scenarios (ADR-0002).

**Dependencies:** `ConsumedOperation`/`ProviderOperation`/`Violation` types + verdict-code constants (ticket 03).

**Subagent instruction:** write the 6 unit tests → implement `CompareOperation` + `typesMatch` in
`internal/validate/compare/rules.go` → run units → green → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/... && go test ./internal/validate/compare/...`.

**Acceptance:** package `compare` builds/vets clean; 6 unit tests green.

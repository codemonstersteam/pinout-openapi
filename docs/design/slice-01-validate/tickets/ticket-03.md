---
id: 03
type: module
slice: slice-01-validate
blocked_by: [01, 02]
inputs: [docs/design/slice-01-validate/module-tree.md, docs/design/slice-01-validate/contracts.md, api-specification/config.schema.json, api-specification/report.schema.json]
outputs: [internal/validate/domain.go, internal/validate/errors.go]
io: none
skills: []
---

### TICKET 03 — slice-01-validate/foundation: shared slice types + sentinel errors (package `validate`)

**io:** none → skills: [] (pure type/sentinel files — carry NO behavioral contract; needed FIRST by
every module, so they are their own foundation ticket, not folded onto a leaf).

**Context (only these two files):**
- `internal/validate/domain.go` — the slice's shared value types (`module-tree.md` layout row):
  `Invocation`, `RawConfig`, `Config`, `Settings`, `ProviderConfig`, `ConsumedContract`,
  `ConsumedOperation` (`Ref`, `Sends`, `Reads`), `Provenance`, `ProviderSpec`, `ProviderOperation`
  (`Requires`, `Provides`), `OperationRef`, `Comparison`, `Violation`, `ComparisonOutcome`, `Report`.
  Field domains/shapes come from `config.schema.json` + `report.schema.json` (frozen) — do NOT invent.
- `internal/validate/errors.go` — sentinel errors (rise untransformed): `ErrConfig`, `ErrFileNotFound`,
  `ErrParse`, `ErrHTTP`, `ErrTimeout`; the four verdict `Violation` codes (`OP_NOT_IN_PROVIDER`,
  `MISSING_REQUIRED_REQUEST_FIELD`, `READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH`) as constants.
- unit tests: **none** (pure type/sentinel declarations — no antecedent→consequent to test).
- component scenario(s) to green: none directly (foundation for all).

**Note (import hygiene):** these are leaf types — `domain.go`/`errors.go` MUST NOT import any
`internal/validate/<subpkg>`; the sub-packages import them. Keep package `validate` cycle-free.

**Dependencies:** none beyond scaffold + component (RED-first edge).

**Subagent instruction:** declare the types + sentinels → `go build` the package → done. Touch no
other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `validate` compiles with all shared types + sentinels present; no units.

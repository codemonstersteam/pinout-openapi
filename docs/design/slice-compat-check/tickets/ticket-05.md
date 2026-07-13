---
id: 05
type: module
slice: slice-compat-check
blocked_by: [01, 02, 03, 04]
inputs: [docs/design/slice-compat-check/contracts.md, docs/design/slice-compat-check/module-tree.md, api-specification/report.schema.json]
outputs: [internal/compat-check/compare/operations.go, internal/compat-check/compare/operations_test.go, internal/compat-check/compare/presence.go, internal/compat-check/compare/presence_test.go, internal/compat-check/compare/request.go, internal/compat-check/compare/request_test.go, internal/compat-check/compare/response.go, internal/compat-check/compare/response_test.go, internal/compat-check/compare/status.go, internal/compat-check/compare/status_test.go, internal/compat-check/compare/content.go, internal/compat-check/compare/content_test.go, internal/compat-check/compare/operation.go, internal/compat-check/compare/operation_test.go, internal/compat-check/compare/verdict.go, internal/compat-check/compare/verdict_test.go, internal/compat-check/compare/domain.go]
io: none
skills: []
---

### TICKET S-compat-check.05 — `compare/`: the 5 compatibility rules (PURE CORE)

**io:** `none`   →  skills: (none beyond core `program-implementation`)

**Context (only this package — `internal/compat-check/compare/`; the pure core, the reason the tool
exists — never imports `cobra`/`net/http`/`os`/the `kin-openapi` loader, only parsed `spec.Spec`):**

- **`operations.go` — `NewCheckedOperations(ops []config.Operation, consumer spec.Spec) -> Result<CheckedOps, Error>`**
  (consumer pre-check). Confirms every configured `path`+`method` exists in the **consumer** spec.
  Failure → `ErrOpNotInConsumer` → `OP_NOT_IN_CONSUMER` (exit 2) — **misconfiguration, not an
  incompatibility** (UC Ext. 3a).
- **`presence.go`/`request.go`/`response.go`/`status.go`/`content.go` — `checkXxx(consumerOp,
  providerOp) -> []Finding`**, direction = consumer expectations must be satisfiable by the provider
  master (`allOf` already flattened by `spec/`; `oneOf`/`anyOf`/enum-narrowing/`format` deferred —
  non-goals, do not implement):
  - `checkPresence` → provider exposes `path`+`method`? absent → `Finding{rule: presence}`.
  - `checkRequestContravariance` → `provider.required(request) ⊆ consumer.sent(request)`? extra
    provider-required item → `Finding{rule: request_contravariance}`.
  - `checkResponseCovariance` → `consumer.read(response) ⊆ provider.provided(response)`? missing field →
    `Finding{rule: response_covariance}`.
  - `checkStatusCodes` → consumer codes ⊆ provider codes? missing code → `Finding{rule: status_codes}`.
  - `checkContentTypes` → content-types matched? mismatch → `Finding{rule: content_types}`.
  - **`rule` is the UNDERSCORE form** (`request_contravariance`, not hyphenated) — per frozen
    `report.schema.json`. Do not emit the hyphen prose form.
- **`operation.go` — `NewComparison(checkedOps, consumer, provider) -> Comparison`** (unites the 3
  entities, one-data-argument discipline) and **`compareOperation(comparisonOp) -> OperationResult`**
  (runs the 5 rules, concatenates findings; `status=incompatible` iff `findings≥1`, else `compatible`).
- **`verdict.go` — `aggregateVerdict(results) -> Verdict`** (any incompatible op ⇒ overall
  `incompatible`) and **`buildReport(config, results, providerSource) -> Report`** (assemble
  `report.schema.json` canon shape: `verdict`, `consumer.name`, `provider.{name,source}`,
  `operations[]{path,method,status,findings[]}`; **invariant by construction:**
  `status=compatible ⇒ findings=[]` / `status=incompatible ⇒ findings≥1`).
- **`domain.go`** — `Report`, `OperationResult`, `Finding`, `Rule`, `Verdict` canon types (not tested).

**`Compare(comparison) -> Report`** (the pure-core entry the head calls directly) is the composition of
`compareOperation` per op → `aggregateVerdict` → `buildReport`; it returns a `Report` for **both**
verdicts (never an error) — `INCOMPATIBLE` is a verdict carried forward, not a pipe short-circuit.

**Unit tests (18 by formula):** `NewCheckedOperations` (2: happy + op absent from consumer) ·
`checkPresence` (2: happy + absent in provider) · `checkRequestContravariance` (2: happy + extra
provider-required item) · `checkResponseCovariance` (2: happy + consumer reads a field provider omits) ·
`checkStatusCodes` (2: happy + consumer code ∉ provider codes) · `checkContentTypes` (2: happy + mismatch)
· `compareOperation` (2: all rules pass → compatible; ≥1 finding → incompatible) · `NewComparison`
(1: happy, unites 3 entities) · `aggregateVerdict` (2: all compatible; ≥1 incompatible op → incompatible)
· `buildReport` (1: happy, invariant holds by construction). `domain.go` and `Compare` itself are not
separately unit-tested (types / composition of already-tested pieces).

**Dependencies:** imports `internal/compat-check/config` (ticket 03, `Operation` type) and
`internal/compat-check/spec` (ticket 04, `Spec` type). Does not import `report/`, `cli/`, or the
package-root files.

**Subagent instruction:** write the 8 `_test.go` files (18 cases total, distributed per the table above) →
implement `domain.go` → the 5 `checkXxx` files → `operations.go` → `operation.go` → `verdict.go` → run
tests → green? mark this ticket done. Do not touch `config/`, `spec/`, `report/`, `cli/`, or the
package-root files.

**Verify:** `go build ./internal/compat-check/compare/... && go vet ./internal/compat-check/compare/... && go test ./internal/compat-check/compare/...`

**Acceptance:** package builds and vets clean; all 18 unit tests green, including the `OP_NOT_IN_CONSUMER`
boundary (`NewCheckedOperations`) and the `INCOMPATIBLE` rule-boundaries (`compareOperation`/
`aggregateVerdict`) — these are **unit** boundaries per `contracts.md`'s anti-gaming reconciliation, never
component scenarios.

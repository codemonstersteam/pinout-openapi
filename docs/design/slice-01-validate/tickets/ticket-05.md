---
id: 05
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, api-specification/report.schema.json]
outputs: [internal/validate/cli/exit.go]
io: none
skills: []
---

### TICKET 05 — slice-01-validate/cli.ResolveExitCode: Result → exit code (adapter)

**io:** none → skills: [] (mechanical Result→exit mapping, `cli-io` grid). **Not unit-tested** —
asserted by component scenarios.

**Context (only this module):**
- contract (`contracts.md` §cli.ResolveExitCode): `ResolveExitCode(res: Result[Report, Error]) -> int`.
  Grid: `Ok(compatible)`→0; `Ok(incompatible)`→1; `ErrConfig`→2;
  `ErrFileNotFound|ErrParse|ErrHTTP|ErrTimeout`→3 (per `report.schema.json` `x-exit-codes`).
  Deps = —. antecedent: the head's `Result`. consequent: one exit code ∈ {0,1,2,3}.
- unit tests: **none** (mechanical adapter — exit codes asserted by component scenarios).
- component scenario(s) to green: contributes exit codes to scenarios 1–6 — greened later.

**Note:** this file holds only the Result→int mapping; **printing** the report to stdout / logs to
stderr + `os.Exit` is done by `main` (wiring, ticket 17), not here.

**Dependencies:** sentinels + `Report` type (ticket 03).

**Subagent instruction:** implement `cli.ResolveExitCode` in `internal/validate/cli/exit.go` →
`go build`/`vet` → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `cli` builds and vets clean; no units. Component GREEN is @fagan's step.

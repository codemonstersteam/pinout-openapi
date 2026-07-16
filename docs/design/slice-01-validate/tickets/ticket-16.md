---
id: 16
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03, 06, 07, 08, 09, 12, 13, 14, 15]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md]
outputs: [internal/validate/head.go]
io: none
skills: []
---

### TICKET 16 — slice-01-validate/ProcessValidate (head): the ROP composition pipe

**io:** none → skills: [] (composition root — a linear ROP pipe of already-tested parts; no branching
of its own, **not unit-tested**).

**Context (only this module):**
- contract (`contracts.md` §ProcessValidate + `module-tree.md` head-pipe pseudocode):
  `ProcessValidate(inv: Invocation, d: Deps) -> Result[Report, Error]`. Deps =
  `Deps{ ConfigStore, ContractStore, SpecLoader, ReportWriter, Clock, *slog.Logger }` — autonomous I/O
  objects + orthogonal tools (**no** raw `*os.File`/`*http.Client`). **Define the `Deps` ports struct in
  `head.go`** (the composition root owns the ports it needs) so the package builds standalone; the
  wiring ticket (17) supplies the concrete construction in `register.go`.
  - Linear pipe (short-circuit on the first sentinel, risen untransformed):
    `ConfigStore.Load` → `NewConfig` → `ContractStore.Load` → `SpecLoader.Load` → `NewComparison` →
    `CompareContracts` → `FoldReport` → `ReportWriter.Write` → `Ok(report)`.
  - antecedent: a valid `Invocation`. consequent: Ok `Report` (`compatible ⇔ errors == []`); Fail any
    child's sentinel, short-circuited. The head **never branches on an error**; verdict codes (exit 1)
    are folded on the **success** path by `CompareContracts`, not short-circuits.
- unit tests: **none** (pipe of already-tested parts — exit codes asserted by component scenarios).
- component scenario(s) to green: 1–6 end-to-end (with wiring, ticket 17) — greened later by @fagan.

**Dependencies:** `ConfigStore.Load`(06), `NewConfig`(07), `ContractStore.Load`(08), `SpecLoader.Load`(09),
`NewComparison`(12), `CompareContracts`(13), `FoldReport`(14), `ReportWriter.Write`(15); shared types (03).

**Subagent instruction:** implement `ProcessValidate` + the `Deps` ports struct in
`internal/validate/head.go` → `go build`/`vet` → done. Touch no other module; do NOT run the component
harness or drive scenarios GREEN (that grows past one module).

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `validate` builds and vets clean; no units (composition pipe). Component GREEN
is the @fagan acceptance step.

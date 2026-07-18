---
id: 16
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03, 06, 07, 08, 09, 12, 13, 14, 15]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md, docs/design/slice-01-validate/adr/0005-late-specloader-construction-via-factory.md]
outputs: [internal/validate/head.go]
io: none
skills: []
---

### TICKET 16 — slice-01-validate/ProcessValidate (head): the ROP composition pipe

> **REWORK (ADR-0005) — edit the EXISTING `internal/validate/head.go`.** Two changes: (1) the `Deps`
> ports struct now carries a **`BuildSpecLoader func(Settings) SpecLoader` factory** instead of a ready
> `SpecLoader`; (2) the pipe gains **one late-construction step** — `d.BuildSpecLoader(cfg.Settings)` —
> between `NewConfig`/`ContractStore.Load` and `loader.Load`. Nothing else in the head moves.

**io:** none → skills: [] (composition root — a linear ROP pipe of already-tested parts; no branching
of its own, **not unit-tested**).

**Context (only this module):**
- contract (`contracts.md` §ProcessValidate + `module-tree.md` head-pipe pseudocode):
  `ProcessValidate(inv: Invocation, d: Deps) -> Result[Report, Error]`. **Define the `Deps` ports struct
  in `head.go`** (the composition root owns the ports it needs). Under ADR-0005 `Deps` is now:
  `Deps{ ConfigStore, ContractStore, BuildSpecLoader func(Settings) SpecLoader, ReportWriter, Clock, *slog.Logger }`
  — autonomous I/O objects + orthogonal tools (**no** raw `*os.File`/`*http.Client`). Note the shape
  change: **`BuildSpecLoader func(Settings) SpecLoader`** (a wired-once factory) **replaces** the former
  ready `SpecLoader` field. The `SpecLoader` interface stays declared in `head.go`; the wiring ticket
  (17) supplies the concrete factory closure in `register.go`.
  - Linear pipe (short-circuit on the first sentinel, risen untransformed) — the **one new step** in
    **bold**:
    `ConfigStore.Load` → `NewConfig` → `ContractStore.Load` → **`loader := d.BuildSpecLoader(cfg.Settings)`**
    → `loader.Load(cfg.Provider)` → `NewComparison` → `CompareContracts` → `FoldReport` →
    `ReportWriter.Write` → `Ok(report)`.
  - `d.BuildSpecLoader(cfg.Settings)` is a **pure, total factory call — NEVER an error step** (no
    `Result`, no short-circuit): it returns a `SpecLoader` whose HTTP branch is bounded by the **real**
    `cfg.Settings.Timeout`, computed after `NewConfig` validated it (> 0). This is why it must run
    *inside* the pipe, after `NewConfig` — the timeout is unknown at wiring time (ADR-0005). It removes
    the former eager, hardcoded-`30` construction that left `settings.timeout` inert (scenario 6 could
    never fire).
  - antecedent: a valid `Invocation`. consequent: Ok `Report` (`compatible ⇔ errors == []`); Fail any
    child's sentinel, short-circuited. The head **never branches on an error**; verdict codes (exit 1)
    are folded on the **success** path by `CompareContracts`, not short-circuits.
- unit tests: **none** (pipe of already-tested parts — exit codes asserted by component scenarios; the
  factory step is total, so it adds no branch to test).
- component scenario(s) to green: 1–6 end-to-end (with wiring, ticket 17) — greened later by @fagan.
  Scenario 6 (`TIMEOUT_ERROR`) becomes reachable precisely because of the late `BuildSpecLoader` step.

**Dependencies:** `ConfigStore.Load`(06), `NewConfig`(07), `ContractStore.Load`(08), `SpecLoader.Load`(09),
`NewComparison`(12), `CompareContracts`(13), `FoldReport`(14), `ReportWriter.Write`(15); shared types (03,
incl. `Settings`).

**Subagent instruction:** edit `internal/validate/head.go` — change the `Deps` field from `SpecLoader`
to `BuildSpecLoader func(Settings) SpecLoader`, insert the `loader := d.BuildSpecLoader(cfg.Settings)`
step after `NewConfig` and before `loader.Load(cfg.Provider)` → `go build`/`vet` → done. Touch no other
module; do NOT run the component harness or drive scenarios GREEN.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `validate` builds and vets clean; `Deps` carries `BuildSpecLoader func(Settings)
SpecLoader` (no ready `SpecLoader` field); the pipe constructs the loader late via the factory before
`loader.Load`; no units (composition pipe). Component GREEN is the @fagan acceptance step.

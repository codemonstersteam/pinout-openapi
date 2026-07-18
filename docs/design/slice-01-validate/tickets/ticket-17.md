---
id: 17
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03, 04, 05, 06, 07, 08, 09, 10, 11, 12, 13, 14, 15, 16]
inputs: [docs/design/slice-01-validate/module-tree.md, docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/adr/0005-late-specloader-construction-via-factory.md]
outputs: [internal/validate/register.go, cmd/app/main.go]
io: none
skills: []
---

### TICKET 17 — slice-01-validate/wiring: concrete Deps + mount the command (exposes the CLI)

> **REWORK (ADR-0005) — edit the EXISTING `internal/validate/register.go`.** The wiring no longer
> constructs a `SpecLoader` eagerly. It stops passing a ready `SpecLoader` and instead supplies the
> **factory closure** `Deps.BuildSpecLoader`; the `defaultProviderTimeoutSeconds = 30` hardcode is
> **deleted**. `cmd/app/main.go` is unaffected by this delta (verify it still builds).

**io:** none → skills: [] (wiring only — implements NO module logic; single concern = expose the
command by constructing the concrete `Deps`).

**Context (only wiring — two files):**
- `internal/validate/register.go` — construct the concrete `Deps` for `ProcessValidate`: real
  `ConfigStore`, `ContractStore`, `ReportWriter` (from the `internal/validate/**` module packages) +
  `Clock` + `*slog.Logger` (log level from `settings.log_level`). **ADR-0005 change:** instead of
  building a `SpecLoader` at wiring time, supply the factory field —
  `BuildSpecLoader: func(s Settings) SpecLoader { return provider.NewSpecLoader(s.Timeout) }` — so the
  head constructs the loader late with the real `cfg.Settings.Timeout`. **Delete the
  `defaultProviderTimeoutSeconds = 30` constant and its eager `provider.NewSpecLoader(30)` call** — they
  no longer exist; the timeout is always the config value. No domain logic.
- `cmd/app/main.go` — the live wiring per `module-tree.md`: `cli.Parse(os.Args)` → build `Deps`
  (register.go) → `ProcessValidate(inv, deps)` → `code := cli.ResolveExitCode(res)`; **always print the
  report JSON to stdout** (machine channel), diagnostics/logs to **stderr**; `os.Exit(code)`. Keep the
  binary at `cmd/app/` (no `cmd/<slug>/` rename). Command surface: `validate <config.yaml>` +
  `version`/`--help`. **This file needs no change for ADR-0005** (it never saw the loader) — only verify
  it still builds against the updated `Deps`/`register.go`.

- unit tests: **none** (wiring).
- component scenario(s) to green: this ticket keeps scenarios 1–6 reachable end-to-end; scenario 6
  (`TIMEOUT_ERROR`) now truly fires because the loader is built with the real `settings.timeout`.
  **Greening + `@wip`-removal is the @fagan acceptance step, NOT this ticket's deliverable.**

**Dependencies:** every module (03–16) — `cli.Parse`(04), `cli.ResolveExitCode`(05), the I/O objects
(06, 08, 15), `provider.NewSpecLoader`(09, now called via the factory closure), `ProcessValidate` + its
`Deps`/`BuildSpecLoader` field (16).

**Subagent instruction:** edit `register.go` — replace the eager `SpecLoader` construction with the
`BuildSpecLoader` factory closure, delete `defaultProviderTimeoutSeconds` → verify `cmd/app/main.go`
still wires against the updated `Deps` → `go build ./...` → done. Do NOT implement module logic; do NOT
run the component harness to GREEN (that is @fagan).

**Verify:** `go build ./... && go vet ./...` (whole module builds; no `defaultProviderTimeoutSeconds`
remains — `grep -R defaultProviderTimeoutSeconds internal cmd` returns nothing).

**Acceptance:** `go build ./...` green; `register.go` supplies `Deps.BuildSpecLoader = func(s Settings)
SpecLoader { return provider.NewSpecLoader(s.Timeout) }` and no longer references
`defaultProviderTimeoutSeconds`; `pinout-openapi validate <config.yaml>` runs the real pipe, report
printed to stdout, exit code from `cli.ResolveExitCode`. Slice closure (remove `@wip` + component GREEN
+ verify every `TASK §DoD`) is the **@fagan acceptance step**.

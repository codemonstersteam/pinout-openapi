---
id: 17
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03, 04, 05, 06, 07, 08, 09, 10, 11, 12, 13, 14, 15, 16]
inputs: [docs/design/slice-01-validate/module-tree.md, docs/design/slice-01-validate/contracts.md]
outputs: [internal/validate/register.go, cmd/app/main.go]
io: none
skills: []
---

### TICKET 17 — slice-01-validate/wiring: concrete Deps + mount the command (exposes the CLI)

**io:** none → skills: [] (wiring only — implements NO module logic; single concern = expose the
endpoint by replacing the placeholder).

**Context (only wiring — two files):**
- `internal/validate/register.go` — construct the concrete `Deps` for `ProcessValidate`: real
  `ConfigStore`, `ContractStore`, `SpecLoader`, `ReportWriter` (from the `internal/validate/**` module
  packages) + `Clock` + `*slog.Logger` (log level from `settings.log_level`). No domain logic.
- `cmd/app/main.go` — replace the template placeholder (`run` → `NOT_IMPLEMENTED`/exit 3) with the live
  wiring per `module-tree.md`: `cli.Parse(os.Args)` → build `Deps` (register.go) →
  `ProcessValidate(inv, deps)` → `code := cli.ResolveExitCode(res)`; **always print the report JSON to
  stdout** (machine channel), diagnostics/logs to **stderr**; `os.Exit(code)`. Keep the binary at
  `cmd/app/` (no `cmd/<slug>/` rename). Command surface: `validate <config.yaml>` + `version`/`--help`.

- unit tests: **none** (wiring).
- component scenario(s) to green: this ticket makes scenarios 1–6 reachable end-to-end; **greening +
  `@wip`-removal is the @fagan acceptance step, NOT this ticket's deliverable**.

**Dependencies:** every module (03–16) — `cli.Parse`(04), `cli.ResolveExitCode`(05), the four I/O
objects (06, 08, 09, 15), `ProcessValidate`(16).

**Subagent instruction:** wire `register.go` + `cmd/app/main.go` → `go build ./...` → done. Do NOT
implement module logic; do NOT run the component harness to GREEN (that is @fagan).

**Verify:** `go build ./... && go vet ./...` (whole module builds; placeholder replaced, endpoint live).

**Acceptance:** `go build ./...` green; `pinout-openapi validate <config.yaml>` runs the real pipe
(no longer `NOT_IMPLEMENTED`), report printed to stdout, exit code from `cli.ResolveExitCode`. Slice
closure (remove `@wip` + component GREEN + verify every `TASK §DoD`) is the **@fagan acceptance step**.

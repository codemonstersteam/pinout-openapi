---
id: 09
type: module
slice: slice-compat-check
blocked_by: [01, 02, 03, 04, 05, 06, 07, 08]
inputs: [docs/design/slice-compat-check/module-tree.md, docs/design/slice-compat-check/contracts.md]
outputs: [internal/compat-check/register.go, cmd/app/main.go]
io: none
skills: []
---

### TICKET S-compat-check.09 — wiring: assemble concrete `Deps` + mount the real CLI in `cmd/app/main.go`

**io:** `none`   →  skills: (none beyond core `program-implementation`)

**Context (exactly two files — do not touch anything else, single concern: exposing the CLI):**
- **`internal/compat-check/register.go` — `NewDeps() Deps`.** The **concrete** assembly only (the `Deps`
  struct type and the four port interfaces already exist — ticket 06's `domain.go` — this file does NOT
  redeclare them). Plug the concrete adapters built by tickets 03/04/07 into the four fields:
  `Deps{Config: config.NewReader(), Consumer: spec.NewConsumerReader(), Provider: spec.NewProviderReader(),
  Report: report.NewWriter()}` (constructor names per each package's actual exported factory — use
  whatever ticket 03/04/07 actually named them, do not invent new ones).
- **`cmd/app/main.go`** — replace the template's `internal/example` placeholder wiring
  (`example.Head(deps, in)` → `NOT_IMPLEMENTED`) with the real pipe: `cli.Parse(args) -> Request` →
  `compatcheck.ProcessCompatCheck(req, register.NewDeps())` → `cli.Exit(result)`. Delete the now-dead
  `internal/example/` placeholder package (no longer imported by `main.go`, no longer needed — it was
  scaffold-only boilerplate, never this slice's code). Keep `cmd/app` as the binary directory name
  (scaffold.sh does not rename `cmd/`; there is no `cmd/<slug>/` — do not invent one).

**Unit tests:** none — this ticket writes no logic, only composition (proven by ALL 9 component scenarios
going GREEN once this lands, via `@fagan`'s acceptance step, not this ticket's own deliverable).

**Dependencies:** imports every module ticket's package: `config`/`spec`/`compare`/`report` (03/04/05/07,
concrete adapters), the package root `head`/`domain`/`errors` (06), and `cli` (08). This is the terminal
integration point of the module DAG.

**Subagent instruction:** implement `register.go` (`NewDeps`) → edit `cmd/app/main.go` to call
`cli.Parse` → `ProcessCompatCheck` → `cli.Exit` instead of the `internal/example` placeholder → delete
`internal/example/` → run build. Do **not** write README.md (separate ticket 10, ∥) and do not touch any
`internal/compat-check/<pkg>/*.go` file besides `register.go` — those are already done.

**Verify:** `go build ./... && go vet ./...` (whole repo, `internal/example` must be gone and unreferenced)
then a manual smoke: `go run ./cmd/app run <a CS1 fixture from component-tests/fixtures/>` → exit `0`,
JSON report on stdout. Full component-suite GREEN is **not** required by this ticket — that check (and
removing `@wip`) is `@fagan`'s acceptance step, run once, after this ticket lands.

**Acceptance:** `go build ./...` green with `internal/example` removed; `pinout-openapi run
<config>` (via `go run ./cmd/app run <config>`) drives the real `internal/compat-check` pipe end-to-end
(no more `NOT_IMPLEMENTED` anywhere) — this is what "exposes the CLI" means for this ticket. Full
component-scenario GREEN + `@wip` removal is explicitly **NOT** this ticket's deliverable.

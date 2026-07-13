---
id: 08
type: module
slice: slice-compat-check
blocked_by: [01, 02, 06]
inputs: [docs/design/slice-compat-check/contracts.md, docs/design/slice-compat-check/module-tree.md, api-specification/exit-codes.md]
outputs: [internal/compat-check/cli/command.go, internal/compat-check/cli/exit.go]
io: none
skills: [cli-io]
---

### TICKET S-compat-check.08 — `cli/`: the driving door (args → `Request`, `Result` → exit code)

**io:** `none` (inbound door — parses/serializes, not an outbound I/O object)   →  skills: `cli-io`

**Context (only this package — `internal/compat-check/cli/`; the ONE driving adapter of the hexagon):**
- **`command.go` — `Parse(args []string) -> Result<Request, Error>`.** Deps: — (cobra only; no domain, no
  I/O object). Flags/positional `<config>` → one flat `Request{ConfigPath}` (from
  `internal/compat-check/domain.go`, ticket 06). No default-path invention beyond what `contracts.md`/
  `use-case.md` specify (config path is a required positional arg; do not add flags not in the frozen
  `config.schema.json`/use-case).
- **`exit.go` — `Exit(Result<Outcome, Error>) -> (exitCode int, stdout, stderr)`.** Maps the head's
  result to the process boundary — **this is the ONLY place `error.code`/verdict → exit code happens**
  (per `contracts.md` §"Error model", mapped ONLY at the `cli/` door):

  | Outcome | exit |
  |---|---|
  | `Outcome.Verdict = compatible` | `0` |
  | `Outcome.Verdict = incompatible` | `1` |
  | `error.code ∈ {CONFIG_INVALID, SPEC_UNREADABLE, SPEC_PARSE_ERROR, OP_NOT_IN_CONSUMER}` | `2` |
  | `error.code ∈ {PROVIDER_UNREACHABLE, PROVIDER_TIMEOUT, PROVIDER_PARSE_ERROR, REPORT_WRITE_ERROR}` | `3` |

  Report JSON on stdout (already written by `report/`, ticket 07 — `Exit` does not re-serialize the
  report, only emits the exit code and routes logs to stderr per `settings.log_level`). No domain logic
  here — this door only parses/serializes and maps codes.

**Unit tests:** none — adapter (proven by all 9 component scenarios' `(exit code, stdout)` black-box
assertions, via the fixer, not this ticket).

**Dependencies:** imports `internal/compat-check` package root (ticket 06) for `Request`/`Outcome`/the
error sentinels. Does not call `ProcessCompatCheck` itself and does not build `Deps` — that orchestration
(`cli.Parse` → `head.ProcessCompatCheck` → `cli.Exit`) is assembled in `cmd/app/main.go` by the wiring
ticket (09), not here. Does not import `config/`, `spec/`, `compare/`, `report/`.

**Subagent instruction:** implement `command.go` (`Parse`) → `exit.go` (`Exit`, the code table above) →
run build/vet. No test file to write. Do not touch any other package and do not edit `cmd/app/main.go`
(that is ticket 09).

**Verify:** `go build ./internal/compat-check/cli/... && go vet ./internal/compat-check/cli/...`

**Acceptance:** package builds and vets clean; `Exit`'s code table matches `api-specification/exit-codes.md`
exactly (0/1/2/3, all 9 `error.code`s mapped). No component-green claim here — that is `@fagan`'s job
after wiring (ticket 09).

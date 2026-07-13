---
id: 06
type: module
slice: slice-compat-check
blocked_by: [01, 02, 03, 04, 05]
inputs: [docs/design/slice-compat-check/module-tree.md, docs/design/slice-compat-check/contracts.md]
outputs: [internal/compat-check/head.go, internal/compat-check/domain.go, internal/compat-check/errors.go]
io: none
skills: []
---

### TICKET S-compat-check.06 — package root: the head pipe + DTOs + port interfaces + sentinels

**io:** `none`   →  skills: (none beyond core `program-implementation`)

**Context (only the package-root files `internal/compat-check/{head.go,domain.go,errors.go}` — NOT
`register.go`, which is a separate later ticket, see Dependencies below):**

- **`head.go` — `ProcessCompatCheck(req Request, deps Deps) -> Result<Outcome, Error>`.** The linear
  ROP pipe (no branching of its own — a failing step short-circuits, the `ErrXxx` rises untransformed):
  ```
  deps.Config.Read(req.ConfigPath)                        -> RawConfig    # CONFIG_INVALID (unreadable)
  config.NewConfig(RawConfig)                              -> Config      # CONFIG_INVALID (bad YAML/field/enum/XOR/range)
  deps.Consumer.Load(Config.Consumer.SpecPath)             -> ConsumerSpec # SPEC_UNREADABLE | SPEC_PARSE_ERROR
  compare.NewCheckedOperations(Config.Operations, ConsumerSpec) -> CheckedOps # OP_NOT_IN_CONSUMER
  deps.Provider.Load(Config.Provider.Source)               -> ProviderSpec # PROVIDER_UNREACHABLE | PROVIDER_TIMEOUT | PROVIDER_PARSE_ERROR
  compare.NewComparison(CheckedOps, ConsumerSpec, ProviderSpec) -> Comparison
  compare.Compare(Comparison)                              -> Report      # verdict compatible|incompatible — NOT an error, flows on
  deps.Report.Write(Report, Config.Settings)                -> Outcome    # REPORT_WRITE_ERROR; Outcome carries verdict
  ```
  This head imports `internal/compat-check/config` and `internal/compat-check/compare` **directly**
  (their pure functions are called inline, not through `Deps` — only the four I/O ports are swappable).
  `INCOMPATIBLE` is a **verdict carried to the door**, never a short-circuit here.
- **`domain.go`** — `Request{ConfigPath}`, `Outcome{Verdict}` DTOs; **and** the `Deps` struct + the four
  **port interfaces** the head depends on (`ConfigReader`, `ConsumerSpecReader`, `ProviderSpecReader`,
  `ReportWriter` — signatures exactly as in `contracts.md` §"Driven I/O objects"). Declaring the
  interfaces here (not in `register.go`) is what lets this ticket's `head.go` compile standalone, and lets
  `config/`/`spec/`/`report/` (tickets 03/04/07) satisfy them structurally without importing this package
  (Go structural typing — no import cycle). `register.go` (ticket 09, wiring) later adds ONLY the
  **concrete** `NewDeps()` assembly (wiring concrete adapters into the interfaces declared here) — it does
  not redeclare the interfaces.
- **`errors.go`** — sentinel errors + `error.code` strings, the full map from `contracts.md` §"Error
  model": `ErrConfigInvalid`→`CONFIG_INVALID`, `ErrSpecUnreadable`→`SPEC_UNREADABLE`,
  `ErrSpecParse`→`SPEC_PARSE_ERROR`, `ErrOpNotInConsumer`→`OP_NOT_IN_CONSUMER`,
  `ErrProviderUnreachable`→`PROVIDER_UNREACHABLE`, `ErrProviderTimeout`→`PROVIDER_TIMEOUT`,
  `ErrProviderParse`→`PROVIDER_PARSE_ERROR`, `ErrReportWrite`→`REPORT_WRITE_ERROR`. Exit-code mapping
  itself (`error.code`/verdict → `0|1|2|3`) is the **door's** job (`cli/`, ticket 08) — this file only
  owns the sentinels + code strings, not the exit-code arithmetic.

**Unit tests:** none — `head.go` is the composition pipe (pipe, not unit-tested, proven by CS1/CS2 via the
fixer); `domain.go`/`errors.go` are types/sentinels (no test).

**Dependencies:** imports `internal/compat-check/config` (ticket 03) and `internal/compat-check/compare`
(ticket 05) for their pure functions/types; the four port-interface method signatures reference
`internal/compat-check/spec.Spec` (ticket 04) as a return type. Does not import `report/` or `cli/`
(those depend on this ticket, not vice versa) and does not write `register.go`.

**Subagent instruction:** implement `errors.go` → `domain.go` (DTOs + `Deps` + 4 port interfaces) →
`head.go` (the pipe above) → run build/vet. No test file to write. Do not touch `config/`, `spec/`,
`compare/`, `report/`, `cli/`, or `register.go`.

**Verify:** `go build ./internal/compat-check/... && go vet ./internal/compat-check/...` (package-root
files only import already-built sub-packages at this point; `cli/`/`report/`/`register.go` do not exist
yet and are not required for this build to pass — this package's own files must compile standalone).

**Acceptance:** `head.go`/`domain.go`/`errors.go` build and vet clean against `config/`, `spec/`,
`compare/` (tickets 03–05, already landed); the 4 port interfaces exactly match `contracts.md`'s driven
I/O signatures (no drift). No component/unit-green claim — that is later tickets' and `@fagan`'s job.

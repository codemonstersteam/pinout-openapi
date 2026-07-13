---
id: 03
type: module
slice: slice-compat-check
blocked_by: [01, 02]
inputs: [docs/design/slice-compat-check/contracts.md, docs/design/slice-compat-check/module-tree.md, api-specification/config.schema.json, api-specification/exit-codes.md]
outputs: [internal/compat-check/config/reader.go, internal/compat-check/config/config.go, internal/compat-check/config/config_test.go, internal/compat-check/config/domain.go]
io: none
skills: []
---

### TICKET S-compat-check.03 — `config/`: read `contract-tests.yaml` + valid-by-construction VOs

**io:** `none` (header value — the package's one filesystem-reading function, `ConfigReader.Read`, has no
dedicated remote sub-skill per `module-tree.md`'s `io: file` note; only `io: http` gets an extra skill)   →
skills: (none beyond core `program-implementation`)

**Context (only this package — `internal/compat-check/config/`):**
- **`reader.go` — `ConfigReader.Read(path string) -> Result<RawConfig, Error>`.** Deps: —. Reads
  `contract-tests.yaml` bytes, no transformation (I/O pipe — **not** unit-tested, proven by CS3).
  Failure → `ErrConfigInvalid` (missing/unreadable) → `error.code=CONFIG_INVALID`, exit 2.
- **`config.go` — `NewConfig(raw RawConfig) -> Result<Config, Error>`** plus the 7 VO factories
  `NewConsumer`/`NewProvider`/`NewProviderSource`/`NewOperation`/`NewMethod`/`NewOperationPath`/`NewSettings`.
  Deps: —. Each factory range-checks its field; a naked `Config{...}` literal outside this package is
  impossible (unexported fields) — illegal config states are unrepresentable past this boundary.
  Field rules (verbatim from `contracts.md`):
  - `NewMethod(raw)` → verb ∈ `{get,put,post,delete,options,head,patch,trace}`, case-insensitive → lower.
  - `NewOperationPath(raw)` → non-empty, leading `/`.
  - `NewOperation(path, method)` → composes the two validated VOs.
  - `NewProviderSource(url, path)` → exactly one of `url`/`path` present; output `kind ∈ {http, file}`.
  - `NewSettings(logLevel, saveJSON, reportFile, timeout)` → `log_level ∈ {debug,info,warn,error}` (def
    `info`); `save_json_report` def `true`; `timeout` int `1..600` (def `30`); `json_report_file` required
    when `save_json_report=true` (def `compatibility_report.json`).
  - `NewConsumer(name, operations)` → `name` non-empty (trim≠""); `operations` ≥1.
  - `NewProvider(name, source)` → `name` non-empty.
  - `NewConfig(raw)` → unmarshal YAML, assemble the above; unparseable YAML or any child VO rejection →
    `ErrConfigInvalid` → `CONFIG_INVALID`.
- **`domain.go`** — `Config`, `Consumer`, `Provider`, `ProviderSource`, `Operation`, `Method`, `Settings`
  types, all **unexported fields** (types only, no test).

**Unit tests (21 by formula — `1 happy + Σ distinguishable antecedent branches`; write these, nothing else):**
`NewMethod` (2: happy + invalid verb) · `NewOperationPath` (2: happy + no leading `/`) · `NewOperation`
(1: happy, composes validated VOs) · `NewProviderSource` (4: `url` happy, `path` happy, both present,
neither present) · `NewSettings` (5: happy + bad `log_level` + `timeout<1` + `timeout>600` +
`save_json_report && json_report_file missing`) · `NewConsumer` (3: happy + empty `name` + empty
`operations`) · `NewProvider` (2: happy + empty `name`) · `NewConfig` (2: happy + unparseable YAML).
`reader.go` and `domain.go` are **not** unit-tested (I/O pipe / types).

**Dependencies:** none — this is the first (leaf) module ticket. Do not import any other
`internal/compat-check/*` sub-package.

**Subagent instruction:** write `config_test.go` (21 cases above) → implement `domain.go` → `config.go` →
`reader.go` → run tests → green? mark this ticket done. Do not touch `cli/`, `spec/`, `compare/`, `report/`,
or the package-root files (`head.go`/`domain.go`/`errors.go`/`register.go`) — those are other tickets.

**Verify:** `go build ./internal/compat-check/config/... && go vet ./internal/compat-check/config/... && go test ./internal/compat-check/config/...`

**Acceptance:** package builds and vets clean; all 21 unit tests green; `reader.go` builds clean (no unit
test — I/O pipe, proven later by component scenario CS3 via the fixer, **not** this ticket's deliverable).

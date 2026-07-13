# Module tree — slice `compat-check`

> Stage 5 (program-design / Wirth). Package root: `internal/compat-check/` (slug = `compat-check`).
> Target shape = **cli** → ingress door = `cli-io`. One external input (`pinout-openapi <contract-tests.yaml>`)
> → one `Request` → one `Result<Outcome, Error>`. Neighbor spec: [`use-case.md`](./use-case.md).
> Contracts + `io:` + component scenarios: [`contracts.md`](./contracts.md). C4: [`c4.md`](./c4.md).

## The secret each module hides (Parnas)

A module is a **secret**, not a layer. This slice hides five decisions behind five package boundaries:

| Sub-package | Secret it hides (the one decision that may change) | Kind |
|---|---|---|
| `cli/` | how the tool is spoken as a CLI (cobra flags/args → DTO, `Result` → exit code + streams) | driving adapter |
| `config/` | the `contract-tests.yaml` format and how it is read + validated | I/O (file) + logic |
| `spec/` | how an OpenAPI 3.x doc is loaded, `$ref`-resolved, `allOf`-flattened (**honest reuse of `kin-openapi`**) | I/O (file/http) |
| `compare/` | **the 5 compatibility rules** — the pure core, the reason the tool exists | pure logic |
| `report/` | how the canon report is persisted (file + stdout line) | I/O (file) |

The **pure core** (`compare/` + `config/` constructors) never imports `cobra`, `net/http`, `os`, or
`kin-openapi`'s loader — only its parsed model. The door and the four driven ports are the only nodes
that touch the external world. Swap `kin-openapi`, swap the CLI framework, swap file for HTTP provider
fetch — the core does not move.

## Ports (hexagon)

- **Driving (primary) adapter** — `cli/` (cobra `run <config>`).
- **Driven (secondary) ports** — four, each an autonomous I/O object the head sees only by its interface:
  - `ConfigReader` — read `contract-tests.yaml` bytes (file).
  - `ConsumerSpecReader` — load consumer OpenAPI doc (file) via `kin-openapi`.
  - `ProviderSpecReader` — load provider OpenAPI doc (**HTTP GET `spec_url` or file `spec_path`**) via `kin-openapi`.
  - `ReportWriter` — write JSON report to `json_report_file` (file) + machine-readable stdout line.

The head (`ProcessCompatCheck`) depends on the four **port interfaces** (`Deps`), never on
`*http.Client` / `*os.File` / `kin-openapi` loader directly.

## Head-pipe pseudocode (the composition root — pure linear pipe, no branching of its own)

```
ProcessCompatCheck(req: Request, deps: Deps) -> Result<Outcome, Error>:
    | deps.Config.Read(req.ConfigPath)                        -> RawConfig      # io:file  short-circuit → CONFIG_INVALID (unreadable)
    | NewConfig(RawConfig)                                    -> Config         # io:none  short-circuit → CONFIG_INVALID (bad YAML / field / enum / XOR / range)
    | deps.Consumer.Load(Config.Consumer.SpecPath)           -> ConsumerSpec   # io:file  short-circuit → SPEC_UNREADABLE | SPEC_PARSE_ERROR
    | NewCheckedOperations(Config.Operations, ConsumerSpec)  -> CheckedOps     # io:none  short-circuit → OP_NOT_IN_CONSUMER
    | deps.Provider.Load(Config.Provider.Source)             -> ProviderSpec   # io:http  short-circuit → PROVIDER_UNREACHABLE | PROVIDER_TIMEOUT | PROVIDER_PARSE_ERROR
    | NewComparison(CheckedOps, ConsumerSpec, ProviderSpec)  -> Comparison     # io:none  (unites the 3 entities into one — one-data-argument rule)
    | Compare(Comparison)                                    -> Report         # io:none  verdict compatible|incompatible — NOT an error, flows on
    | deps.Report.Write(Report, Config.Settings)            -> Outcome        # io:file  short-circuit → REPORT_WRITE_ERROR; Outcome carries verdict
```

**ROP short-circuit.** A failing step skips the rest; the `ErrXxx` rises **untransformed**; only the
`cli/` door maps `error → error.code → exit code`. The head has no `if`/`for` of its own.

**Key design decision — `incompatible` is a verdict, not a pipe error.** `Compare` returns a `Report`
(success) whether the verdict is `compatible` or `incompatible`; the report is **written for both
verdicts** (UC step 7 / Extension 5a). Only genuine failures short-circuit. Therefore `INCOMPATIBLE`
is a **core-logic outcome carried to the door** (door maps `verdict → exit 0|1`), **not** an
I/O-adapter branch (see `contracts.md` §Component scenarios).

## Node → file map (every path rooted in `internal/compat-check/`)

| File | Node | `io:` | Unit-tested? |
|---|---|---|---|
| `internal/compat-check/head.go` | `ProcessCompatCheck(req, deps)` head pipe | none | no (pipe) |
| `internal/compat-check/domain.go` | `Request`, `Outcome` DTOs | — | — (types) |
| `internal/compat-check/errors.go` | sentinel errors + `error.code` strings | — | — (sentinels) |
| `internal/compat-check/register.go` | `Deps` struct + port interfaces + wiring | — | — (wiring) |
| `internal/compat-check/cli/command.go` | cobra root + `run <config>`; `Parse(args) -> Request` | none | no (adapter) |
| `internal/compat-check/cli/exit.go` | `Outcome`/`Error → exit code (0/1/2/3)` + stdout(report)/stderr(logs) | none | no (adapter) |
| `internal/compat-check/config/reader.go` | `ConfigReader.Read(path) -> RawConfig` | **file** | no (I/O pipe) |
| `internal/compat-check/config/config.go` | `NewConfig`, `NewConsumer`, `NewProvider`, `NewProviderSource`, `NewOperation`, `NewMethod`, `NewOperationPath`, `NewSettings` | none | **yes** |
| `internal/compat-check/config/domain.go` | `Config`, `Consumer`, `Provider`, `ProviderSource`, `Operation`, `Method`, `Settings` (unexported fields) | — | — (types) |
| `internal/compat-check/spec/consumer_reader.go` | `ConsumerSpecReader.Load(path) -> Spec` | **file** | no (I/O pipe) |
| `internal/compat-check/spec/provider_reader.go` | `ProviderSpecReader.Load(source) -> Spec` | **http** | no (I/O pipe) |
| `internal/compat-check/spec/spec.go` | `Spec` (wraps `kin-openapi` model; `HasOperation`, `Operation(path,method)` queries) | — | — (types) |
| `internal/compat-check/compare/operations.go` | `NewCheckedOperations(ops, consumerSpec) -> CheckedOps` (consumer pre-check) | none | **yes** |
| `internal/compat-check/compare/presence.go` | `checkPresence` | none | **yes** |
| `internal/compat-check/compare/request.go` | `checkRequestContravariance` | none | **yes** |
| `internal/compat-check/compare/response.go` | `checkResponseCovariance` | none | **yes** |
| `internal/compat-check/compare/status.go` | `checkStatusCodes` | none | **yes** |
| `internal/compat-check/compare/content.go` | `checkContentTypes` | none | **yes** |
| `internal/compat-check/compare/operation.go` | `NewComparison`, `compareOperation` (aggregate 5 rules → `OperationResult`) | none | **yes** |
| `internal/compat-check/compare/verdict.go` | `aggregateVerdict`, `buildReport` | none | **yes** |
| `internal/compat-check/compare/domain.go` | `Report`, `OperationResult`, `Finding`, `Rule`, `Verdict` (canon report types) | — | — (types) |
| `internal/compat-check/report/writer.go` | `ReportWriter.Write(report, settings) -> Outcome` | **file** | no (I/O pipe) |

> `io: file` — filesystem I/O objects. The Step-5 enum (`none|http|llm|queue|db`) has no `file` value
> because filesystem I/O has **no dedicated remote sub-skill** (unlike `db`/`http`/`queue`); the `cli-io`
> skill already governs a CLI's local file handling. `file` marks these nodes as **I/O pipes** (not
> unit-tested, proven by component scenarios) rather than mislabelling them `none` (which would falsely
> claim pure logic). The ticket-writer (stage 6) attaches no external sub-skill for `io: file`; it attaches
> `http-io` for the one `io: http` node (`ProviderSpecReader`).

## Valid-by-construction (Go private-constructor analog)

`config/domain.go` types have **unexported fields**; the only way to obtain a `Config`/`Operation`/
`Method`/`ProviderSource`/`Settings` is its `NewX` factory, which range-checks **every** field
(no naked `Config{...}` literal outside `config/`). Illegal config states are unrepresentable past the
`config/` boundary — the compare core never re-checks them.

## Unit-test formula — `N = 1 (happy) + Σ (distinguishable antecedent branches)`

Only **constructors + pure logic** are unit-tested. Head, the four I/O objects, and the `cli/` adapter
are **not** unit-covered (pipes / parsers) — proven by component scenarios (`contracts.md`).

| Module | Happy | Distinguishable branches | Units |
|---|---|---|---|
| `NewMethod` | 1 | invalid verb | 2 |
| `NewOperationPath` | 1 | no leading `/` | 2 |
| `NewOperation` | 1 | (composes validated VOs) | 1 |
| `NewProviderSource` | 2 (`url` \| `path`, distinguishable output) | both present, neither present | 4 |
| `NewSettings` | 1 | bad `log_level`; `timeout<1`; `timeout>600`; `save_json_report && json_report_file missing` | 5 |
| `NewConsumer` | 1 | empty `name`; empty `operations` | 3 |
| `NewProvider` | 1 | empty `name` | 2 |
| `NewConfig` | 1 | unparseable YAML | 2 |
| `NewCheckedOperations` | 1 | an op absent from consumer spec (→ `OP_NOT_IN_CONSUMER`) | 2 |
| `checkPresence` | 1 | op absent in provider (→ finding) | 2 |
| `checkRequestContravariance` | 1 | provider requires an item consumer omits (→ finding) | 2 |
| `checkResponseCovariance` | 1 | consumer reads a field provider omits (→ finding) | 2 |
| `checkStatusCodes` | 1 | consumer code ∉ provider codes (→ finding) | 2 |
| `checkContentTypes` | 1 | content-type mismatch (→ finding) | 2 |
| `compareOperation` | 1 (all rules pass → compatible) | ≥1 finding → incompatible | 2 |
| `NewComparison` | 1 (unites 3 entities) | — | 1 |
| `aggregateVerdict` | 1 (all compatible) | ≥1 incompatible op → incompatible | 2 |
| `buildReport` | 1 (pure assembly; invariant `compatible ⇒ 0 findings` by construction) | — | 1 |
| **Total** | | | **39** |

> The `CONFIG_INVALID` field boundaries and `OP_NOT_IN_CONSUMER` and each rule's compatible/incompatible
> boundary are **units here** (input / logic boundaries), never component scenarios (`contracts.md` §Component
> scenarios explains the split and the anti-gaming count).

<!-- DONE: moduledesigner slice-compat-check -->
</content>
</invoke>

# Module tree — slice-01-validate

> Design package for slice **slice-01-validate** (`Owns package: internal/validate/`).
> In: frozen CLI contract (`api-specification/config.schema.json` + `report.schema.json`,
> `x-frozen: 2026-07-16`) + `use-case.md` (UC-1). Out: module tree + head-pipe pseudocode (this
> file), module contracts (`contracts.md`), C4 (`c4.md`). Target shape = **cli** (ingress door via
> `cli-io`; scaffold `template-go-cli`). Comparison algorithm is **ported, not reinvented**, from
> `sandbox/ALGORITHM.md` (`check.mjs`) onto Go + `kin-openapi`. Neighbor: [`use-case.md`](./use-case.md).

## What this slice is (one external input → one outcome)

One external input: the one-shot invocation `pinout-openapi validate <config.yaml>`. One outcome:
`Result[Report, Error]` serialized to **(exit code 0/1/2/3, stdout JSON report)**. The slice is a
black box: same input bytes ⇒ same verdict and same report bytes (deterministic pre-merge gate).

The **secret each module hides** (Parnas) — a module is a decision that can change without touching
its callers, not a file or a layer:

| Module / package | The one secret it hides |
|---|---|
| `cli` (ingress door) | how a CLI invocation is spoken (cobra, argv, flags) + how the outcome serializes (exit code, stdout/stderr) |
| `config` | the config-file format (YAML) and the config **validity rules** (schema invariants) |
| `contract` | the consumed-contract artifact format (typed `{sends, reads}` + `provenance` from the E-harness) |
| `provider` | how the provider spec is **acquired and parsed** — `kin-openapi`, `$ref` resolution, `spec_path` XOR `spec_url` HTTP fetch |
| `compare` | the four forward-compatibility **rules** (R1–R4) — the domain semantics |
| `report` | the report DTO shape (report.schema.json) and how it is persisted |

Slice = one bounded context; all sub-concerns are **Go sub-packages under `internal/validate/`** so
the vertical boundary is never a layer-cake. This mirrors the twin `../pinout-asyncapi` structurally:
its `parser/` (loads+parses the spec) ↔ our `provider/`; its `validator/` (comparison core) ↔ our
`compare/`.

## Module tree (C3 = this tree)

```
ProcessValidate (head — ROP pipe, no branching)          internal/validate/head.go        io: none
├── cli.Parse            (ingress: argv → Invocation)     internal/validate/cli/parse.go    io: none
├── ConfigStore.Load     (I/O pipe: read config file)     internal/validate/config/store.go io: none  [CONFIG_ERROR]
├── NewConfig            (constructor: validate config)    internal/validate/config/config.go io: none
├── ContractStore.Load   (I/O pipe: read+parse contract)  internal/validate/contract/store.go io: none [FILE_NOT_FOUND, PARSE_ERROR]
├── SpecLoader.Load      (I/O: kin-openapi acquire+parse) internal/validate/provider/loader.go io: http [FILE_NOT_FOUND, PARSE_ERROR, HTTP_ERROR, TIMEOUT_ERROR]
├── NewComparison        (constructor: unite the trio)     internal/validate/compare/comparison.go io: none
├── CompareContracts     (logic: fold R1–R4 over ops)      internal/validate/compare/compare.go io: none
│   ├── DeriveProviderOperation (logic: navigate spec)     internal/validate/provider/operation.go io: none [R1]
│   └── CompareOperation (logic: R2/R3/R4 for one op)      internal/validate/compare/rules.go   io: none
│         └── typesMatch (helper: R4 type equality)        internal/validate/compare/rules.go   io: none
├── FoldReport           (logic: outcome → Report DTO)     internal/validate/report/build.go   io: none
└── ReportWriter.Write   (I/O pipe: write JSON report)     internal/validate/report/writer.go  io: none

cli.ResolveExitCode      (adapter: Result → exit + stdout) internal/validate/cli/exit.go       io: none
```

`main` (in `cmd/pinout-openapi/main.go`) does three things only — **parse → `ProcessValidate(...)` →
`os.Exit(code)` + write report** — this is **wiring, not a contract-bearing module**.

## Head-pipe pseudocode (the composition root — pure, linear, no branching of its own)

```
ProcessValidate(inv Invocation, d Deps) -> Result[Report, Error]:
    | d.ConfigStore.Load(inv.ConfigPath)             -> RawConfig         -- fs read     [ErrConfig     → CONFIG_ERROR / 2]
    | NewConfig(rawConfig)                           -> Config            -- validate    [ErrConfig     → CONFIG_ERROR / 2]
    | d.ContractStore.Load(cfg.ConsumedContractPath) -> ConsumedContract  -- fs+parse    [ErrFileNotFound → FILE_NOT_FOUND / 3, ErrParse → PARSE_ERROR / 3]
    | d.SpecLoader.Load(cfg.Provider)               -> ProviderSpec      -- kin-openapi [ErrFileNotFound/3, ErrParse/3, ErrHTTP → HTTP_ERROR/3, ErrTimeout → TIMEOUT_ERROR/3]
    | NewComparison(cfg, consumed, spec)            -> Comparison        -- unite (scoped ops + sends/reads + provider ops + provenance)
    | CompareContracts(comparison)                  -> ComparisonOutcome -- R1..R4 per op, folded; verdict codes are exit 1
    | FoldReport(outcome)                           -> Report            -- compatible⇔errors==[]; provenance echo; uncovered_operations[]
    | d.ReportWriter.Write(cfg.Settings, report)    -> Report            -- write JSON iff settings.save_json_report (ROP pass-through)
    -> Ok(report)
```

Then in `main`: `code := cli.ResolveExitCode(res)`; the **report is always printed to stdout**
(machine channel), logs to stderr; `os.Exit(code)`.

### ROP short-circuit — errors rise untransformed, mapped once at the edge

A failing step **short-circuits the pipe**; the sentinel error rises untransformed; **only the `cli`
adapter** maps it to `(exit code, error.code, stderr)`. The head never branches on an error, has no
logic of its own — it is a straight pipe of already-designed parts.

The **verdict** codes (`OP_NOT_IN_PROVIDER`, `MISSING_REQUIRED_REQUEST_FIELD`,
`READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH`; exit **1**) are **not** pipe errors — they are the honest
domain answer produced by `CompareContracts` on the **success** path and folded into
`ComparisonOutcome.Violations`. `compatible ⇔ errors == []`; exit 1 is a legitimate outcome, not a
short-circuit. (See ADR-0002.)

## Key design decisions (promoted to ADRs — see [`adr/`](./adr/))

- **ADR-0001** — provider-spec acquisition **and** parse are unified behind **one** `SpecLoader`
  (`io: http`, `kin-openapi`), with `spec_path` XOR `spec_url` an internal strategy — not split into a
  filesystem object + an HTTP object.
- **ADR-0002** — the four verdict codes (exit 1) are a **domain verdict proven by UNIT tests** over
  `CompareOperation`, **not** component scenarios; only the happy path + the 5 io/config **adapter
  branches** are component scenarios (count = `1 + Σ adapter branches = 6`).
- **ADR-0003** — filesystem loaders (`ConfigStore`, `ContractStore`, `ReportWriter`) are tagged
  `io: none`: the `io:` enum (`none|http|llm|queue|db`) has no `file` value and local reads route to
  no io sub-skill; they remain **isolated I/O pipes** (their failure branches are still component
  scenarios), the tag only means "no metered/network sub-skill applies".

## File layout (slice-aligned — every path under `internal/validate/`)

| Node (module tree) | File |
|---|---|
| `ProcessValidate` (head) | `internal/validate/head.go` |
| `Deps` + wiring | `internal/validate/register.go` |
| slice types (`Invocation`, `Config`, `ConsumedContract`, `ProviderSpec`, `ProviderOperation`, `OperationRef`, `Comparison`, `Violation`, `ComparisonOutcome`, `Report`) | `internal/validate/domain.go` |
| slice sentinel errors (`ErrConfig`, `ErrFileNotFound`, `ErrParse`, `ErrHTTP`, `ErrTimeout`) + verdict codes | `internal/validate/errors.go` |
| `cli.Parse` (ingress door) | `internal/validate/cli/parse.go` |
| `cli.ResolveExitCode` + stdout/stderr write | `internal/validate/cli/exit.go` |
| `ConfigStore.Load` (I/O) | `internal/validate/config/store.go` |
| `NewConfig` (constructor) | `internal/validate/config/config.go` |
| `ContractStore.Load` (I/O) | `internal/validate/contract/store.go` |
| `SpecLoader.Load` (I/O — kin-openapi) | `internal/validate/provider/loader.go` |
| `DeriveProviderOperation` (logic, R1) | `internal/validate/provider/operation.go` |
| `NewComparison` (constructor) | `internal/validate/compare/comparison.go` |
| `CompareContracts` (logic, fold) | `internal/validate/compare/compare.go` |
| `CompareOperation` + `typesMatch` (logic, R2/R3/R4) | `internal/validate/compare/rules.go` |
| `FoldReport` (logic) | `internal/validate/report/build.go` |
| `ReportWriter.Write` (I/O) | `internal/validate/report/writer.go` |
| `main` (wiring only, not a module) | `cmd/pinout-openapi/main.go` |

Shared cross-slice types → `internal/shared/` (none for this single-slice service). No layer-keyed
roots (`internal/io`, `internal/logic`, …) — the vertical boundary is preserved.

## Unit-test formula (`N = 1 happy + Σ antecedent branches`) — logic modules only

The **head, all I/O modules, and the `cli` adapter (parse + `ResolveExitCode`) are NOT unit-tested**
(pipe / parse-map / mechanical Result→exit). Exit codes are asserted by component scenarios. Only
constructors and pure logic are unit-tested:

| Module | Happy | Distinguishable antecedent branches | Units |
|---|---|---|---|
| `NewConfig` | 1 | empty name (consumer/provider), empty `operations`, bad `path` pattern (`^/`), bad `method` enum, `spec_*` both-set, `spec_*` neither-set, `timeout ≤ 0`, bad `log_level` enum | 9 |
| `NewComparison` | 1 | (uniting constructor over already-valid inputs — no failing antecedent) | 1 |
| `DeriveProviderOperation` | 1 | operation absent in provider (→ R1 `OP_NOT_IN_PROVIDER`) | 2 |
| `CompareOperation` | 1 | R2 required **body** field not sent, R2 required **param** (path/query/header) not sent, R3 read field not provided, R4 type mismatch (request), R4 type mismatch (response) | 6 |
| `CompareContracts` | 1 | multi-operation fold accumulation, `uncovered_operations[]` detection | 3 |
| `FoldReport` | 1 | non-empty `errors` ⇒ `compatible=false`, `uncovered_operations[]` populated | 3 |
| **Total** | | | **24** |

> The four verdict codes are covered **here**, as `CompareOperation` unit boundaries (R2/R3/R4) +
> `DeriveProviderOperation` (R1) — never as component scenarios (ADR-0002). The `body` vs
> `path/query/header` equivalence for R2/R4 is unit-level (a `Request`-shaped boundary, lesson D1).

## Component-scenario count (`N = 1 happy + Σ distinguishable adapter branches`)

**6 scenarios** = 1 happy (compatible, exit 0) + 5 adapter branches (one per distinct io/config
`error.code`): `CONFIG_ERROR` (2), `FILE_NOT_FOUND` (3), `PARSE_ERROR` (3), `HTTP_ERROR` (3),
`TIMEOUT_ERROR` (3). The full set + Gherkin-mapping is designed in [`contracts.md`](./contracts.md)
(tagged `@wip`); realization into `.feature` + harness is `@wirth-tester`, not this stage.

<!-- DONE: moduledesigner slice-01-validate -->

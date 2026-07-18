# PLAN — slice-01-validate

> Plan-index (assembled by `wirth-planner`, izi stage: plan-index). **In:** the finished design
> package for `slice-01-validate` (module tree, contracts, C4, use case, ADRs, 17 cut tickets, frozen
> contract). **Out:** this file — a path index + the Gate #1 operator summary. It links; it does not
> duplicate, except the head-pipe block and the failure map below (the one allowed content copy).
>
> Single-slice service: `#slices = 1` (`.agent/planner/slices.md`) — this is the only `PLAN.md` this
> repo needs.

## Design package (links, no content copy)

| Artifact | Path |
|---|---|
| FRD (source requirements) | [`.agent/planner/frd.md`](../../../.agent/planner/frd.md) |
| Slice backlog (1 slice, rationale) | [`.agent/planner/slices.md`](../../../.agent/planner/slices.md) |
| Cockburn use case (UC-1) | [`use-case.md`](./use-case.md) |
| Module tree + head-pipe pseudocode | [`module-tree.md`](./module-tree.md) |
| Module contracts + error model + component-scenario design | [`contracts.md`](./contracts.md) |
| C4 (C2 container + C3 component = module tree) | [`c4.md`](./c4.md) |
| Domain glossary (ubiquitous language) | [`CONTEXT.md`](./CONTEXT.md) |
| ADR-0001 — `SpecLoader` unifies file + HTTP acquisition | [`adr/0001-specloader-unifies-file-and-http.md`](./adr/0001-specloader-unifies-file-and-http.md) |
| ADR-0002 — verdict codes are unit-level, not component-level | [`adr/0002-verdict-is-unit-not-component.md`](./adr/0002-verdict-is-unit-not-component.md) |
| ADR-0003 — filesystem loaders tagged `io: none` | [`adr/0003-filesystem-loaders-tagged-io-none.md`](./adr/0003-filesystem-loaders-tagged-io-none.md) |
| Frozen contract — config input | [`api-specification/config.schema.json`](../../../api-specification/config.schema.json) |
| Frozen contract — report output + `x-exit-codes` | [`api-specification/report.schema.json`](../../../api-specification/report.schema.json) |
| Root README (failure map, quick start) | [`README.md`](../../../README.md) |
| Tickets (17, scaffold → component RED → modules → wiring) | [`tickets/`](./tickets/) |

## Operator summary for Gate #1

### Head module, functional style (verbatim from `module-tree.md`)

```text
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

Then in `main`: `code := cli.ResolveExitCode(res)`; the report is always printed to stdout (machine
channel), logs to stderr; `os.Exit(code)`. A failing step short-circuits the pipe untransformed; only
the `cli` adapter maps the sentinel to `(exit code, error.code, stderr)` — the head never branches.

### Failure-mode map (`error.code` → exit + client/operator action)

Source: `api-specification/report.schema.json` `x-exit-codes` + `contracts.md` error model.

| `error.code` | exit | class | client/operator action |
|---|---|---|---|
| *(none — `compatible: true`, `errors: []`)* | 0 | compatible | none — contracts are forward-compatible, safe to proceed |
| `OP_NOT_IN_PROVIDER` | 1 | verdict (R1) | provider dropped/renamed the operation — reconcile provider spec or consumer's `consumed-contract` |
| `MISSING_REQUIRED_REQUEST_FIELD` | 1 | verdict (R2) | provider now requires a body/param field the consumer doesn't send — update consumer's request |
| `READS_FIELD_NOT_PROVIDED` | 1 | verdict (R3) | consumer reads a response field the provider no longer returns — update consumer or restore the field |
| `TYPE_MISMATCH` | 1 | verdict (R4) | request/response field type diverged — align consumer and provider types |
| `CONFIG_ERROR` | 2 | config/usage | fix invocation: config file missing/unreadable/schema-invalid, or `spec_path`/`spec_url` not exactly-one (never appears in `errors[]`) |
| `FILE_NOT_FOUND` | 3 | io·parse | `consumed-contract` or provider spec file missing/unreadable at its configured path — fix the path |
| `PARSE_ERROR` | 3 | io·parse | `consumed-contract` or provider spec is unparseable (invalid OpenAPI/YAML) — fix the artifact |
| `HTTP_ERROR` | 3 | io·parse | provider `spec_url` unreachable / non-2xx — check provider endpoint availability |
| `TIMEOUT_ERROR` | 3 | io·parse | provider `spec_url` fetch exceeded `settings.timeout` — raise the timeout or check network/provider latency |

Verdict codes (exit 1) are the **domain answer on the success path** (proven by unit tests over
`CompareOperation`/`DeriveProviderOperation`, ADR-0002) — not pipe errors. `CONFIG_ERROR` is detected
before a report is written, so it never appears inside `errors[]`.

### Ticket count/order (17 tickets, `tickets/`)

| Stage | Tickets | What |
|---|---|---|
| Scaffold | 01 | clone `template-go-cli`, runnable placeholder (`go build` + smoke green) |
| Component RED | 02 | realize the 6 designed scenarios as RED `.feature` (1 happy + 5 adapter-branch codes) |
| Foundation | 03 | shared slice types + sentinel errors (package `validate`) |
| Modules (parallel after 01–03) | 04–12, 14–15 | `cli.Parse`, `cli.ResolveExitCode`, `ConfigStore.Load`, `NewConfig`, `ContractStore.Load`, `SpecLoader.Load`, `DeriveProviderOperation`, `CompareOperation`+`typesMatch`, `NewComparison`, `FoldReport`, `ReportWriter.Write` |
| Modules (dependent) | 13 | `CompareContracts` (folds R1–R4; blocked by 10, 11) |
| Head (composition) | 16 | `ProcessValidate` — the ROP pipe (blocked by all I/O + logic modules: 06–09, 12–15) |
| Wiring | 17 | concrete `Deps` + mount the command, exposes the CLI (blocked by everything, 01–16) |

Test formulas designed alongside (in `module-tree.md`/`contracts.md`, realized per-ticket): **24 unit
tests** (logic modules, formula `1 happy + Σ antecedent branches`) + **6 component scenarios** (`1
happy + 5 adapter branches`, `contracts.md` Gherkin outline, tag `@wip` until `@wirth-tester`/`fagan`
sign-off).

### Open questions / tech debt

None open — the design package carries no `TODO`/open-question markers; OQ3 (exit-code grid) and OQ5
(provenance echo) are already resolved and folded into `report.schema.json` `x-exit-codes` /
`provenance`. Comparison algorithm is **ported, not reinvented** from `sandbox/ALGORITHM.md`
(`check.mjs`) — see `module-tree.md` intro.

## Completeness check (this stage's antecedent, verified before assembly)

- [x] Design docs present: `use-case.md`, `module-tree.md`, `contracts.md`, `c4.md`, `CONTEXT.md`, 3 ADRs.
- [x] Contract frozen: `config.schema.json` + `report.schema.json`, both `x-frozen: 2026-07-16`.
- [x] Tickets cut: 17/17 (`ticket-01.md` … `ticket-17.md`), each with `blocked_by` dependency chain.
- [x] Single slice (`#slices = 1`), no `CONTEXT-MAP.md` needed (one bounded context).

<!-- DONE: planner slice-01-validate -->

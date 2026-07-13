# PLAN — slice `compat-check` (path index + Gate #1 summary)

> Assembled by `wirth-planner`. This file is an **index**, not a copy: everything below links to the
> real artifact except the two inlined sections marked **(verbatim)** — the head-pipe pseudocode and
> the failure-mode map, per the plan-index skill's one allowed content copy.
>
> Backfill note: this stage was skipped in the first pass (ticketer → mills jumped over `wirth-planner`);
> `plan-reviewer` (mills) flagged the gap as advisory **A1** in `.agent/plan-reviewer/plan-review.md`
> (verdict `GO_WITH_CHANGES`, round 1, no blocker) and named this file as the fix-in-flight. All source
> artifacts existed and validated before this file was written; nothing here was designed — only indexed.

## Slice

**Slice 01 — Validate consumer↔provider OpenAPI compatibility** (`internal/compat-check/`). One external
input (CLI one-shot `pinout-openapi <contract-tests.yaml>`) = one slice, per `wirth-triage` (`level=modular`)
and `wirth-slicer`. Target shape: `cli`.

## Path index

| Artifact | Path |
|---|---|
| BRD (FRD-level doc) | `.agent/planner/brd.md` |
| Triage verdict | `.agent/triage.md` |
| Slice backlog | `.agent/planner/slices.md` |
| Use case (Cockburn, fully-dressed) | `docs/design/slice-compat-check/use-case.md` |
| Module tree + head pipe | `docs/design/slice-compat-check/module-tree.md` |
| Module contracts + component scenarios | `docs/design/slice-compat-check/contracts.md` |
| C4 (C2 Container + C3 Component) | `docs/design/slice-compat-check/c4.md` |
| Bounded-context glossary | `docs/design/slice-compat-check/CONTEXT.md` |
| Frozen contract — input DTO | `api-specification/config.schema.json` |
| Frozen contract — output canon | `api-specification/report.schema.json` |
| Frozen contract — exit codes | `api-specification/exit-codes.md` |
| Rollout baseline (canary + golden signals) | `.agent/planner/rollout-plan.md` |
| Plan-review verdict (Gate #1 critic) | `.agent/plan-reviewer/plan-review.md` |
| Tickets (10, scaffold → component RED → modules → README) | `docs/design/slice-compat-check/tickets/ticket-01.md` … `ticket-10.md` |

## Operator summary (Gate #1)

### Head module — functional style (verbatim from `module-tree.md`)

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

ROP short-circuit: a failing step skips the rest; the `ErrXxx` rises untransformed; only the `cli/` door
maps `error → error.code → exit code`. `INCOMPATIBLE` is a core-logic **verdict**, not a pipe error —
`Compare` always returns a `Report` (written for both verdicts); only genuine failures short-circuit.

### Failure-mode map (verbatim from `api-specification/exit-codes.md`)

| `error.code` | exit | verdict | operator action |
|---|---|---|---|
| — (`compatible`) | `0` | compatible | none |
| `INCOMPATIBLE` | `1` | incompatible | fix consumer/provider before merge (see `findings[]`) |
| `CONFIG_INVALID` | `2` | error | fix `contract-tests.yaml` |
| `SPEC_UNREADABLE` | `2` | error | fix consumer `spec_path` |
| `SPEC_PARSE_ERROR` | `2` | error | fix consumer spec |
| `OP_NOT_IN_CONSUMER` | `2` | error | fix config or consumer spec |
| `PROVIDER_UNREACHABLE` | `3` | error | check `spec_url` / network |
| `PROVIDER_TIMEOUT` | `3` | error | check network / raise `timeout` |
| `PROVIDER_PARSE_ERROR` | `3` | error | fix provider spec |
| `REPORT_WRITE_ERROR` | `3` | error | check path / permissions / disk |

Streams: stdout = machine JSON report only (always emitted); stderr = logs + `error.code` diagnostic on
any non-zero exit. All 9 `error.code`s trace 1:1 to the use case's 9 Extensions (`use-case.md`).

### Ticket order (10 tickets, dependency-ordered)

| # | Ticket | Type | `blocked_by` | What it builds |
|---|---|---|---|---|
| 01 | `ticket-01.md` | scaffold | — | clone `template-go-cli` → runnable placeholder (build green, `version` exits 0) |
| 02 | `ticket-02.md` | component RED | 01 | realize the 9-scenario set (CS1–CS9), drive RED, tag `@wip` |
| 03 | `ticket-03.md` | module (`io:none`) | 01, 02 | `config/` — read `contract-tests.yaml` + valid-by-construction VOs |
| 04 | `ticket-04.md` | module (`io:http`) | 01, 02, 03 | `spec/` — load + `$ref`-resolve OpenAPI 3.x (consumer file, provider file\|HTTP) |
| 05 | `ticket-05.md` | module (`io:none`) | 01, 02, 03, 04 | `compare/` — the 5 compatibility rules (pure core) |
| 06 | `ticket-06.md` | module (`io:none`) | 01, 02, 03, 04, 05 | package root — head pipe + DTOs + port interfaces + sentinels |
| 07 | `ticket-07.md` | module (`io:none`) | 01, 02, 05, 06 | `report/` — persist the canon report (file + stdout) |
| 08 | `ticket-08.md` | module (`io:none`, `cli-io`) | 01, 02, 06 | `cli/` — driving door (args → `Request`, `Result` → exit code) |
| 09 | `ticket-09.md` | module (`io:none`) | 01, 02, 03, 04, 05, 06, 07, 08 | wiring — assemble concrete `Deps`, mount real CLI in `cmd/app/main.go` |
| 10 | `ticket-10.md` | module (`io:none`) | 01, 02 | README — usage + build-run + Карта режимов отказа |

Shape: scaffold (01) → component RED (02) → modules (03–09, dependency-ordered pure-core-first) → docs (10,
parallelizable with 09 per `mills`' S1 check). `@fagan` strips `@wip` at slice acceptance, not a ticket here.

### Plan-review verdict

**GO_WITH_CHANGES** (mills, round 1, `.agent/plan-reviewer/plan-review.md`) — no blocker; all deterministic
validators green; per-ticket semantic walk clean; all 9 `error.code`s traced to owning ticket + scenario/unit.

### Open questions / tech debt (advisories, non-blocking — carried from plan-review)

- **A1** (this file) — `wirth-planner` stage was skipped in the first pass; fixed by this document.
- **A2** — `validate-frd` defaults to `.agent/planner/frd.md` (absent); the FRD-level doc is
  `.agent/planner/brd.md` (`node harness/validate-frd.mjs .agent/planner/brd.md` = exit 0). Path
  convention only. Optional fix: symlink/alias `frd.md` → `brd.md`.
- **A3** — `TASK.md` has no numbered `## Definition of done`, so `validate-plan`'s DoD-closure check
  auto-skipped. Coverage was verified manually (all 9 `error.code`s + verdict owned — see plan-review §S3).
  Optional: add a `## Definition of done` section to `TASK.md`.
- **A4** — the determinism acceptance criterion (`slices.md`: "two runs on frozen inputs yield diff-empty
  report bodies") isn't pinned to one ticket; it's emergent from the pure compare core + timestamp-excluded
  report, covered by `@fagan`'s DoD inspection. Optional: add an explicit determinism assertion to CS1 in
  `ticket-02.md`.
- **A5** — `ticket-07.md` `blocked_by` omits `03` though `report/writer.go` uses `config.Settings`; build
  order is correct transitively (05 and 06 both depend on 03). Cosmetic, no reordering needed.

## Rollout baseline

Canary = staged binary-version rollout across the CI fleet (10% cohort, 24h or ≥200 invocations), 4
golden signals (latency/traffic/errors/saturation), rollback = repin previous version. Full detail:
`.agent/planner/rollout-plan.md`.

<!-- DONE: wirth-planner slice-compat-check -->

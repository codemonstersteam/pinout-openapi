# Module contracts — slice `compat-check`

> Stage 5 (program-design / Wirth). Antecedent → consequent per module; every module carries a
> mandatory `io:` field (the sole router key the ticket-writer uses for io sub-skills).
> Tree + head pipe: [`module-tree.md`](./module-tree.md). Frozen contract: `api-specification/`
> (`config.schema.json` input, `report.schema.json` output, `exit-codes.md`).

## Error model (`error.code` → exit code — mapped ONLY at the `cli/` door)

| `error.code` | exit | sentinel | raised by |
|---|---|---|---|
| — (verdict `compatible`) | `0` | — | `Compare` → door maps verdict |
| `INCOMPATIBLE` | `1` | — (verdict `incompatible`) | `Compare` → door maps verdict |
| `CONFIG_INVALID` | `2` | `ErrConfigInvalid` | `ConfigReader` (unreadable) / `NewConfig` (invalid) |
| `SPEC_UNREADABLE` | `2` | `ErrSpecUnreadable` | `ConsumerSpecReader` |
| `SPEC_PARSE_ERROR` | `2` | `ErrSpecParse` | `ConsumerSpecReader` |
| `OP_NOT_IN_CONSUMER` | `2` | `ErrOpNotInConsumer` | `NewCheckedOperations` |
| `PROVIDER_UNREACHABLE` | `3` | `ErrProviderUnreachable` | `ProviderSpecReader` |
| `PROVIDER_TIMEOUT` | `3` | `ErrProviderTimeout` | `ProviderSpecReader` |
| `PROVIDER_PARSE_ERROR` | `3` | `ErrProviderParse` | `ProviderSpecReader` |
| `REPORT_WRITE_ERROR` | `3` | `ErrReportWrite` | `ReportWriter` |

---

## Driving adapter

### cli.Parse / cli.Exit (the door)

- **Signature:** `Parse(args []string) -> Result<Request, Error>` · `Exit(Result<Outcome, Error>) -> (exitCode, stdout, stderr)`
- **Input (data):** raw `os.Args` (parse); the head's `Result<Outcome, Error>` (exit).
- **Dependencies (deps):** — (cobra only; no domain, no I/O object).
- **io:** `none` (inbound door — parses/serializes, not an I/O object).
- **What it does:** flags/positional `<config>` → one flat `Request`; `Result`/`Outcome` → exit code + stdout(report)/stderr(logs).
- **Antecedent:** a CLI invocation; `<config>` positional present (or default path).
- **Consequent:** success → `Request{ConfigPath}`; the outcome serialized: verdict→`0|1`, `error.code`→`2|3`, report on stdout, logs on stderr. No domain logic.

---

## Pure core — `config/` (constructors, valid-by-construction)

### NewConfig

- **Signature:** `NewConfig(raw: RawConfig) -> Result<Config, Error>`
- **Input (data):** `RawConfig` (the config bytes).
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** unmarshal YAML, assemble the validated value-objects (`NewConsumer`/`NewProvider`/`NewSettings`) into a `Config`.
- **Antecedent:** `raw` is bytes; must be parseable YAML matching `config.schema.json` (draft-2020-12, `additionalProperties:false`).
- **Consequent:** success → a fully-validated `Config`. Failure → `ErrConfigInvalid` (unparseable YAML, or any child VO rejects) → `CONFIG_INVALID`.

### NewConsumer / NewProvider / NewProviderSource / NewOperation / NewMethod / NewOperationPath / NewSettings

- **Signature:** `NewX(...) -> Result<X, Error>` (each a value-object factory with unexported fields).
- **Dependencies (deps):** — · **io:** `none`
- **Antecedents / consequents (each field range-checked; failure → `ErrConfigInvalid` → `CONFIG_INVALID`):**
  - `NewMethod(raw)` → verb ∈ `{get,put,post,delete,options,head,patch,trace}`, case-insensitive → normalized lower-case.
  - `NewOperationPath(raw)` → non-empty, leading `/`.
  - `NewOperation(path, method)` → composes the two validated VOs.
  - `NewProviderSource(url, path)` → **exactly one** of `url`/`path` present (both/neither → invalid); output distinguishes `kind ∈ {http, file}`.
  - `NewSettings(logLevel, saveJSON, reportFile, timeout)` → `log_level ∈ {debug,info,warn,error}` (default `info`); `save_json_report` default `true`; `timeout` integer `1..600` (default `30`); `json_report_file` **required when `save_json_report=true`** (default `compatibility_report.json`).
  - `NewConsumer(name, operations)` → `name` non-empty (trim≠""); `operations` ≥1 item.
  - `NewProvider(name, source)` → `name` non-empty.

---

## Pure core — `compare/` (the 5 compatibility rules)

### NewCheckedOperations (consumer pre-check)

- **Signature:** `NewCheckedOperations(ops: []Operation, consumer: ConsumerSpec) -> Result<CheckedOps, Error>`
- **Input (data):** the configured operations + the loaded consumer spec (two cohesive domain entities, à la `buildResponse(cmd, id)`).
- **Dependencies (deps):** — · **io:** `none`
- **What it does:** confirm every configured `path`+`method` exists in the **consumer** spec.
- **Antecedent:** `ops` are validated; `consumer` is a parsed `Spec`.
- **Consequent:** success → `CheckedOps` (all present). Failure → `ErrOpNotInConsumer` → `OP_NOT_IN_CONSUMER` (misconfiguration, **not** an incompatibility).

### checkPresence / checkRequestContravariance / checkResponseCovariance / checkStatusCodes / checkContentTypes

- **Signature:** `checkXxx(consumerOp, providerOp) -> []Finding`
- **Dependencies (deps):** — · **io:** `none`
- **What each does / consequent** (direction: consumer expectations must be satisfiable by provider master; `allOf` flattened by the reader; `oneOf`/`anyOf`/enum/`format` deferred):
  - `checkPresence` → provider exposes `path`+`method`? absent → `Finding{rule: presence}`.
  - `checkRequestContravariance` → `provider.required(request) ⊆ consumer.sent(request)`? extra provider-required item → `Finding{rule: request_contravariance}`.
  - `checkResponseCovariance` → `consumer.read(response) ⊆ provider.provided(response)`? missing field → `Finding{rule: response_covariance}`.
  - `checkStatusCodes` → consumer codes ⊆ provider codes? missing code → `Finding{rule: status_codes}`.
  - `checkContentTypes` → content-types matched? mismatch → `Finding{rule: content_types}`.
- **Antecedent:** both operand operations resolved from parsed specs.
- **Consequent:** `[]Finding` (empty ⇒ that rule passed). `rule` value is **underscore-form** per the frozen `report.schema.json`.

### NewComparison / compareOperation / aggregateVerdict / buildReport

- **NewComparison(checkedOps, consumer, provider) -> Comparison** · io: `none` — unites the three entities into one (one-data-argument discipline).
- **compareOperation(comparisonOp) -> OperationResult** · io: `none` — runs the 5 rules, concatenates findings; `status = incompatible` iff `findings ≥ 1`, else `compatible`.
- **aggregateVerdict(results) -> Verdict** · io: `none` — any incompatible operation ⇒ overall `incompatible`.
- **buildReport(config, results, providerSource) -> Report** · io: `none` — assemble the canon `report.schema.json` shape: `verdict`, `consumer.name`, `provider.{name,source}`, `operations[]{path,method,status,findings[]}`. **Invariant (by construction):** `status=compatible ⇒ findings=[]`; `status=incompatible ⇒ findings≥1`.
- **Compare(comparison) -> Report** · io: `none` — the pure core entry: `compareOperation` per op → `aggregateVerdict` → `buildReport`. Returns a `Report` for **both** verdicts (never an error).

---

## Driven I/O objects (autonomous — head sees only the interface; not unit-tested)

### ConfigReader

- **Signature:** `Read(path: string) -> Result<RawConfig, Error>` · **Input:** config path · **Deps:** — · **io:** `file`
- **What it does:** read `contract-tests.yaml` bytes. Pipe — no transformation.
- **Consequent:** success → `RawConfig` bytes. Failure → `ErrConfigInvalid` (missing/unreadable) → `CONFIG_INVALID`.

### ConsumerSpecReader

- **Signature:** `Load(path: string) -> Result<Spec, Error>` · **Input:** `consumer.spec_path` · **Deps:** — · **io:** `file`
- **What it does:** read the consumer file + `kin-openapi` load (local+external `$ref` resolved, `allOf` flattened). Pipe.
- **Consequent:** success → `Spec`. Failures → `ErrSpecUnreadable` (`SPEC_UNREADABLE`) · `ErrSpecParse` (`SPEC_PARSE_ERROR`).

### ProviderSpecReader

- **Signature:** `Load(source: ProviderSource) -> Result<Spec, Error>` · **Input:** resolved `ProviderSource` (`kind ∈ {http,file}`) · **Deps:** — · **io:** `http`
- **What it does:** **HTTP(S) GET `spec_url` (unauthenticated, bounded by `settings.timeout`) OR read `spec_path`** + `kin-openapi` load. Pipe. Design sub-skill: **`http-io`** (HTTP variant — load/payload budgets, provider fetch, failure branches).
- **Consequent:** success → `Spec`. Failures → `ErrProviderUnreachable` (`PROVIDER_UNREACHABLE`, non-2xx/unreachable) · `ErrProviderTimeout` (`PROVIDER_TIMEOUT`) · `ErrProviderParse` (`PROVIDER_PARSE_ERROR`). Never mutates the master spec.

### ReportWriter

- **Signature:** `Write(report: Report, settings: Settings) -> Result<Outcome, Error>` · **Input:** the `Report` · **Deps:** — · **io:** `file`
- **What it does:** write JSON report to `json_report_file` when `save_json_report=true`, **plus** the machine-readable stdout line (always). Pipe.
- **Consequent:** success → `Outcome{Verdict}` (carries verdict for the door's exit mapping). Failure → `ErrReportWrite` (`REPORT_WRITE_ERROR`, bad path/permission/disk).

---

## Head

### ProcessCompatCheck

- **Signature:** `ProcessCompatCheck(req: Request, deps: Deps) -> Result<Outcome, Error>`
- **Input (data):** the `Request` (one external input). · **Deps:** `Deps{Config, Consumer, Provider, Report}` (four **port interfaces**, never raw `*http.Client`/`*os.File`). · **io:** `none`
- **What it does:** the linear ROP pipe of `module-tree.md`. No branching of its own.
- **Antecedent:** a valid `Request{ConfigPath}`. **Consequent:** `Outcome{Verdict}` or the first step's `ErrXxx` (untransformed).

---

## Component scenarios (design-only; realized by `@wirth-tester`)

**Formula** — `N_failure = 1 (happy) + Σ (distinguishable I/O-adapter branches)`. Boundaries / input-value
/ pure-logic outcomes stay **unit**-level (module-tree §formula). Each I/O-adapter branch is one
distinguishable **outcome** = one scenario. All tagged `@wip` (RED until realized).

**Classification of the 9 `error.code`s (UC Extensions) → owner:**

| UC Ext. | `error.code` | exit | Owner branch | Test level |
|---|---|---|---|---|
| MSS row0 | — (`compatible`) | 0 | core verdict | **component** (happy, CS1) |
| 5a | `INCOMPATIBLE` | 1 | core verdict (pure logic; read succeeded) | **unit** boundaries (compare rules) **+** CS2 end-to-end verdict guard |
| 1a | `CONFIG_INVALID` | 2 | `ConfigReader` (unreadable, I/O) — content boundaries are units | **component** (CS3) + units |
| 2a | `SPEC_UNREADABLE` | 2 | `ConsumerSpecReader` (I/O) | **component** (CS4) |
| 2b | `SPEC_PARSE_ERROR` | 2 | `ConsumerSpecReader` (I/O) | **component** (CS5) |
| 3a | `OP_NOT_IN_CONSUMER` | 2 | `NewCheckedOperations` (pure logic; read succeeded) | **unit** boundary |
| 4a | `PROVIDER_UNREACHABLE` | 3 | `ProviderSpecReader` (I/O) | **component** (CS6) |
| 4b | `PROVIDER_TIMEOUT` | 3 | `ProviderSpecReader` (I/O) | **component** (CS7) |
| 4c | `PROVIDER_PARSE_ERROR` | 3 | `ProviderSpecReader` (I/O) | **component** (CS8) |
| 7a | `REPORT_WRITE_ERROR` | 3 | `ReportWriter` (I/O) | **component** (CS9) |

**Anti-gaming reconciliation (Step 8.6):** `#component_failure_scenarios == #distinguishable_I/O-adapter_branches == #error.codes_of_those_branches`
→ **7 == 7 == 7** (`CONFIG_INVALID, SPEC_UNREADABLE, SPEC_PARSE_ERROR, PROVIDER_UNREACHABLE, PROVIDER_TIMEOUT, PROVIDER_PARSE_ERROR, REPORT_WRITE_ERROR`).
The remaining 2 codes are **pure-logic outcomes** (the external read succeeded, then logic decided): `OP_NOT_IN_CONSUMER`
→ unit; `INCOMPATIBLE` → unit boundaries. All 9 codes traced. This **refines** the use-case's "9 Extensions →
9 scenario names": 2 of the 9 are logic/input boundaries, not I/O-adapter failures.

**Scenario set** (formula count = **1 happy + 7 failure = 8**; **CS2** is a 9th success-family verdict
guard for exit-1 black-box coverage — a realized extension of the happy scenario, outside the failure Σ):

| # | Scenario (verbatim UC wording where applicable) | Family | Given → outcome | Exit | @wip |
|---|---|---|---|---|---|
| CS1 | all configured operations compatible (+ deterministic report) | happy | valid config, both specs load, no findings → report `verdict=compatible`; two identical runs → byte-identical `json_report_file` (generated_at excluded) | 0 | yes |
| CS2 | ≥1 operation fails a compatibility rule | verdict | valid config, both specs load, a rule fails → report `verdict=incompatible`, `findings≥1` | 1 | yes |
| CS3 | config missing / invalid / conflict / empty operations | I/O-adapter | unreadable `contract-tests.yaml` → `CONFIG_INVALID` | 2 | yes |
| CS4 | consumer spec missing or unreadable | I/O-adapter | `consumer.spec_path` unreadable → `SPEC_UNREADABLE` | 2 | yes |
| CS5 | consumer spec not valid OpenAPI 3.x / unparseable | I/O-adapter | consumer doc unparseable → `SPEC_PARSE_ERROR` | 2 | yes |
| CS6 | provider spec unreachable / HTTP non-2xx | I/O-adapter | `spec_url` stub returns 500 / unreachable → `PROVIDER_UNREACHABLE` | 3 | yes |
| CS7 | provider fetch exceeds `settings.timeout` | I/O-adapter | `spec_url` stub delays past `timeout` → `PROVIDER_TIMEOUT` | 3 | yes |
| CS8 | provider spec not valid OpenAPI 3.x / unparseable | I/O-adapter | provider doc unparseable → `PROVIDER_PARSE_ERROR` | 3 | yes |
| CS9 | JSON report file cannot be written | I/O-adapter | unwritable `json_report_file` → `REPORT_WRITE_ERROR` | 3 | yes |

> CLI component tests: static binary (`CGO_ENABLED=0`), one-shot against config fixtures, asserting
> `(exit code, stdout JSON)` as a black box. The provider `spec_url` fetch gets a **real-protocol HTTP stub**
> in compose (CS6/CS7); only the core is never stubbed.

**CS1 determinism acceptance clause (Gate #1 A4 — pinned to CS1, single owning ticket 02).** Because the
report canon marks `generated_at` optional/excluded (`report.schema.json`), the JSON report MUST be
byte-stable across identical runs. Concrete, testable assertion added to CS1 (RED @wip until realized):

- **Given** a valid config with `save_json_report=true` and both specs loading with no findings,
- **When** the tool is run **twice** on the **byte-identical** inputs (same `contract-tests.yaml`, same
  consumer + provider spec bytes),
- **Then** both runs exit `0`, and the two `json_report_file` outputs are **byte-identical** once
  `generated_at` is excluded from the diff (held fixed or filtered before comparison) — i.e.
  `diff <(jq 'del(.generated_at)' run1.json) <(jq 'del(.generated_at)' run2.json)` is empty.
- **Provided by:** `buildReport` (deterministic assembly — no map-iteration order, no clock except
  `generated_at`) → `ReportWriter.Write` (stable serialization). Only `generated_at` may differ between runs.

## Gherkin ↔ module reconciliation (each Then → a graph node)

| Scenario | Then-step | Provided by (graph node / door mapping) |
|---|---|---|
| CS1 happy | exit 0 | `cli.Exit`: `Outcome.Verdict=compatible → 0` |
| CS1 happy | stdout report `verdict=compatible`, `operations[].status=compatible`, `findings=[]` | `buildReport` (invariant) → `ReportWriter` (stdout line, success) |
| CS1 happy | report file written at `json_report_file` | `ReportWriter.Write` (success branch) |
| CS1 happy | two identical runs → byte-identical `json_report_file` (excl. `generated_at`) | `buildReport` (deterministic assembly) → `ReportWriter.Write` (stable serialization) |
| CS2 verdict | exit 1 | `cli.Exit`: `Outcome.Verdict=incompatible → 1` |
| CS2 verdict | stdout report `verdict=incompatible`, `findings[].rule ∈ {presence,…,content_types}` | `compareOperation` + `aggregateVerdict` + `buildReport` |
| CS3 config | exit 2 + `error.code=CONFIG_INVALID` | `ConfigReader.Read` (Failure: `ErrConfigInvalid`) → `cli.Exit` map → 2 |
| CS4 consumer unreadable | exit 2 + `SPEC_UNREADABLE` | `ConsumerSpecReader.Load` (Failure: `ErrSpecUnreadable`) → 2 |
| CS5 consumer unparseable | exit 2 + `SPEC_PARSE_ERROR` | `ConsumerSpecReader.Load` (Failure: `ErrSpecParse`) → 2 |
| CS6 provider unreachable | exit 3 + `PROVIDER_UNREACHABLE` | `ProviderSpecReader.Load` (Failure: `ErrProviderUnreachable`) → 3 |
| CS7 provider timeout | exit 3 + `PROVIDER_TIMEOUT` | `ProviderSpecReader.Load` (Failure: `ErrProviderTimeout`) → 3 |
| CS8 provider unparseable | exit 3 + `PROVIDER_PARSE_ERROR` | `ProviderSpecReader.Load` (Failure: `ErrProviderParse`) → 3 |
| CS9 report write | exit 3 + `REPORT_WRITE_ERROR` | `ReportWriter.Write` (Failure: `ErrReportWrite`) → 3 |
| CS9 report write | provider master spec **not** mutated | `ProviderSpecReader` read-only (no write path exists) |

**Reconciliation checklist:** [x] every node exists in the contracts above · [x] each error Then has a
Failure branch with the same sentinel · [x] the door records `error.code → exit` for each · [x] all
`.feature` Thens covered — no orphan line. `OP_NOT_IN_CONSUMER` (3a) and `INCOMPATIBLE` rule-boundaries
are **unit** boundaries (module-tree §formula), by design not `.feature` lines.

<!-- DONE: moduledesigner slice-compat-check -->
</content>

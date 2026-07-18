# Module contracts — slice-01-validate

> Antecedent → consequent per behavioral module (`program-design` Step 5). Wiring (`main`,
> `register.go`) and pure type/sentinel files (`domain.go`, `errors.go`) carry no contract. Every
> module carries the mandatory **`io:`** field. Neighbor: [`module-tree.md`](./module-tree.md),
> [`use-case.md`](./use-case.md). Error→exit mapping is `report.schema.json` `x-exit-codes`.

## Error model (edge-mapped once, in `cli.ResolveExitCode`)

| Sentinel (rises untransformed) | `error.code` | exit | class |
|---|---|---|---|
| `ErrConfig` | `CONFIG_ERROR` | 2 | config / usage (never in `errors[]`) |
| `ErrFileNotFound` | `FILE_NOT_FOUND` | 3 | io·parse of an input artifact |
| `ErrParse` | `PARSE_ERROR` | 3 | io·parse of an input artifact |
| `ErrHTTP` | `HTTP_ERROR` | 3 | provider `spec_url` unreachable / non-2xx |
| `ErrTimeout` | `TIMEOUT_ERROR` | 3 | provider `spec_url` fetch > `settings.timeout` |
| (no error — success path) `Violation{code}` folded into report | `OP_NOT_IN_PROVIDER` · `MISSING_REQUIRED_REQUEST_FIELD` · `READS_FIELD_NOT_PROVIDED` · `TYPE_MISMATCH` | 1 | **verdict** (domain said no) — not a pipe error |
| — | (none) | 0 | compatible |

---

## Contracts

### cli.Parse

- **Signature:** `Parse(args: []string) -> Result[Invocation, Error]`
- **Input (data):** the process argv (`os.Args`).
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** the ingress door — parse `validate <config.yaml>` + flags (`--json`, `--verbose`,
  `--version`, `--help`) into one flat `Invocation` DTO. Parsing only, no logic, no file I/O.
- **Antecedent:** `args` is the real invocation vector.
- **Consequent:**
  - Success: `Invocation{ConfigPath, JSONReport, Verbose}` — a single flat DTO; config **path** only
    (the file is read downstream by `ConfigStore`).
  - Failure: `ErrConfig` (missing positional `<config.yaml>` / unknown flag) → `CONFIG_ERROR`.

### ProcessValidate (head)

- **Signature:** `ProcessValidate(inv: Invocation) -> Result[Report, Error]`
- **Input (data):** one `Invocation` DTO.
- **Dependencies (deps):** `Deps{ ConfigStore, ContractStore, BuildSpecLoader, ReportWriter, Clock, *slog.Logger }`
  (autonomous I/O objects + orthogonal tools — **no** raw `*os.File`/`*http.Client`). Note the change
  (ADR-0005): `Deps` carries **`BuildSpecLoader func(Settings) SpecLoader`** — a wired-once factory —
  **not** a ready `SpecLoader`, so the head can construct the loader late, bounded by the real
  `cfg.Settings.Timeout`.
- **io:** `none` (composition root — a pipe of already-tested parts)
- **What it does:** the linear ROP pipe of `module-tree.md`; no branching of its own. The one non-adapter
  step it now performs is calling `d.BuildSpecLoader(cfg.Settings)` (a pure, total factory — never an
  error step) to obtain the timeout-bounded `SpecLoader` before `loader.Load(cfg.Provider)`.
- **Antecedent:** a valid `Invocation`.
- **Consequent:**
  - Success: `Report` (schema-valid; `compatible ⇔ errors == []`).
  - Failure: any child's sentinel, risen untransformed (short-circuit).

### ConfigStore.Load

- **Signature:** `Load(path: string) -> Result[RawConfig, Error]`
- **Input (data):** the config file path (from `Invocation`).
- **Dependencies (deps):** — (the OS filesystem is encapsulated in the object)
- **io:** `none` (filesystem I/O pipe — no transformation; ADR-0003)
- **What it does:** read the config file bytes and YAML-unmarshal into an **unvalidated** `RawConfig`.
- **Antecedent:** `path` non-empty.
- **Consequent:**
  - Success: `RawConfig` (structurally decoded, not yet valid).
  - Failure: `ErrConfig` (not found / unreadable / malformed YAML) → `CONFIG_ERROR`, exit 2.

### NewConfig

- **Signature:** `NewConfig(raw: RawConfig) -> Result[Config, Error]`
- **Input (data):** one `RawConfig`.
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** valid-by-construction constructor — validate every field against
  `config.schema.json` invariants; illegal states unrepresentable past this boundary.
- **Antecedent:** `RawConfig` decoded.
- **Consequent:**
  - Success: `Config` (unexported fields) — non-empty names; `operations` ≥ 1, each `path` matches
    `^/` and `method` in the enum; provider source **exactly one** of `spec_path`/`spec_url`; settings
    defaulted + in-range (`timeout > 0`, `log_level` enum).
  - Failure: `ErrConfig` (any invariant) → `CONFIG_ERROR`, exit 2. *(schema-invalidity is UNIT-covered;
    it shares the `CONFIG_ERROR` code with the `ConfigStore` read failure — ADR-0002/0003.)*

### ContractStore.Load

- **Signature:** `Load(path: string) -> Result[ConsumedContract, Error]`
- **Input (data):** `consumer.consumed_contract_path`.
- **Dependencies (deps):** —
- **io:** `none` (filesystem I/O pipe; ADR-0003)
- **What it does:** read + parse the E-harness `consumed-contract` artifact (already typed
  per-operation `{sends, reads}` + `provenance`). No type inference — types arrive typed.
- **Antecedent:** `path` non-empty.
- **Consequent:**
  - Success: `ConsumedContract{ Operations: []{ Ref, Sends, Reads }, Provenance }`.
  - Failure: `ErrFileNotFound` (missing/unreadable) → `FILE_NOT_FOUND`, exit 3; `ErrParse`
    (unparseable) → `PARSE_ERROR`, exit 3.

### BuildSpecLoader (factory) — *late construction, ADR-0005*

- **Signature:** `BuildSpecLoader(s: Settings) -> SpecLoader` (as a `Deps` field:
  `BuildSpecLoader func(domain.Settings) SpecLoader`).
- **Input (data):** the validated `cfg.Settings` (from `NewConfig`), whose `Timeout` is guaranteed `> 0`.
- **Dependencies (deps):** — (a pure closure; the concrete wiring in `register.go` is
  `func(s Settings) SpecLoader { return provider.NewSpecLoader(s.Timeout) }`).
- **io:** `none` (a pure, **total** constructor call — never an error step; not unit-tested).
- **What it does:** resolves the frozen-design contradiction **(a) "SpecLoader bounded by
  settings.timeout"** vs **(b) "SpecLoader encapsulated, constructed once inside a fully-built Deps"**.
  Resolution: keep (a); relax (b) — the *wired-once, encapsulated* object becomes this **factory**, and
  the `SpecLoader` itself is built **once per invocation inside the head, after `NewConfig`**, receiving
  `s.Timeout`. This removes the former `defaultProviderTimeoutSeconds = 30` hardcode (which made
  `settings.timeout` inert and component scenario 6 unable to emit `TIMEOUT_ERROR`). `provider.NewSpecLoader(timeoutSeconds int)`
  keeps its signature — only its **call site moves** from `register.go` (eager, hardcoded 30) into the
  head via this factory (lazy, `cfg.Settings.Timeout`).
- **Antecedent:** `Settings` valid-by-construction (`Timeout > 0`).
- **Consequent:** Success: a `SpecLoader` whose HTTP branch is bounded by `s.Timeout`. No failure path.

### SpecLoader.Load

- **Signature:** `Load(p: ProviderConfig) -> Result[ProviderSpec, Error]` *(unchanged — only the
  construction lifecycle moved; see `BuildSpecLoader`, ADR-0005)*
- **Input (data):** one `ProviderConfig` (`name` + `spec_path` XOR `spec_url`).
- **Dependencies (deps):** — (`kin-openapi` loader + `*http.Client` + timeout + bearer token are
  encapsulated inside the object; the head sees only `Load`). The instance is **built late** by
  `BuildSpecLoader(cfg.Settings)` inside the head — **not** pre-built in `Deps` — so its `timeout` is the
  real `cfg.Settings.Timeout`, never a wiring-time default (ADR-0005).
- **io:** `http` → designed with the **`http-io`** skill (timeout & payload budgets, provider spec as
  the frozen machine contract, real-protocol stub for component tests). ADR-0001.
- **What it does:** acquire the provider OpenAPI (`spec_path` = local file OR `spec_url` = HTTP GET
  with `Authorization: Bearer $PINOUT_PROVIDER_TOKEN` from env, bounded by `settings.timeout`) and
  parse + resolve `$ref` via `kin-openapi`. Pure pipe: no domain logic (that is `compare`).
- **Antecedent:** `ProviderConfig` valid (exactly-one source — guaranteed by `NewConfig`).
- **Consequent:**
  - Success: `ProviderSpec` (parsed `*openapi3.T`, `$ref` resolved).
  - Failure: `ErrFileNotFound` (path missing) → `FILE_NOT_FOUND`/3; `ErrParse` (bad OpenAPI/YAML) →
    `PARSE_ERROR`/3; `ErrHTTP` (unreachable/non-2xx) → `HTTP_ERROR`/3; `ErrTimeout` (fetch > timeout) →
    `TIMEOUT_ERROR`/3.

### NewComparison

- **Signature:** `NewComparison(cfg: Config, consumed: ConsumedContract, spec: ProviderSpec) -> Result[Comparison, Error]`
- **Input (data):** the three validated prior pipe outputs — **united** into one domain entity (the
  single-argument rule's sanctioned constructor: 2+ entities ⇒ a uniting constructor node).
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** build the `Comparison{ ScopedOps, Consumed, Spec, Provenance }` — the scope
  (`cfg.operations`) intersected with the consumed per-op data and the provider spec.
- **Antecedent:** all three inputs valid.
- **Consequent:** Success: one `Comparison`. Failure: — (inputs already valid by construction).

### DeriveProviderOperation

- **Signature:** `DeriveProviderOperation(spec: ProviderSpec, ref: OperationRef) -> Result[ProviderOperation, NotPresent]`
- **Input (data):** one `ProviderSpec` (+ the operation `ref` selector).
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** navigate `paths[path][method]`; derive `{ requires, provides }` over **body +
  parameters (path/query/header)** — required request fields/params (contravariant) and response body
  props (covariant). Implements **R1** (operation exists).
- **Antecedent:** `ProviderSpec` parsed.
- **Consequent:** Success: `ProviderOperation{ Requires, Provides }`. `NotPresent`: signals R1 →
  `CompareOperation` records `OP_NOT_IN_PROVIDER` (a **verdict**, exit 1 — not an error).

### CompareOperation

- **Signature:** `CompareOperation(consumedOp: ConsumedOperation, providerOp: ProviderOperation|NotPresent) -> []Violation`
- **Input (data):** one paired `{ consumedOp, providerOp }` operation (a domain struct).
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** the comparison core (ported verbatim from `sandbox/check.mjs::compareOp`), over
  **body + parameters**: **R1** absent ⇒ `OP_NOT_IN_PROVIDER`; **R2** `requires(provider) ⊆
  sends(consumer)` else `MISSING_REQUIRED_REQUEST_FIELD`; **R3** `reads(consumer) ⊆ provides(provider)`
  else `READS_FIELD_NOT_PROVIDED`; **R4** shared-field types match (`typesMatch`) else `TYPE_MISMATCH`.
- **Antecedent:** a paired operation.
- **Consequent:** Success: `[]Violation` (possibly empty ⇒ compatible for this op). No failure path —
  a violation is a **value**, not an error.

### CompareContracts

- **Signature:** `CompareContracts(c: Comparison) -> ComparisonOutcome`
- **Input (data):** one `Comparison`.
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** pure fold — for each scoped operation call `DeriveProviderOperation` +
  `CompareOperation`, `concatMap` the violations; compute `uncovered_operations[]` (provider ops
  outside `cfg.operations`, informational); carry `Provenance` through. The **loop lives here**, not
  in the head.
- **Antecedent:** valid `Comparison`.
- **Consequent:** Success: `ComparisonOutcome{ Violations, UncoveredOps, Provenance }`
  (`compatible ⇔ Violations == []`). No failure path.

### FoldReport

- **Signature:** `FoldReport(o: ComparisonOutcome) -> Report`
- **Input (data):** one `ComparisonOutcome`.
- **Dependencies (deps):** —
- **io:** `none`
- **What it does:** shape the frozen `report.schema.json` DTO — `schema_version="1.0"`,
  `compatible = Violations == []`, `errors[] = Violations`, top-level `provenance` echo,
  informational `uncovered_operations[]`.
- **Antecedent:** valid outcome.
- **Consequent:** Success: schema-valid `Report`. No failure path.

### ReportWriter.Write

- **Signature:** `Write(s: Settings, r: Report) -> Result[Report, Error]`
- **Input (data):** one `Report` (+ `settings` selecting persistence).
- **Dependencies (deps):** — (OS filesystem encapsulated)
- **io:** `none` (filesystem write pipe; ADR-0003)
- **What it does:** iff `settings.save_json_report`, atomically write the JSON to
  `settings.json_report_file`; ROP pass-through of `Report` unchanged (never partially mutates an
  existing report — minimal guarantee).
- **Antecedent:** schema-valid `Report`.
- **Consequent:** Success: the same `Report` (stdout copy printed by `main`). Failure: environment
  write failure → exit 3, logged to stderr — **not** a contract `error.code` (no enum code exists for
  it, so it is **not** a counted component branch); the happy scenario asserts the file **is** written.

### cli.ResolveExitCode

- **Signature:** `ResolveExitCode(res: Result[Report, Error]) -> int`
- **Input (data):** one `Result`.
- **Dependencies (deps):** —
- **io:** `none` (adapter — mechanical Result→exit; **not unit-tested**, asserted by component
  scenarios)
- **What it does:** the CLI status-line mapping (`cli-io` grid): `Ok(compatible)`→0;
  `Ok(incompatible)`→1; `ErrConfig`→2; `ErrFileNotFound|ErrParse|ErrHTTP|ErrTimeout`→3. `main` prints
  the report JSON to **stdout** and diagnostics to **stderr**.
- **Antecedent:** the head's `Result`.
- **Consequent:** one exit code in `{0,1,2,3}` per `report.schema.json` `x-exit-codes`.

---

## Component scenarios (DESIGN half — `@wip`, realized later by `@wirth-tester`)

Formula `N = 1 (happy) + Σ distinguishable adapter branches`. Adapters here = **filesystem** (config,
consumed-contract, `spec_path`, report) + **HTTP** (`spec_url`). A distinguishable branch ≡ one
consumer-visible `error.code`. The four **verdict** codes (exit 1) are the domain core → **UNIT**
tests over `CompareOperation`/`DeriveProviderOperation`, **not** component scenarios (ADR-0002).
Input-validation Extensions (config schema-invalidity, `spec_*` not-exactly-one) → **UNIT** boundaries
of `NewConfig`. Black box: assert **(exit code, stdout JSON)** against a config fixture in a container;
`spec_url` reached via a **real-protocol HTTP stub** (`http-io`).

**Gate:** `#component_failure_scenarios (5) == #distinguishable adapter branches (5) == #error.codes of
those branches (5)` → total **6** (1 happy + 5).

| # | Scenario (Cockburn wording verbatim) | Extension | Adapter branch (`error.code`) | exit | stdout | tag |
|---|---|---|---|---|---|---|
| 1 | happy: forward-compatible pair → report + `compatible=true` | MSS step 7 (F0) | — (SpecLoader + all loaders **success**) | 0 | schema-valid report, `errors=[]` | `@wip` |
| 2 | Config not found, unreadable, malformed YAML, schema-invalid, or `spec_path`/`spec_url` not exactly-one | 1a | `CONFIG_ERROR` (ConfigStore read/YAML) | 2 | no report (config-side breach) | `@wip` |
| 3 | `consumed-contract` or provider spec file missing / unreadable at its path | 2a | `FILE_NOT_FOUND` (ContractStore / SpecLoader file) | 3 | report w/ `errors[0].code=FILE_NOT_FOUND` | `@wip` |
| 4 | `consumed-contract` or provider spec unparseable (invalid OpenAPI / YAML) | 2b | `PARSE_ERROR` (ContractStore / SpecLoader parse) | 3 | report w/ `errors[0].code=PARSE_ERROR` | `@wip` |
| 5 | Provider `spec_url` unreachable (HTTP failure / non-2xx / connection refused) | 3a | `HTTP_ERROR` (SpecLoader HTTP) | 3 | report w/ `errors[0].code=HTTP_ERROR` | `@wip` |
| 6 | Provider `spec_url` fetch exceeds `settings.timeout` seconds | 3b | `TIMEOUT_ERROR` (SpecLoader HTTP timeout) | 3 | report w/ `errors[0].code=TIMEOUT_ERROR` | `@wip` |

> **NOT component scenarios** (mapped elsewhere, do not inflate the count): Extensions **5a–5d**
> (`OP_NOT_IN_PROVIDER`, `MISSING_REQUIRED_REQUEST_FIELD`, `READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH`,
> exit 1) → **unit** boundaries of `DeriveProviderOperation` (R1) and `CompareOperation` (R2/R3/R4);
> uncovered-provider-surface → informational report field (asserted inside scenario 1). Report-write
> failure has no enum `error.code` → not counted.

> **Delta (ADR-0005):** the scenario **count is unchanged (6)**. Scenario **6 (`TIMEOUT_ERROR`)** is now
> **realizable**: because the `SpecLoader` is built late with `cfg.Settings.Timeout` (no hardcoded 30),
> a `spec_url` stub that stalls past the config's `settings.timeout` actually trips `ErrTimeout`. The
> stub/fixture must set `settings.timeout` low enough to fire deterministically against the stall delay.

### Gherkin outline (designed set — realization = `.feature` by `@wirth-tester`)

```gherkin
@component @slice-01-validate
Feature: Forward-compatibility validation of a consumer↔provider pair

  @wip
  Scenario: forward-compatible pair yields a compatible report            # scenario 1 (happy)
    Given a config whose consumed-contract and provider spec are reachable and compatible
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 0
    And stdout is a schema-valid report with compatible=true and errors==[]
    And uncovered provider operations are listed in uncovered_operations[]

  @wip
  Scenario: Config not found, unreadable, malformed YAML, schema-invalid, or spec source not exactly-one
    Given a config file that cannot be read or is schema-invalid          # scenario 2 (CONFIG_ERROR)
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 2

  @wip
  Scenario: consumed-contract or provider spec file missing / unreadable at its path
    Given a config pointing at a consumed-contract path that does not exist  # scenario 3 (FILE_NOT_FOUND)
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "FILE_NOT_FOUND"

  @wip
  Scenario: consumed-contract or provider spec unparseable (invalid OpenAPI / YAML)
    Given a config pointing at a provider spec that is not valid OpenAPI   # scenario 4 (PARSE_ERROR)
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "PARSE_ERROR"

  @wip
  Scenario: Provider spec_url unreachable (HTTP failure / non-2xx / connection refused)
    Given a config with spec_url pointing at a stub that returns 503       # scenario 5 (HTTP_ERROR)
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "HTTP_ERROR"

  @wip
  Scenario: Provider spec_url fetch exceeds settings.timeout seconds
    Given a config with spec_url pointing at a stub that stalls past settings.timeout  # scenario 6 (TIMEOUT_ERROR)
    When I run `pinout-openapi validate config.yaml`
    Then the exit code is 3
    And stdout report errors[0].code == "TIMEOUT_ERROR"
```

## Gherkin-mapping (every `Then` → a call-graph node)

| Scenario | Then-step | Provided by (graph node / adapter mapping) |
|---|---|---|
| 1 happy | exit 0 | head Ok(compatible) → `cli.ResolveExitCode` → 0 |
| 1 happy | schema-valid report, `compatible=true`, `errors==[]` | `CompareContracts` (∅ violations) → `FoldReport` → `ReportWriter` (Success) |
| 1 happy | `uncovered_operations[]` listed | `CompareContracts` (uncovered detection) → `FoldReport` |
| 2 CONFIG_ERROR | exit 2 | `ConfigStore.Load`→`ErrConfig` (or `NewConfig`→`ErrConfig`) → `cli.ResolveExitCode` → 2 |
| 3 FILE_NOT_FOUND | exit 3 + `errors[0].code=FILE_NOT_FOUND` | `ContractStore.Load` / `SpecLoader.Load` → `ErrFileNotFound` → adapter map |
| 4 PARSE_ERROR | exit 3 + `errors[0].code=PARSE_ERROR` | `ContractStore.Load` / `SpecLoader.Load` → `ErrParse` → adapter map |
| 5 HTTP_ERROR | exit 3 + `errors[0].code=HTTP_ERROR` | `SpecLoader.Load` (HTTP) → `ErrHTTP` → adapter map |
| 6 TIMEOUT_ERROR | exit 3 + `errors[0].code=TIMEOUT_ERROR` | `SpecLoader.Load` (HTTP timeout) → `ErrTimeout` → adapter map |

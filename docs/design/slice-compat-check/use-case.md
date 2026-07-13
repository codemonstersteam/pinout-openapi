# UC-01: Validate consumer↔provider OpenAPI compatibility

- **Primary actor**: CI pipeline (secondary actor: Developer running the same CLI locally).
- **Scope**: `pinout-openapi` compatibility-check CLI (bounded context: Compatibility comparison —
  consumer expectations checked against provider master/prod OpenAPI spec).
- **Level**: user-goal
- **Stakeholders & interests**:
  - **CI pipeline** — needs a deterministic pass/fail before merge: a `compatible | incompatible`
    verdict plus an exit code it can gate on, without runtime surprises.
  - **Developer** — needs a structured JSON report enumerating each mismatch (`{rule, location, detail}`)
    to know precisely what to fix in the consumer or provider spec.
  - **Provider team** — needs assurance the check reads the master (prod) spec as the source of truth
    and never mutates it (pure comparison: no stubs, no test execution, no SDK generation, no
    self-conformance check).
- **Precondition**: `contract-tests.yaml` is present and parseable; the consumer OpenAPI 3.x spec is
  readable; exactly one of `provider.spec_url` / `provider.spec_path` is configured.
- **Trigger**: one-shot CLI invocation `pinout-openapi <contract-tests.yaml>` (config path given as
  argument, or the default path) — the slice's single external input.
- **Minimal guarantee**: on any failure the process emits a diagnostic carrying one committed
  `error.code` and exits with the code's committed exit status (`1` incompatible · `2` config error ·
  `3` spec-I/O error); no partial or corrupt report is presented as authoritative; the provider master
  spec is never modified.
- **Success guarantee (postcondition)**: a compatibility **verdict** (`compatible | incompatible`) is
  produced; a **JSON report** (canon schema) enumerating per-operation results is emitted to
  `settings.json_report_file` when `save_json_report=true`, **plus** a machine-readable line to stdout;
  the **exit code** reflects the verdict (`0` compatible · `1` incompatible). Determinism holds:
  identical inputs → identical verdict, exit code, and report body (modulo timestamps).

## Main Success Scenario

1. CI pipeline invokes the CLI with the path to `contract-tests.yaml`; the CLI reads and parses the
   config (validating required fields, enums, `timeout ∈ 1..600`, and the `spec_url` XOR `spec_path`
   provider rule).
2. CLI loads and parses the **consumer** OpenAPI 3.x spec via `kin-openapi`, resolving local and
   external `$ref` (`allOf` flattened).
3. CLI verifies each configured operation (`path` + normalized-lowercase `method`) exists in the
   consumer spec.
4. CLI loads and parses the **provider** OpenAPI 3.x spec via `kin-openapi` — HTTP(S) GET `spec_url`
   **or** file `spec_path`, unauthenticated, bounded by `settings.timeout`; local and external `$ref`
   resolved, `allOf` flattened.
5. For each configured operation, CLI applies the 5 committed compatibility rules — presence,
   request-contravariance, response-covariance, status-codes, content-types — recording each violation
   as a `finding {rule, location, detail}`.
6. CLI aggregates the per-operation results into one verdict (any incompatible operation → the whole
   result is `incompatible`).
7. CLI emits the JSON report (to `json_report_file` when `save_json_report=true`, plus a machine-readable
   line to stdout) and exits with the verdict's code (`0` compatible).

## Extensions

- **1a. Config missing / unreadable / invalid YAML / missing required field / bad enum / `spec_url`+`spec_path` both-or-neither / empty `operations`**: CLI rejects the config before any spec I/O → `error.code = CONFIG_INVALID`, exit `2`.
- **2a. Consumer spec missing or unreadable** at `consumer.spec_path`: CLI cannot open the expectations side → `error.code = SPEC_UNREADABLE`, exit `2`.
- **2b. Consumer spec not valid OpenAPI 3.x / unparseable**: `kin-openapi` fails to parse the consumer document → `error.code = SPEC_PARSE_ERROR`, exit `2`.
- **3a. Configured operation absent from the consumer spec**: the config references an operation the consumer does not expose (misconfiguration, not an incompatibility) → `error.code = OP_NOT_IN_CONSUMER`, exit `2`.
- **4a. Provider spec unreachable / HTTP non-2xx**: the provider `spec_url` fetch fails or returns a non-2xx status → `error.code = PROVIDER_UNREACHABLE`, exit `3`.
- **4b. Provider fetch exceeds `settings.timeout`**: the bounded HTTP fetch expires → `error.code = PROVIDER_TIMEOUT`, exit `3`.
- **4c. Provider spec not valid OpenAPI 3.x / unparseable**: `kin-openapi` fails to parse the fetched/read provider document → `error.code = PROVIDER_PARSE_ERROR`, exit `3`.
- **5a. ≥1 operation fails a compatibility rule**: at least one operation is `incompatible` with its findings recorded; the report and verdict are produced normally → `verdict = incompatible`, `error.code = INCOMPATIBLE`, exit `1`.
- **7a. JSON report file cannot be written** (bad path / permission / disk full): the verdict was computed but cannot be persisted to `json_report_file` → `error.code = REPORT_WRITE_ERROR`, exit `3`.

## Technology & data variations (optional)

- **4′. Provider spec source** (step 4): resolved from **exactly one** of HTTP(S) GET `provider.spec_url`
  **or** file `provider.spec_path` — mutually exclusive, exactly one present (both/neither is caught at
  step 1a, `CONFIG_INVALID`). The HTTP variant is unauthenticated (public fetch); private-git auth is a
  deferred non-goal.
- **7′. Report emission** (step 7): the JSON report is written to file only when `save_json_report=true`;
  the machine-readable stdout line is always emitted regardless of the file toggle.

> **Traceability**: 9 consumer-visible Extensions == 9 failure-mode-map rows (BRD rows 1–9) == 9
> `error.code` values — `CONFIG_INVALID`, `SPEC_UNREADABLE`, `SPEC_PARSE_ERROR`, `OP_NOT_IN_CONSUMER`,
> `PROVIDER_UNREACHABLE`, `PROVIDER_TIMEOUT`, `PROVIDER_PARSE_ERROR`, `INCOMPATIBLE`, `REPORT_WRITE_ERROR`.
> Row 0 (all operations compatible) is the Main Success Scenario, not an Extension. Each Extension's
> wording is reused verbatim as the matching component-test scenario name (stage 5).

<!-- DONE: usecase slice-compat-check -->

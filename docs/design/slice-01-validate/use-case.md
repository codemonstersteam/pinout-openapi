# UC-1: Validate forward compatibility of a consumer↔provider pair

> Fully-dressed expansion of FRD § UC-1 for slice **slice-01-validate**
> (package `internal/validate/`). Behavioral source of truth; the wording of each Extension is
> reused verbatim as the matching component-test scenario name. The cross-artifact traceability key
> is the **`error.code`** (stable FRD ↔ use-case ↔ contract), not the Cockburn `a/b/c` label.

- **Primary actor**: Consumer CI pipeline (pre-merge). Secondary: consumer developer (local run).
- **Scope**: `pinout-openapi` — Forward Compatibility Check (single bounded context).
- **Level**: user-goal (sea level).
- **Stakeholders & interests**:
  - **Consumer** — must not silently break the integration; wants a deterministic pre-merge verdict.
  - **Provider** — wants to extend responses freely without breaking honest consumers.
  - **pinout-netlist / E2** — needs the structured JSON report (`schema_version="1.0"`) to aggregate.
  - **CI pipeline** — needs a machine-readable exit code to gate the merge.
- **Precondition**: the config file exists and is schema-valid; the consumer's typed
  `consumed-contract` is reachable at `consumer.consumed_contract_path`; the provider spec is
  reachable via **exactly one** of `provider.spec_path` / `provider.spec_url`.
- **Trigger**: `pinout-openapi validate <config.yaml>` (one-shot CLI invocation — the slice's single
  external input).
- **Minimal guarantee**: on any outcome the tool terminates deterministically (same input ⇒ same
  verdict and same report bytes), emits a machine-readable exit code, writes logs to stderr, and
  never partially mutates or corrupts an existing report; the provider token (if any) is read only
  from env `PINOUT_PROVIDER_TOKEN`, never from config or git.
- **Success guarantee (postcondition)**: for each operation in `consumer.operations` a verdict is
  rendered by the four core rules over **body + parameters** (path/query/header); the aggregate
  verdict is `compatible ⇔ errors == []`; when `settings.save_json_report` a schema-valid JSON report
  (top-level `provenance` echo, `errors[]`, informational `uncovered_operations[]`) is written; the
  exit code reflects the verdict (0 compatible / 1 incompatible / 2 config / 3 io·parse).

## Main Success Scenario

1. The **CI pipeline** invokes `pinout-openapi validate <config.yaml>`; the system reads the config
   and validates it against the config schema (FRD Data dictionary A).
2. The system loads the consumer's `consumed-contract` — per-operation `{sends, reads}` with
   `provenance` — already typed (types arrive from the E-harness; the system does **not** infer them).
3. The system loads and parses the provider spec via `kin-openapi` from the one configured source
   (`spec_path` or `spec_url`), resolving `$ref`.
4. For each operation in `consumer.operations`, the system navigates `paths[path][method]` of the
   provider spec and derives `{requires, provides}` over **body + parameters** (path/query/header).
5. The system applies the fixed comparison core (`sandbox/ALGORITHM.md`) per operation:
   **R1** operation exists · **R2 (contravariant)** `requires(provider) ⊆ sends(consumer)` ·
   **R3 (covariant)** `reads(consumer) ⊆ provides(provider)` · **R4** types of shared fields match.
6. The system folds all violations across operations into `{ compatible, errors[] }`
   (`compatible ⇔ errors == []`).
7. The system prints the verdict to stdout; when `settings.save_json_report`, writes the JSON report
   (Data dictionary C) for pinout-netlist/E2; returns exit **0** (compatible). *(Failure-mode F0.)*

## Extensions

*Each Extension = one failure-mode-map row = one `error.code`. Counts: 9 Extensions == 9 rows
(F1–F9) == 9 error codes. F0 is the MSS success end (step 7); the uncovered-surface case is
informational, not a failure row (see Technology & data variations).*

- **1a. Config not found, unreadable, malformed YAML, schema-invalid, or `spec_path`/`spec_url` not
  exactly-one** (step 1): the system rejects the config before any comparison →
  `error.code = CONFIG_ERROR`, exit **2** *(F5; FRD Extensions 1a/1b/1c)*.
- **2a. `consumed-contract` or provider spec file missing / unreadable at its path** (steps 2–3):
  the system reports the missing artifact → `error.code = FILE_NOT_FOUND`, exit **3**
  *(F7; FRD Extensions 2a/3a)*.
- **2b. `consumed-contract` or provider spec unparseable (invalid OpenAPI / YAML)** (steps 2–3):
  the system reports the parse failure → `error.code = PARSE_ERROR`, exit **3**
  *(F6; FRD Extensions 2a/3b)*.
- **3a. Provider `spec_url` unreachable (HTTP failure / non-2xx / connection refused)** (step 3):
  the system reports the fetch failure → `error.code = HTTP_ERROR`, exit **3** *(F8; FRD Extension 3a)*.
- **3b. Provider `spec_url` fetch exceeds `settings.timeout` seconds** (step 3): the system aborts the
  fetch → `error.code = TIMEOUT_ERROR`, exit **3** *(F9; FRD Extension 3a)*.
- **5a. An operation from `consumer.operations` is absent in the provider spec** (step 5, R1): the
  system records the violation → `error.code = OP_NOT_IN_PROVIDER`, verdict `incompatible`, exit **1**
  *(F1; FRD Extension 4a)*.
- **5b. The provider mandates a request field/parameter the consumer does not send** (step 5, R2 over
  body + path/query/header): the system records the violation →
  `error.code = MISSING_REQUIRED_REQUEST_FIELD`, verdict `incompatible`, exit **1**
  *(F2; FRD Extension 5a)*.
- **5c. The consumer reads a field the provider does not offer** (step 5, R3 — catches field removal
  on an open schema): the system records the violation → `error.code = READS_FIELD_NOT_PROVIDED`,
  verdict `incompatible`, exit **1** *(F3; FRD Extension 5b)*.
- **5d. A shared field/parameter's type does not match between consumer and provider** (step 5, R4):
  the system records the violation → `error.code = TYPE_MISMATCH`, verdict `incompatible`, exit **1**
  *(F4; FRD Extension 5c)*.

> Extensions **5a–5d** are a *verdict* (`incompatible` is a legitimate answer, exit 1), not a tool
> error. Extensions **1a** (config, exit 2) and **2a/2b/3a/3b** (io·parse of input artifacts, exit 3)
> are tool errors on the input side. All exit-1 violations are folded across operations (step 6)
> before the verdict is rendered — one run can surface several.

## Technology & data variations

- **Uncovered provider surface** (step 6): provider operations outside `consumer.operations` are
  listed in the report as `uncovered_operations[]` — **informational only**, no verdict or exit-code
  effect (FRD F5d; not a failure-mode row, not an Extension).
- **Provider spec source** (step 3): `spec_path` (local file) XOR `spec_url` (HTTP GET). A private
  `spec_url` is fetched with header `Authorization: Bearer $PINOUT_PROVIDER_TOKEN` sourced from env;
  public URLs work without a token.
- **Compared request surface** (step 4): body **and** parameters (path/query/header); enum/format/
  nullable depth is taken as given by `kin-openapi`.

<!-- DONE: usecase slice-01-validate -->

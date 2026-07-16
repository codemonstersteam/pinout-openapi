# pinout-openapi

> Part of the **pinout** contract-testing platform — domain & concept live in [`docs/CONCEPT.md`](docs/CONCEPT.md).

A deterministic, runtime-free CLI that answers pre-merge whether a REST consumer is still forward-compatible with its provider's OpenAPI spec.

It compares a consumer's typed `consumed-contract` against the provider's master OpenAPI spec — field by field, over request body **and** path/query/header parameters — and emits a verdict, a structured JSON report, and a machine-readable exit code. It is the synchronous twin of `pinout-asyncapi`.

## Can / Cannot

**Can:**

- Verdict `compatible | incompatible` for one consumer↔provider pair.
- Compare body **and** path/query/header parameters per operation.
- Apply four forward-compatibility rules (R1–R4) over each operation.
- Load the provider spec from a local file **or** an HTTP(S) URL.
- Emit a schema-valid JSON report for `pinout-netlist`/E2 aggregation.
- Run offline, deterministically: same input bytes ⇒ same report bytes.

**Cannot:**

- Diff two full specs against each other.
- Extract or infer the `consumed-contract` (the E-harness supplies it, typed).
- Raise stubs, tests, or a service at runtime.
- Detect breaking changes over time (that is `pinout-netlist`/E2).
- Check a service's conformance to its own spec.
- Compare against more than one provider per run.

## Stack

| Component | Technology |
|---|---|
| Language / runtime | Go |
| CLI ingress (door) | cobra, one-shot invocation |
| Provider spec parse | `kin-openapi` (`$ref`-resolving) |
| Input contract (config) | YAML file, validated against `config.schema.json` |
| Output contract (report) | JSON, shaped by `report.schema.json` |
| Provider token source | env `PINOUT_PROVIDER_TOKEN` (never config/git) |

## Command

Source of truth: [`api-specification/config.schema.json`](api-specification/config.schema.json) (input DTO) and [`api-specification/report.schema.json`](api-specification/report.schema.json) (output DTO + exit-code grid), both `x-frozen: 2026-07-16`.

| Command | Args | Action |
|---|---|---|
| `validate` | `<config.yaml>` | Compare consumer↔provider, write report to stdout, return exit `0 \| 1 \| 2 \| 3` |

The config selects the provider spec via **exactly one** of `provider.spec_path` (local file) or `provider.spec_url` (HTTP GET); a private URL is fetched with `Authorization: Bearer $PINOUT_PROVIDER_TOKEN`.

## How it works (data-flow pipe — where it works and where it breaks)

```text
pinout-openapi validate <config.yaml>
| Read + schema-validate config (YAML)                        [CONFIG_ERROR → 2]
| Load consumer consumed-contract (typed {sends, reads})     [FILE_NOT_FOUND, PARSE_ERROR → 3]
| Acquire + parse provider spec (kin-openapi, path XOR url)  [FILE_NOT_FOUND, PARSE_ERROR, HTTP_ERROR, TIMEOUT_ERROR → 3]
| Derive {requires, provides} per operation; apply R1..R4    [OP_NOT_IN_PROVIDER, MISSING_REQUIRED_REQUEST_FIELD, READS_FIELD_NOT_PROVIDED, TYPE_MISMATCH → 1]
| Fold violations → Report (compatible ⇔ errors == [])
| Print JSON report to stdout (+ file iff save_json_report)  → exit 0 | 1 | 2 | 3
```

The four rules: **R1** operation exists · **R2** `requires(provider) ⊆ sends(consumer)` (contravariant request) · **R3** `reads(consumer) ⊆ provides(provider)` (covariant response, catches field removal) · **R4** shared-field types match. Provider operations outside `consumer.operations` are listed as `uncovered_operations[]` — informational only, no verdict or exit effect.

## Failure map (exit codes & error model)

Source: `report.schema.json` `x-exit-codes` + [`use-case.md`](docs/design/slice-01-validate/use-case.md) Extensions (1 Extension = 1 `error.code`). Exit **1 is a verdict** (the domain honestly said "no"), not a tool error; exits **2/3** are tool errors on the input side.

| Exit | Class | Meaning | Error codes |
|---|---|---|---|
| 0 | compatible | contracts compatible (`compatible == true`, `errors == []`) | — |
| 1 | incompatible | a verdict: contract broken | `OP_NOT_IN_PROVIDER`, `MISSING_REQUIRED_REQUEST_FIELD`, `READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH` |
| 2 | config | bad invocation / unreadable / schema-invalid config, or provider source not exactly-one | `CONFIG_ERROR` |
| 3 | io·parse | io or parse failure of an input artifact (consumed-contract or provider spec) | `PARSE_ERROR`, `FILE_NOT_FOUND`, `HTTP_ERROR`, `TIMEOUT_ERROR` |

Each `errors[]` element has the shape `{ code, message, location, details, context }` — `location` is `METHOD PATH` plus the field/parameter. `CONFIG_ERROR` is detected before a report is written, so it drives exit 2 but never appears in `errors[].code`. Rule: anything unchecked or degraded is **visible** in the report (`compatible == false` with an `errors[]` entry, or a non-zero exit) — never masked as success.

## Build & run

```bash
# Build the binary
go build -o pinout-openapi ./cmd/pinout-openapi

# Run the check (exit code is the machine verdict; JSON report on stdout)
./pinout-openapi validate ./config.yaml
echo "exit: $?"

# Private provider spec_url — token from env only
PINOUT_PROVIDER_TOKEN=… ./pinout-openapi validate ./config.yaml
```

Behaviour from outside the binary is proven by [`component-tests/`](component-tests/) (6 black-box scenarios: 1 happy + one per io/config `error.code`).

## Learn more (retrievability ladder)

Read in this order — each level adds context:

1. **This README** — what it is, how to run it.
2. [`component-tests/`](component-tests/) — how it behaves from outside (black-box scenarios).
3. `docs/design/slice-01-validate/` — how the one slice is designed:
   - [`use-case.md`](docs/design/slice-01-validate/use-case.md) — the fully-dressed Cockburn use case (MSS + 9 Extensions).
   - [`module-tree.md`](docs/design/slice-01-validate/module-tree.md) — module tree + head-pipe pseudocode.
   - [`contracts.md`](docs/design/slice-01-validate/contracts.md) — module contracts + component-scenario set.
   - [`c4.md`](docs/design/slice-01-validate/c4.md) — C4 architecture (C2 container + C3 component tree).
   - [`CONTEXT.md`](docs/design/slice-01-validate/CONTEXT.md) — ubiquitous language.
4. [`docs/design/slice-01-validate/adr/`](docs/design/slice-01-validate/adr/) — why it was built this way (SpecLoader unification, verdict-is-unit, io-none loaders).

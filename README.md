# pinout-openapi

`pinout-openapi` is a Go CLI that contract-checks a consumer's expected operations against a
provider's master (= prod) OpenAPI spec. It is a **pure comparison** tool: given a consumer OpenAPI
document and the set of operations (`path` + `method`) the consumer expects to use, it checks each
one against the provider's OpenAPI spec and reports a `compatible` / `incompatible` verdict plus a
structured JSON report of every mismatch. It does **not** stand up stubs, run tests, generate an SDK,
or check that a service conforms to its own spec — it only compares two already-existing specs so a
CI pipeline can gate a merge on the result deterministically.

Part of the [pinout](../pinout/README.md) family; symmetric to `pinout-asyncapi` in config shape and
report format.

## Usage

```sh
pinout-openapi run <contract-tests.yaml>
```

The single positional argument is the path to a config file, shape-validated against
[`api-specification/config.schema.json`](api-specification/config.schema.json):

- **`consumer`** — the expectations side:
  - `spec_path` — filesystem path to the consumer's local OpenAPI 3.x document.
  - `name` — free-form consumer label (non-empty).
  - `operations[]` — operations the consumer expects to be compatible; each item is
    `{path, method}` (`path` starts with `/`; `method` is one of `get|put|post|delete|options|
    head|patch|trace`, case-insensitive on input, normalized to lower-case). At least one item
    required.
- **`provider`** — the source-of-truth side:
  - exactly one of `spec_url` (HTTP(S) URL, unauthenticated fetch) **or** `spec_path` (local file)
    — mutually exclusive, exactly one required.
  - `name` — free-form provider label (non-empty).
- **`settings`** — optional, every field has a default:
  - `log_level` — one of `debug|info|warn|error`, default `info`.
  - `save_json_report` — boolean, default `true`; when `true`, the report is also written to
    `json_report_file`.
  - `json_report_file` — writable path for the JSON report, default `compatibility_report.json`;
    required when `save_json_report` is `true`.
  - `timeout` — provider fetch timeout in seconds, integer in `[1, 600]`, default `30`.

Example `contract-tests.yaml`:

```yaml
consumer:
  spec_path: "./openapi/consumer.yaml"
  name: "mq-rest-sync-adapter"
  operations:
    - path: "/balance"
      method: GET
provider:
  spec_url: "https://git.example.com/wallet-balance-service/openapi.yml"
  name: "wallet-balance-service"
settings:
  log_level: info
  save_json_report: true
  json_report_file: "compatibility_report.json"
  timeout: 30
```

**Output** — a JSON report shaped by
[`api-specification/report.schema.json`](api-specification/report.schema.json):
`verdict` (`compatible|incompatible`), `consumer.name`, `provider.{name,source}`, and
`operations[]` — one entry per configured operation, each carrying `{path, method, status,
findings[]}` (`findings[]` is empty when `status` is `compatible`, and each finding is
`{rule, location, detail}`). The report is always written to stdout as one machine-readable JSON
line; it is additionally written to `settings.json_report_file` when `save_json_report=true`.

## Build & run

```sh
go build ./... && go test ./...            # build the binary + run unit tests
go run ./cmd/app run <config>              # run the compatibility check against a config
./component-tests/scripts/run-tests.sh     # component tests (Docker) — not `go test` from the host
```

## Карта режимов отказа

One row per outcome, matching
[`api-specification/exit-codes.md`](api-specification/exit-codes.md) 1:1 — row 0 (success) plus the
9 committed `error.code` values (9 use-case Extensions == 9 distinguishable failures):

| # | Condition | `error.code` | exit | verdict | operator action |
|---|---|---|---|---|---|
| 0 | all configured operations compatible | — | `0` | compatible | none |
| 1 | ≥1 operation fails a compatibility rule | `INCOMPATIBLE` | `1` | incompatible | fix consumer/provider before merge (see `findings[]`) |
| 2 | config missing/invalid YAML/missing field/bad enum/`spec_url`+`spec_path` both-or-neither/empty `operations` | `CONFIG_INVALID` | `2` | error | fix `contract-tests.yaml` |
| 3 | consumer spec missing/unreadable | `SPEC_UNREADABLE` | `2` | error | fix consumer `spec_path` |
| 4 | consumer spec unparseable / not OpenAPI 3.x | `SPEC_PARSE_ERROR` | `2` | error | fix consumer spec |
| 5 | configured op absent from consumer spec | `OP_NOT_IN_CONSUMER` | `2` | error | fix config or consumer spec |
| 6 | provider spec unreachable / HTTP non-2xx | `PROVIDER_UNREACHABLE` | `3` | error | check `spec_url` / network |
| 7 | provider fetch exceeds `settings.timeout` | `PROVIDER_TIMEOUT` | `3` | error | check network / raise `timeout` |
| 8 | provider spec unparseable / not OpenAPI 3.x | `PROVIDER_PARSE_ERROR` | `3` | error | fix provider spec |
| 9 | JSON report file cannot be written | `REPORT_WRITE_ERROR` | `3` | error | check path / permissions / disk |

**Streams:** stdout carries only the machine report (one JSON body, always emitted); stderr carries
logs/diagnostics, including the `error.code` diagnostic line on any non-zero exit.

# pinout-openapi — exit-code table (CLI contract, frozen)

> Third artifact of the **cli** target profile's frozen contract, beside `config.schema.json`
> (input DTO) and `report.schema.json` (output canon). The exit code is to this CLI what the HTTP
> status is to a service: the machine-readable serialization of the outcome (Design by Contract —
> the breach channel). **Frozen: 2026-07-12.** Derived 1:1 from the BRD failure-mode map and the
> fully-dressed use case's 9 Extensions.

## Exit convention (committed)

| exit | class | meaning |
|---|---|---|
| `0` | Ok | compatible — core succeeded, report on stdout |
| `1` | domain failure | incompatible — input valid, the domain said "no" |
| `2` | config / usage | bad invocation, unreadable/invalid config, malformed input DTO |
| `3` | environment / spec-I/O | external failure not caused by the input (network, unreachable/unparseable provider, report write) |

## Full mapping — 9 `error.code` + row 0 (success)

Every row is one distinguishable failure; the `error.code` strings and exit codes are committed,
traceable constraints. 9 `error.code` values == 9 use-case Extensions == this table's rows 1–9.

| # | Condition (UC Extension) | `error.code` | exit | verdict | operator action |
|---|---|---|---|---|---|
| 0 | all configured operations compatible (MSS) | — | `0` | compatible | none |
| 1 | ≥1 operation fails a compatibility rule (5a) | `INCOMPATIBLE` | `1` | incompatible | fix consumer/provider before merge (see `findings[]`) |
| 2 | config missing/invalid YAML/missing field/bad enum/`spec_url`+`spec_path` both-or-neither/empty `operations` (1a) | `CONFIG_INVALID` | `2` | error | fix `contract-tests.yaml` |
| 3 | consumer spec missing/unreadable (2a) | `SPEC_UNREADABLE` | `2` | error | fix consumer `spec_path` |
| 4 | consumer spec unparseable / not OpenAPI 3.x (2b) | `SPEC_PARSE_ERROR` | `2` | error | fix consumer spec |
| 5 | configured op absent from consumer spec (3a) | `OP_NOT_IN_CONSUMER` | `2` | error | fix config or consumer spec |
| 6 | provider spec unreachable / HTTP non-2xx (4a) | `PROVIDER_UNREACHABLE` | `3` | error | check `spec_url` / network |
| 7 | provider fetch exceeds `settings.timeout` (4b) | `PROVIDER_TIMEOUT` | `3` | error | check network / raise `timeout` |
| 8 | provider spec unparseable / not OpenAPI 3.x (4c) | `PROVIDER_PARSE_ERROR` | `3` | error | fix provider spec |
| 9 | JSON report file cannot be written (7a) | `REPORT_WRITE_ERROR` | `3` | error | check path / permissions / disk |

## Streams

- **stdout** — only the machine report (`report.schema.json` shape), one JSON body. Always emitted.
- **stderr** — logs / diagnostics / the `error.code` diagnostic line on any non-zero exit.

## Notes on the two exit-2 spec errors

`SPEC_UNREADABLE` / `SPEC_PARSE_ERROR` (rows 3–4) are exit `2` because they concern the **consumer**
side (the input the caller supplies). The symmetric **provider**-side failures (rows 6–8) are exit `3`
because the provider master spec is an **external** resource, not the caller's input. A provider file
that is missing/unreadable surfaces as `SPEC_UNREADABLE` (exit 2) per the data dictionary, while a
provider file that parses-but-is-not-valid-OpenAPI surfaces as `PROVIDER_PARSE_ERROR` (exit 3).

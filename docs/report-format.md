# Report format canon — `docs/report-format.md`

**Version: `1.1` · Canon owner: `pinout-openapi` (epic E1) · superseded: docs/report-format.md@2e0c232^**

The revision at `docs/report-format.md@2e0c232^` (deleted by commit `2e0c232`, "clean slate for
harness run") described a `verdicts[]{subject, compatible, errors[]}` + `provider{}` shape without
`provenance`. That `1.0` document was never implemented by any validator — `pinout-openapi` and
`pinout-asyncapi` both froze a different, flat `1.0` shape instead. This canon does not restore the
deleted revision; it is written fresh from the shape both validators actually froze, versioned
forward to `1.1`.

## Purpose and audience

This is the single, ecosystem-wide written source of truth for the JSON report every `pinout`
validator emits on `stdout` (and optionally to a configured file). It is written so that
`pinout-asyncapi` and `pinout-netlist` can align their own contracts against this document **alone**,
without reading `internal/` of this repository. The canon describes the **target** form, `1.1`. This
repository's own `api-specification/report.schema.json` still advertises `1.0` until a follow-up
ticket (`debt/02-report-schema-1.1.md`, tracked as ticket 2/2) brings the schema file itself to
`1.1` — that drift is expected, not a defect of this document. See
[§ Reference: machine schema](#10-reference-machine-schema) below.

## 1. Full 1.1 schema, field by field

Every field of the report object is listed below. The report object is closed
(`additionalProperties: false`); the `consumer` and `provenance` sub-objects are closed the same way.

| Field | Type | Required | Fit criterion (valid range / enum / format) | Meaning; error on violation |
|---|---|---|---|---|
| `schema_version` | string | required | `const "1.1"` (was `const "1.0"` in the actually-frozen `1.0` shape) | Report-canon semver, so `pinout-netlist` can dispatch its reader by version. Violation ⇒ producer bug, the report itself is schema-invalid. |
| `compatible` | boolean | required | `true` \| `false`; invariant `compatible ⇔ errors == []`; `true ⇒ exit 0`, `false ⇒ exit 1` | Aggregate verdict. A breach of the invariant is a producer defect (see [§ 5 Invariant](#5-invariant)). |
| `validator` | string | **required (new in 1.1)** | enum `pinout-openapi` \| `pinout-asyncapi`; a build-time constant, one value per repository | Which validator produced the report. Violation ⇒ schema-invalid report. |
| `interaction` | string | **required (new in 1.1)** | enum `sync` \| `async`; a build-time constant (`sync` for this repository, `async` for its twin) | Interaction style of the checked pair, so `pinout-netlist` can label the edge without guessing. Violation ⇒ schema-invalid report. |
| `consumer` | object | **required (new in 1.1)** | `additionalProperties: false` — closed, like `provenance` and the report root | Consumer identity. Without it `pinout-netlist` knows the provider (from `provenance`) but not who depends on it. Violation ⇒ schema-invalid report. |
| `consumer.name` | string | **required (new in 1.1)** | non-empty; taken from the consumer's own config | The consumer's name. Violation ⇒ schema-invalid report. |
| `consumer.version` | string | optional (new in 1.1) | emitted **only if known**; when unknown the key is **omitted entirely** — never `null`, never `""` | Consumer version, when the caller has one to give. Absence is legal, not an error. |
| `generated_at` | string | **required (new in 1.1)** | RFC 3339 date-time, UTC with the `Z` designator, second precision: `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, e.g. `2026-07-25T14:03:12Z`; no offset form, no fractional seconds | Report-generation timestamp; `pinout-netlist` uses it as the freshness `at` of the verdict record. Injected by the caller through a clock port — never read from the specs or the system clock inside the pure core. Violation ⇒ schema-invalid report. |
| `provenance` | object | required | `additionalProperties: false`; required `provider`, `provider_version`, `captured_hash` | The **single source of provider identity** — see [§ 6](#6-provenance--the-single-source-of-provider-identity). Violation ⇒ schema-invalid report. |
| `provenance.provider` | string | required | provider name as stamped at capture time | Identity of the provider spec the consumer's contract was captured from. |
| `provenance.provider_version` | string | required | version stamp of the source spec (OpenAPI/AsyncAPI `info.version`, or the commit that produced it) | Lets a reader tell which provider revision was checked against. |
| `provenance.captured_hash` | string | required | `pattern ^sha256:` — sha256 of the captured provider-spec bytes | Lets `pinout-netlist` detect provider drift and re-check on change. |
| `errors` | array | required | array of error objects, **flat** — no `verdicts[]` grouping; empty ⇔ `compatible == true`; no rule short-circuits, so one run may carry several entries | The full set of violations found in one run. See [§ 7](#7-why-there-is-no-verdicts). Violation of the invariant ⇒ producer defect. |
| `errors[].code` | string | required | enum = the validator's own dictionary: sync 4 codes + async 9 codes (R1–R9) + shared io/parse 4 codes; see [§ 4](#4-error-code-dictionaries). `CONFIG_ERROR` is **excluded by construction** — it never appears here. Consumers **must tolerate an unknown code**: an unrecognised value is still an error, exit-code semantics unchanged | Stable traceability key. Violation ⇒ schema-invalid report. |
| `errors[].message` | string | required | human-readable, non-empty | Free-text explanation for a human reader. |
| `errors[].subject` | string | **required for verdict codes (exit 1); optional (omitted) for io/parse codes (exit 3)** — new in 1.1 | sync: `METHOD /path` — uppercase HTTP method + one ASCII space + the path verbatim from the consumer contract (leading `/`, template braces preserved), e.g. `GET /wallets/{id}`; async: `<channel address> <direction> <message key>`, where `direction` ∈ `send` \| `receive` as seen **from the consumer** and the message key is the channel's `messages` map key. It is the prefix of `errors[].location`, computed once, never composed a second time by hand | The graph-edge granularity `pinout-netlist` keys on: one `subject` = one edge. Exit-3 errors have no operation/channel identified, so the key is omitted there. Violation ⇒ schema-invalid report. |
| `errors[].location` | string | optional | sync: `METHOD PATH` plus the field/parameter when applicable; async: `subject` plus a dotted field path, e.g. `WALLET.BALANCE.RESPONSE receive getBalanceResponse payload.data.balance` | The precise coordinate of the violation within the subject. |
| `errors[].details` | string | optional | free-form detail (expected vs. actual type, field name, protocol pair) | Extra human-readable detail. |
| `errors[].context` | object | optional | free-form structured diagnostics, e.g. `{"consumer_protocol":"kafka","provider_protocol":"amqp"}` | Machine-readable extra diagnostics. |
| `uncovered_operations` | array of string | optional, **sync only** | provider operations outside `consumer.operations` | Informational; no effect on `compatible` or the exit code. |
| `uncovered_channels` | array of string | optional, **async only** | provider channel addresses outside `consumer.channels` | Informational; async counterpart of `uncovered_operations`. No effect on `compatible` or the exit code. |

The only two allowed differences between the two validators' reports are the `errors[].code` enum
and `uncovered_operations[]` vs. `uncovered_channels[]` — see [§ 9](#9-per-validator-divergence).
Everything else is byte-identical, so `pinout-netlist` reads both report shapes with one reader.

## 2. Delta table: 1.0 → 1.1

The actually-frozen `1.0` shape was `{schema_version, compatible, provenance{provider,
provider_version, captured_hash}, errors[]{code, message, location, details, context},
uncovered_operations[] | uncovered_channels[]}`. `1.1` is that shape plus the seven changes below.
This table restates the same facts as § 1 above, as a delta — the two must agree.

| Field | Status | Value |
|---|---|---|
| `schema_version` | was `"1.0"` → becomes `"1.1"` | `const "1.1"` |
| `validator` | new, required | `pinout-openapi` \| `pinout-asyncapi` |
| `interaction` | new, required | `sync` \| `async` |
| `consumer.name` | new, required | the consumer's name, from its own config |
| `consumer.version` | new, optional | emitted only if known |
| `generated_at` | new, required, RFC 3339 | UTC, `Z`, second precision — timestamped by the caller |
| `errors[].subject` | new, required for verdict codes (exit 1); optional for io/parse codes (exit 3) | sync: `METHOD /path`; async: `<channel address> <direction> <message key>` |

## 3. Exit codes (`x-exit-codes`)

The process **exit code** is the CLI's status line — it is carried on the process, **not** a field of
the report document. It is identical across both validators.

| Exit | Class | Meaning | `errors[].code` values that produce it |
|---|---|---|---|
| `0` | compatible | `compatible == true`, `errors == []` | — (none) |
| `1` | incompatible | a verdict: the domain honestly said the contracts don't match | the 4 sync or 9 async verdict codes, see § 4 |
| `2` | config | bad invocation / unreadable / schema-invalid config, or provider-spec source not exactly one | `CONFIG_ERROR` — never appears in `errors[]` |
| `3` | io-parse | environment/spec-artifact failure (consumed-contract or provider spec) | `PARSE_ERROR`, `FILE_NOT_FOUND`, `HTTP_ERROR`, `TIMEOUT_ERROR` |

## 4. Error-code dictionaries

### 4.1 Sync codes (`pinout-openapi`, 4 codes, exit `1`)

| Code | Rule | Meaning |
|---|---|---|
| `OP_NOT_IN_PROVIDER` | R1 | The consumer expects an operation the provider does not expose. |
| `MISSING_REQUIRED_REQUEST_FIELD` | R2 | The consumer does not send a field the provider requires on the request. |
| `READS_FIELD_NOT_PROVIDED` | R3 | The consumer reads a response field the provider does not provide. |
| `TYPE_MISMATCH` | R4 | Consumer and provider disagree on a field's type. |

### 4.2 Async codes (`pinout-asyncapi`, 9 codes, R1–R9, exit `1`)

| Code | Rule | Meaning |
|---|---|---|
| `CHANNEL_NOT_IN_PROVIDER` | R1 | The consumer expects a channel address the provider does not expose. |
| `PROTOCOL_MISMATCH` | R2 | Consumer and provider disagree on the channel's protocol binding. |
| `DIRECTION_NOT_IN_PROVIDER` | R3 | The consumer's send/receive direction on a channel is not offered by the provider. |
| `MESSAGE_NOT_IN_PROVIDER` | R4 | The consumer expects a message key the provider's channel does not offer. |
| `MISSING_REQUIRED_SENT_FIELD` | R5 | The consumer does not send a payload field the provider requires. |
| `READS_FIELD_NOT_PROVIDED` | R6 | The consumer reads a payload field the provider does not provide. |
| `TYPE_MISMATCH` | R7 | Consumer and provider disagree on a payload field's type. |
| `CONTENT_TYPE_MISMATCH` | R8 | Consumer and provider disagree on the message's content type. |
| `CORRELATION_ID_MISMATCH` | R9 | Consumer and provider disagree on the message's correlation-id binding. |

### 4.3 Shared io/parse codes (both validators, 4 codes, exit `3`)

| Code | Meaning |
|---|---|
| `PARSE_ERROR` | An input artifact (consumed contract or provider spec) is malformed. |
| `FILE_NOT_FOUND` | An input artifact path does not resolve. |
| `HTTP_ERROR` | Fetching a remote provider spec failed at the HTTP layer. |
| `TIMEOUT_ERROR` | Fetching a remote provider spec exceeded its time budget. |

### 4.4 `CONFIG_ERROR` — excluded by construction

`CONFIG_ERROR` (exit `2`) is a config-side breach — bad invocation, unreadable or schema-invalid
config, or a provider-spec source that is not exactly one — detected **before** a report is ever
written. It therefore **never appears in `errors[].code`**; it is not one of the sync, async, or
shared io/parse codes above, and no report on disk or stdout will ever carry it.

## 5. Invariant

**`compatible ⇔ errors == []`.**

`compatible == true` if and only if `errors` is the empty array; `compatible == true ⇒` exit `0`,
`compatible == false ⇒` exit `1`. A report where this does not hold is a producer defect — there is
no runtime code path in a conforming validator that can emit one.

## 6. `provenance` — the single source of provider identity

`provenance` (`provider`, `provider_version`, `captured_hash`) is declared the **single source of
provider identity** in `1.1`. There is deliberately no separate `provider{}` block — the shape from
the superseded `2e0c232^` revision is gone. `provenance` is strictly stronger than a bare `provider{}`
would be: it also carries `captured_hash`, so `pinout-netlist` can detect that the provider spec
changed since capture and re-check, which a plain name/version pair cannot do. A reader that still
looks for `provider{}` is looking for the dead shape; `provenance` is where provider identity lives.

## 7. Why there is no `verdicts[]`

Subject granularity — "which operation/channel is this violation about" — is carried by
`errors[].subject` (§ 1), not by grouping `errors[]` into a `verdicts[]{subject, compatible,
errors[]}` structure. `pinout-netlist` builds one graph edge per `subject`, reading the flat `errors[]`
array directly; it does not need a second, regrouped shape to do so. Regrouping into `verdicts[]` was
considered and **rejected as variant B** (the parent decision, `pinout/debt/report-canon-fork.md`,
chose variant A: the additive flat form documented here).

## 8. Versioning policy

- A **new field** → **minor** version bump.
- **Adding an enum value** to `errors[].code` → **minor** version bump. Consumers of this canon are
  obliged to tolerate an unknown `errors[].code`: an unrecognised code is still an error entry, and
  exit-code semantics are unchanged by it.
- **Regrouping or removing** a field (e.g. reintroducing `verdicts[]`, or dropping `provenance`) →
  **major** version bump.

## 9. Per-validator divergence

`pinout-openapi` (sync) and `pinout-asyncapi` (async) reports are byte-identical in shape except for
exactly two differences:

1. the `errors[].code` enum — 4 sync codes (§ 4.1) vs. 9 async codes (§ 4.2);
2. `uncovered_operations[]` (sync) vs. `uncovered_channels[]` (async).

Everything else in § 1 — every field name, type, requiredness, and fit criterion — is the same for
both validators, so `pinout-netlist` reads both with one reader keyed on `validator` and
`interaction`.

## 10. Reference: machine schema

The machine-readable JSON Schema for this repository's report is
[`api-specification/report.schema.json`](../api-specification/report.schema.json).

Drift note: that schema file still advertises `schema_version: const "1.0"` today; it reaches 1.1 in ticket 2/2
(`debt/02-report-schema-1.1.md`), a separate, later run. Until then this canon document — not the
schema file — is the target-form source of truth for `1.1`.

## Glossary

- **canon** — this document: the single written source of truth for the report format across the
  `pinout` ecosystem, owned by `pinout-openapi` (epic E1).
- **twin** — the mirror validator contract; `pinout-asyncapi`'s report schema is the twin of this
  repository's.
- **subject** — the granularity unit `pinout-netlist` turns into one graph edge (§ 1, `errors[].subject`).
- **provenance** — the stamp of which provider spec the consumer's contract was captured from; the
  sole source of provider identity in `1.1` (§ 6).
- **superseded `1.0`** — `docs/report-format.md@2e0c232^`, a spec no validator ever implemented (see
  the header note above).

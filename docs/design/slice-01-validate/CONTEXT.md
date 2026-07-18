# Forward Compatibility Check (pinout-openapi)

The single bounded context of `pinout-openapi`: given a consumer's typed expectations and a provider's
master OpenAPI spec, decide whether the provider can still serve the consumer — a deterministic
pre-merge verdict. Slice = this context (`internal/validate/`).

## Language

**Consumer**:
The component under test that depends on a provider's HTTP API; contributes a `consumed-contract` and
the `operations` scope to check.

**Provider**:
The component that exposes the master OpenAPI spec (source of truth); may extend responses freely.

**Consumed-contract**:
The consumer's expectation, reconstructed by the E-harness from its component-test stubs into
per-operation typed `{sends, reads}` + `provenance`. Already typed — the tool never infers types.
_Avoid_: expected schema, consumer schema.

**Sends**:
The request fields/parameters (body + path/query/header) the consumer actually transmits for an
operation.

**Reads**:
The response body fields the consumer actually consumes for an operation.

**Requires**:
The provider's mandatory request fields/parameters for an operation (contravariant target of `sends`).

**Provides**:
The provider's response body fields for an operation (covariant target of `reads`).

**Forward compatibility**:
The property that a provider change does not break an honest consumer — the aggregate verdict
`compatible ⇔ errors == []`.
_Avoid_: backward compatibility, compatibility (unqualified).

**Rules R1–R4**:
The fixed comparison core (ported from `sandbox/ALGORITHM.md`): **R1** operation exists · **R2**
`requires(provider) ⊆ sends(consumer)` (contravariant request) · **R3** `reads(consumer) ⊆
provides(provider)` (covariant response, catches field removal on an open schema) · **R4** shared-field
types match.

**Verdict**:
The `compatible` outcome and its `errors[]` (codes `OP_NOT_IN_PROVIDER`,
`MISSING_REQUIRED_REQUEST_FIELD`, `READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH`, exit 1) — the domain
honestly saying "no", **not** a tool error.
_Avoid_: failure, crash (reserve those for io/parse tool errors at exit 2/3).

**Uncovered operation**:
A provider operation outside `consumer.operations` — informational only (`uncovered_operations[]`), no
effect on verdict or exit code.

**Provenance**:
The stamp echoed from the consumed-contract into the report (`provider`, `provider_version`,
`captured_hash`).

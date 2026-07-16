# Forward Compatibility Check

The language of `pinout-openapi`: deciding, before merge and without runtime, whether a REST
consumer is still structurally compatible with the provider's master OpenAPI spec — in the
**forward** direction (consumer → provider). Provider spec is the source of truth.

## Language

**Provider spec**:
The provider's master OpenAPI document, treated as the single source of truth (equals production).
Loaded from a file or an HTTP URL and parsed with `kin-openapi`.
_Avoid_: schema, contract (when referring to the provider), master

**Consumed-contract**:
The consumer's reconstructed expectation of the provider, per operation — a typed, schema-shaped
`{sends, reads}` fragment plus a `provenance` stamp. Produced upstream by the E-harness (the tool
does **not** infer types from examples). This is the consumer's side of the comparison.
_Avoid_: expected spec, consumer spec, stub contract

**Operation**:
One `(path, method)` pair the consumer calls, named in the config; the unit of comparison. Comparable
only if it exists in the provider spec.
_Avoid_: endpoint, route, call

**Sends**:
The set of request fields **and** parameters (path/query/header + body) the consumer transmits, each
already carrying its type in the consumed-contract.
_Avoid_: input, request payload

**Reads**:
The set of response fields the consumer consumes from the provider's answer, each already typed.
_Avoid_: output, response payload

**Requires** / **Provides** (of the provider, per operation):
`requires` = request fields/parameters the provider mandates; `provides` = response fields the
provider returns. Both derived by navigating the provider spec's schema (`$ref` resolved).
_Avoid_: mandatory fields, available fields

**Compatible**:
The verdict for the pair: every checked operation satisfies all four core rules — R1 operation exists,
R2 `requires(provider) ⊆ sends(consumer)` (contravariant), R3 `reads(consumer) ⊆ provides(provider)`
(covariant), R4 types of shared fields match. Formally `compatible ⇔ errors == []`.
_Avoid_: passing, valid, ok

**Incompatible**:
A verdict, **not** an error: at least one core rule is violated. Distinct from an I/O/parse/config
failure, which is a tool error, not a verdict.
_Avoid_: broken, failed, error

**Uncovered operation**:
A provider operation not listed in the consumer's config — reported **informationally** only; never
affects the verdict or the exit code.
_Avoid_: gap, missing coverage

# CONTEXT — bounded context `compat-check` (Compatibility comparison)

> Ubiquitous language for the slice. Bounded context = slice; knowledge co-located here
> (`docs/design/slice-compat-check/`), code in `internal/compat-check/`, bound by slug `compat-check`.

## Ubiquitous language

| Term | Meaning (in this context) |
|---|---|
| **Consumer** | the expectations side — the OpenAPI spec of the service that *calls* the API (`consumer.spec_path`, local file). |
| **Provider** | the source-of-truth side — the *master (= prod)* OpenAPI spec (`provider.spec_url` HTTP **or** `provider.spec_path` file). Never mutated. |
| **Operation** | one `path` + `method` pair the consumer declares it uses. |
| **Compatible** | the consumer's expectations are satisfiable by the provider: all 5 rules hold for every configured operation. |
| **Incompatible** | at least one operation fails at least one rule; each failure is a `finding`. A **verdict**, not an error. |
| **Finding** | one recorded mismatch: `{rule, location, detail}`; `rule` ∈ underscore-form `{presence, request_contravariance, response_covariance, status_codes, content_types}`. |
| **Verdict** | the overall result `compatible | incompatible` (any incompatible operation ⇒ `incompatible`). |
| **Rule** | one compatibility check. Direction: consumer expectations ⊆ provider capability (contravariant on request, covariant on response). |
| **Report** | the canon JSON output (`report.schema.json`): verdict + per-operation results + findings; written to file (when `save_json_report`) + a machine-readable stdout line. |
| **Config** | the validated `contract-tests.yaml` (`config.schema.json`): consumer, provider, operations, settings. |

## Committed decisions (from BRD — do not re-litigate here)

- **Honest reuse**: `kin-openapi` for OpenAPI 3.x parse + local/external `$ref` resolve; `allOf` flattened.
- **Deferred (out of MVP)**: `oneOf`/`anyOf`, enum narrowing, `format` semantics, private-git auth.
- **Purity**: pure comparison of two specs — no stubs, no test execution, no SDK generation, no self-conformance.
- **`OP_NOT_IN_CONSUMER`**: a configured op absent from the **consumer** spec is a *config error* (exit 2), not an incompatibility.
- **Determinism**: identical inputs → identical verdict, exit code, byte-identical report body (modulo timestamps).

## Boundaries

- Owns: config validation, spec loading, the 5 rules, verdict, report emission.
- Does **not** own: authoring OpenAPI specs, running services, generating SDKs. Adjacent `pinout-asyncapi`
  later aligns its report to this canon (symmetry) — a downstream consumer of the format, not part of this context.
</content>

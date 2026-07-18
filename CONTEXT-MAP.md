# Context Map — pinout-openapi

`pinout-openapi` is a **single bounded context**: **Forward Compatibility Check** — deciding, before
merge and without runtime, whether a REST consumer is still structurally compatible with the provider's
master OpenAPI spec (forward direction, provider spec = source of truth).

## Bounded contexts

| Context | Ubiquitous language | Code |
|---|---|---|
| Forward Compatibility Check | [`CONTEXT.md`](CONTEXT.md) (repo language) · [`docs/design/slice-01-validate/CONTEXT.md`](docs/design/slice-01-validate/CONTEXT.md) (slice language) | `internal/validate/` |

The slice `CONTEXT.md` is the same context as the repo `CONTEXT.md`, narrowed to the `slice-01-validate`
vocabulary (`internal/validate/`). Per the domain-modeling rule *slice = bounded context*, this repo has
exactly one, so there is one row and no cross-context relationships to map.

## Relationships

None — a single context has no inter-context edges. External systems (consumer, provider) are actors,
not bounded contexts of this repo; their contracts are the OpenAPI specs, not shared domain models.

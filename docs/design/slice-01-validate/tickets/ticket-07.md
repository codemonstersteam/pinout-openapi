---
id: 07
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md, api-specification/config.schema.json]
outputs: [internal/validate/config/config.go, internal/validate/config/config_test.go]
io: none
skills: []
---

### TICKET 07 — slice-01-validate/NewConfig: valid-by-construction config constructor

**io:** none → skills: [] (pure constructor — **unit-tested by formula**).

**Context (only this module):**
- contract (`contracts.md` §NewConfig): `NewConfig(raw: RawConfig) -> Result[Config, Error]`.
  Validate every field against `config.schema.json` invariants; illegal states unrepresentable past
  this boundary (unexported fields). Deps = —.
  - antecedent: `RawConfig` decoded.
  - consequent: Ok `Config` (non-empty names; `operations` ≥ 1, each `path` matches `^/` and `method`
    in the enum; provider source **exactly one** of `spec_path`/`spec_url`; settings defaulted +
    in-range: `timeout > 0`, `log_level` enum). Fail `ErrConfig` → `CONFIG_ERROR`, exit 2.
- **unit tests: 9** (`module-tree.md` formula) — 1 happy + 8 antecedent branches:
  empty name (consumer/provider), empty `operations`, bad `path` pattern (not `^/`), bad `method` enum,
  `spec_*` both-set, `spec_*` neither-set, `timeout ≤ 0`, bad `log_level` enum.
- component scenario(s) to green: none (schema-invalidity is UNIT-covered here; shares `CONFIG_ERROR`
  code with the `ConfigStore` read failure — ADR-0002/0003).

**Dependencies:** `RawConfig`/`Config`/`Settings` types + `ErrConfig` (ticket 03).

**Subagent instruction:** write the 9 unit tests → implement `NewConfig` → run units → green → done.
Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/... && go test ./internal/validate/config/...`.

**Acceptance:** `config` builds/vets clean; 9 unit tests green.

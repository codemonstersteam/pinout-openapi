---
id: 06
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, api-specification/config.schema.json]
outputs: [internal/validate/config/store.go]
io: none
skills: []
---

### TICKET 06 — slice-01-validate/ConfigStore.Load: read config file (I/O pipe)

**io:** none → skills: [] (filesystem I/O pipe — ADR-0003; local read routes to no io sub-skill). Pure
pipe, **not unit-tested**.

**Context (only this module):**
- contract (`contracts.md` §ConfigStore.Load): `Load(path: string) -> Result[RawConfig, Error]`.
  Read config file bytes + YAML-unmarshal into an **unvalidated** `RawConfig` (no validation here — that
  is `NewConfig`). Deps = — (OS filesystem encapsulated in the object).
  - antecedent: `path` non-empty.
  - consequent: Ok `RawConfig` (structurally decoded); Fail `ErrConfig` (not found / unreadable /
    malformed YAML) → `CONFIG_ERROR`, exit 2.
- unit tests: **none** (I/O pipe — its failure branch is component scenario 2).
- component scenario(s) to green: scenario 2 (`CONFIG_ERROR`, exit 2) — greened later.

**Dependencies:** `RawConfig` type + `ErrConfig` sentinel (ticket 03).

**Subagent instruction:** implement `ConfigStore.Load` in `internal/validate/config/store.go` →
`go build`/`vet` → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `config` builds and vets clean; no units (I/O pipe).

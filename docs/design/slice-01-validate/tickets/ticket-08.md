---
id: 08
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, .agent/planner/frd.md]
outputs: [internal/validate/contract/store.go]
io: none
skills: []
---

### TICKET 08 — slice-01-validate/ContractStore.Load: read+parse consumed-contract (I/O pipe)

**io:** none → skills: [] (filesystem I/O pipe — ADR-0003). Pure pipe, **not unit-tested**.

**Context (only this module):**
- contract (`contracts.md` §ContractStore.Load): `Load(path: string) -> Result[ConsumedContract, Error]`.
  Read + parse the E-harness `consumed-contract` artifact (already typed per-operation `{sends, reads}`
  + `provenance`; FRD Data-dictionary B). **No type inference** — types arrive typed. Deps = —.
  - antecedent: `path` non-empty.
  - consequent: Ok `ConsumedContract{ Operations: []{ Ref, Sends, Reads }, Provenance }`. Fail
    `ErrFileNotFound` (missing/unreadable) → `FILE_NOT_FOUND`/3; `ErrParse` (unparseable) → `PARSE_ERROR`/3.
- unit tests: **none** (I/O pipe — its two failure branches are component scenarios 3 & 4).
- component scenario(s) to green: 3 (`FILE_NOT_FOUND`) + 4 (`PARSE_ERROR`) — greened later.

**Dependencies:** `ConsumedContract`/`Provenance` types + `ErrFileNotFound`/`ErrParse` (ticket 03).

**Subagent instruction:** implement `ContractStore.Load` in `internal/validate/contract/store.go` →
`go build`/`vet` → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `contract` builds and vets clean; no units (I/O pipe).

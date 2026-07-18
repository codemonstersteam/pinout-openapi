---
id: 15
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/adr/0003-filesystem-loaders-tagged-io-none.md]
outputs: [internal/validate/report/writer.go]
io: none
skills: []
---

### TICKET 15 — slice-01-validate/ReportWriter.Write: write JSON report (I/O pipe)

**io:** none → skills: [] (filesystem write pipe — ADR-0003). Pure pipe, **not unit-tested**.

**Context (only this module):**
- contract (`contracts.md` §ReportWriter.Write): `Write(s: Settings, r: Report) -> Result[Report, Error]`.
  Iff `settings.save_json_report`, atomically write the JSON to `settings.json_report_file`; ROP
  pass-through of `Report` unchanged (never partially mutates an existing report). Deps = — (OS
  filesystem encapsulated).
  - antecedent: schema-valid `Report`.
  - consequent: Ok the same `Report` (stdout copy printed by `main`). A write failure → exit 3, logged
    to stderr — **not** a contract `error.code` (no enum code exists) → **not** a counted component branch.
- unit tests: **none** (I/O pipe).
- component scenario(s) to green: the happy scenario (1) asserts the file **is** written — greened later.

**Dependencies:** `Report`/`Settings` types (ticket 03).

**Subagent instruction:** implement `ReportWriter.Write` in `internal/validate/report/writer.go` →
`go build`/`vet` → done. Touch no other module.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `report` builds and vets clean; no units (I/O pipe).

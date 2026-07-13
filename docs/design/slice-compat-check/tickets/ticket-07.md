---
id: 07
type: module
slice: slice-compat-check
blocked_by: [01, 02, 05, 06]
inputs: [docs/design/slice-compat-check/contracts.md, api-specification/report.schema.json]
outputs: [internal/compat-check/report/writer.go]
io: none
skills: []
---

### TICKET S-compat-check.07 — `report/`: persist the canon report (file + stdout)

**io:** `none` (header value — plain filesystem write, no dedicated remote sub-skill per `module-tree.md`'s
`io: file` note)   →  skills: (none beyond core `program-implementation`)

**Context (only this package — `internal/compat-check/report/`):**
- **`writer.go` — `ReportWriter.Write(report compare.Report, settings config.Settings) -> Result<Outcome, Error>`.**
  Deps: —. Write JSON `report` to `settings.JSONReportFile` **when** `settings.SaveJSONReport=true`,
  **plus** the machine-readable stdout line (**always**, regardless of `save_json_report`). Pure I/O pipe
  — no transformation of the already-built `Report`. Implements the `ReportWriter` port interface declared
  in `internal/compat-check/domain.go` (ticket 06) — satisfied structurally, no import of the package root
  needed for the interface itself, only for the `Outcome` return type and the `ErrReportWrite` sentinel.
  Failure → `ErrReportWrite` → `error.code=REPORT_WRITE_ERROR` (exit 3, CS9) — bad path/permission/disk.
  Success → `Outcome{Verdict}` (carries the verdict through for the door's exit mapping, ticket 08).
  **Never writes on the read path** — CS9's Then "provider master spec not mutated" is satisfied by
  construction (this package has no write access to anything but `json_report_file` + stdout).

**Unit tests:** none — I/O pipe (proven by CS1/CS2's report-shape assertions on the success path and CS9
on the failure path, via the fixer, not this ticket).

**Dependencies:** imports `internal/compat-check/compare` (ticket 05, `Report` type) and
`internal/compat-check` package root (ticket 06, `Outcome` type + `ErrReportWrite` sentinel + the
`ReportWriter` interface it must satisfy).

**Subagent instruction:** implement `writer.go` (file write gated on `save_json_report` + unconditional
stdout line) → run build/vet. No test file to write. Do not touch `config/`, `spec/`, `compare/`, `cli/`,
or the package-root files.

**Verify:** `go build ./internal/compat-check/report/... && go vet ./internal/compat-check/report/...`

**Acceptance:** package builds and vets clean; `writer.go` satisfies the `ReportWriter` interface from
`internal/compat-check/domain.go` (compiler-checked once `register.go` wires it in, ticket 09). No unit
test expected — component scenarios CS1/CS2/CS9 prove it later, via the fixer, not this ticket.

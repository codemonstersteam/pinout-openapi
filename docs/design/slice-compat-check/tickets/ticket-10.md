---
id: 10
type: module
slice: slice-compat-check
blocked_by: [01, 02]
inputs: [TASK.md, docs/design/slice-compat-check/use-case.md, docs/design/slice-compat-check/module-tree.md, docs/design/slice-compat-check/c4.md, api-specification/config.schema.json, api-specification/report.schema.json, api-specification/exit-codes.md]
outputs: [README.md]
io: none
skills: []
---

### TICKET S-compat-check.10 — README: usage + build-run + Карта режимов отказа

**io:** `none`   →  skills: (none beyond core `program-implementation`)

**Context:** write the root `README.md` from the already-frozen design (do not invent anything not in
these sources). Target shape is `cli` → `harness/target-profiles.json` `profiles.cli.readme_sections =
["failure-map", "usage", "build-run"]` — cover exactly these three, plus the standard what-this-is intro.
This ticket is independent of wiring (ticket 09) — it documents the frozen contract/use-case, not the
implementation; may run in parallel with 09.

- **What this is:** one paragraph from `TASK.md` (Go CLI, compares consumer OpenAPI to provider
  master OpenAPI, pure comparison — no stubs, no test execution, no SDK generation).
- **Usage:** `pinout-openapi run <contract-tests.yaml>` — config shape from `api-specification/
  config.schema.json` (consumer `{spec_path, name, operations[]{path,method}}`, provider `{spec_url XOR
  spec_path, name}`, settings `{log_level, save_json_report, json_report_file, timeout}`); output = JSON
  report (`api-specification/report.schema.json` shape: `verdict`, `consumer.name`,
  `provider.{name,source}`, `operations[]{path,method,status,findings[]}`) written to
  `json_report_file` (when `save_json_report=true`) **plus** a stdout line, always.
- **Build & run:** `go build ./... && go test ./...` (unit tests) · `go run ./cmd/app run
  <config>` · `./component-tests/scripts/run-tests.sh` (component tests, Docker — not `go test` from the
  host).
- **## Карта режимов отказа (MUST, this exact heading):** one row per `api-specification/exit-codes.md`
  entry — 9 `error.code`s + the `compatible`/`incompatible` verdict rows, each with its exit code and a
  one-line cause (source: `contracts.md` §"Error model" + `exit-codes.md`, do not invent new causes):
  `0` compatible · `1` incompatible (`INCOMPATIBLE`) · `2` `CONFIG_INVALID`/`SPEC_UNREADABLE`/
  `SPEC_PARSE_ERROR`/`OP_NOT_IN_CONSUMER` · `3` `PROVIDER_UNREACHABLE`/`PROVIDER_TIMEOUT`/
  `PROVIDER_PARSE_ERROR`/`REPORT_WRITE_ERROR`.

**Subagent instruction:** write `README.md` at repo root (overwrite the scaffold template's generic
README) covering exactly: intro, Usage, Build & run, `## Карта режимов отказа`. Pull every fact from the
`inputs` list above — do not describe internal package structure beyond what a consumer/dev needs to run
the tool (no need to enumerate `internal/compat-check/*` sub-packages here, that lives in `c4.md`/
`module-tree.md`). Run `md-formatting` self-check before finishing (blank lines around block elements,
fenced code-blocks with a language tag).

**Verify:** the four required sections/headings are present (grep for `## Карта режимов отказа` at
minimum); no dangling links.

**Acceptance:** `README.md` exists at repo root with intro + Usage + Build & run + `## Карта режимов
отказа` (9 exit-code rows, matching `exit-codes.md` 1:1); no invented flags/fields beyond
`config.schema.json`/`report.schema.json`.

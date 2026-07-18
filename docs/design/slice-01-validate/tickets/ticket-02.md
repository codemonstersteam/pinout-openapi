---
id: 02
type: component
slice: slice-01-validate
blocked_by: [01]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/use-case.md, api-specification/config.schema.json, api-specification/report.schema.json]
outputs: [component-tests/features/validate.feature, component-tests/steps/validate_steps.go]
skills: [component-tests]
---

### TICKET 02 — slice-01-validate/component: realize the 6 designed scenarios as RED `.feature`

**io:** — (test artifact) → skills: `component-tests`

**What this ticket does:** mechanically lay the **already-designed** component set (`contracts.md`
§"Component scenarios (DESIGN half)" + Gherkin outline) into an executable `.feature` + step-defs +
stubs/fixtures in the scaffolded `component-tests/` harness, tag every scenario `@wip`, drive to
**RED by business reason** (the modules do not exist yet). Invent **no** new scenario; the set is
frozen at **6** (`1 happy + 5 adapter branches`).

**Scenarios to realize (verbatim from `contracts.md`, black-box — assert `(exit code, stdout JSON)`):**
1. happy — forward-compatible pair → exit 0, schema-valid report `compatible=true`, `errors==[]`, `uncovered_operations[]` listed.
2. `CONFIG_ERROR` → exit 2 (config not found / unreadable / malformed YAML / schema-invalid / `spec_*` not exactly-one).
3. `FILE_NOT_FOUND` → exit 3, `errors[0].code=FILE_NOT_FOUND` (consumed-contract or provider spec file missing).
4. `PARSE_ERROR` → exit 3, `errors[0].code=PARSE_ERROR` (consumed-contract or provider spec unparseable).
5. `HTTP_ERROR` → exit 3, `errors[0].code=HTTP_ERROR` (provider `spec_url` unreachable / non-2xx).
6. `TIMEOUT_ERROR` → exit 3, `errors[0].code=TIMEOUT_ERROR` (`spec_url` fetch > `settings.timeout`).

**Fixtures/stubs:** per-scenario config + consumed-contract + provider-spec fixtures; scenarios 5 & 6
need a **real-protocol HTTP stub** for `spec_url` (503 / stall-past-timeout) — one-shot-binary flavour.
Do **NOT** author unit-level verdict cases here (R1–R4, exit 1) — those are unit boundaries of
`DeriveProviderOperation`/`CompareOperation` (ADR-0002), not component scenarios.

**Dependencies:** scaffolded `component-tests/` harness (ticket 01).

**Verify:** `bash component-tests/scripts/run-tests.sh` — all 6 scenarios present and **RED** (fail
because the binary is still the placeholder), each tagged `@wip`.

**Acceptance:** all slice component scenarios exist, tagged `@wip`, **RED-ready** (RED for a business
reason, not a harness error). Removing `@wip` + GREEN is the **@fagan acceptance step**, NOT this ticket.

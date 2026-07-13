---
id: 02
type: component
slice: slice-compat-check
blocked_by: [01]
inputs: [docs/design/slice-compat-check/use-case.md, docs/design/slice-compat-check/contracts.md, api-specification/config.schema.json, api-specification/report.schema.json, api-specification/exit-codes.md]
outputs: [component-tests/features/compat-check.feature, component-tests/fixtures/, component-tests/compose/provider-stub.Dockerfile, component-tests/docker-compose.test.yml, component-tests/steps/compat_check_steps.go]
skills: [component-tests]
---

### TICKET S-compat-check.02 — component tests: realize the 9-scenario set, drive RED, tag `@wip`

**io:** — (not applicable to `type: component`)   →  skills: `component-tests`

**Context (design already done — this ticket REALIZES, invents nothing new):**
`contracts.md` §"Component scenarios" already designed the full scenario set by formula
(`N = 1 happy + Σ distinguishable I/O-adapter branches` = 8, **+ CS2** as a 9th success-family verdict
guard). Lay these into `component-tests/features/compat-check.feature`, one Scenario per row, **verbatim
Given→outcome wording from `contracts.md`**, ALL tagged `@wip`:

| # | Scenario | Exit | Fixture shape |
|---|---|---|---|
| CS1 | all configured operations compatible (happy) + **deterministic report** | 0 | valid config + both specs load, no findings; run twice → byte-identical `json_report_file` (excl. `generated_at`) |
| CS2 | ≥1 operation fails a compatibility rule (verdict guard) | 1 | valid config + both specs load, ≥1 rule fails |
| CS3 | config missing/invalid/conflict/empty operations | 2 | unreadable `contract-tests.yaml` |
| CS4 | consumer spec missing or unreadable | 2 | `consumer.spec_path` points nowhere |
| CS5 | consumer spec not valid OpenAPI 3.x | 2 | consumer doc unparseable |
| CS6 | provider spec unreachable / HTTP non-2xx | 3 | `spec_url` stub returns 500/unreachable |
| CS7 | provider fetch exceeds `settings.timeout` | 3 | `spec_url` stub delays past `timeout` |
| CS8 | provider spec not valid OpenAPI 3.x | 3 | provider doc unparseable |
| CS9 | JSON report file cannot be written | 3 | unwritable `json_report_file` |

Assert `(exit code, stdout JSON)` as a black box — this is a **one-shot binary** CLI component test
(`ctest: one-shot-binary` per `harness/target-profiles.json` cli profile), not a long-running service.
CS6/CS7 need a **real-protocol HTTP stub** in `component-tests/docker-compose.test.yml` serving the
provider spec (add a `provider-stub` service via `component-tests/compose/provider-stub.Dockerfile` —
one endpoint returning 500/unreachable for CS6, one delaying past `timeout` for CS7, one serving a
malformed doc for CS8); **only the core is never stubbed**. `component-tests/fixtures/` holds one
`contract-tests.yaml` + consumer/provider spec fixture pair per scenario (reuse across CS1/CS2 where the
only difference is provider/consumer content, not config shape).

**CS1 determinism assertion (Gate #1 A4 — pinned to CS1, this ticket owns it as a RED `@wip` scenario):**
CS1 additionally asserts report byte-stability. Realize it as an extra `Then` on the CS1 scenario (keep
`@wip`): run the binary **twice** on the **byte-identical** fixture set, both exit `0`, and the two
`json_report_file` outputs are **byte-identical once `generated_at` is excluded** (filter with
`jq 'del(.generated_at)'` before the diff, or hold `generated_at` fixed). This is the single owning test
for the BRD/report-canon determinism requirement (`generated_at` is optional/excluded per
`report.schema.json`) — do not re-cover it only at `@fagan` acceptance. It stays RED (placeholder head
returns `NOT_IMPLEMENTED`) until module tickets 03–08 land.

**OP_NOT_IN_CONSUMER (3a) and INCOMPATIBLE-rule-boundaries are NOT scenarios here** — per `contracts.md`
§Component-scenarios anti-gaming reconciliation (`7==7==7`) they are **unit** boundaries (module tickets
03/05), not I/O-adapter failures. Do not add scenarios for them.

**Subagent instruction:** replace `component-tests/features/smoke.feature`'s domain content with
`compat-check.feature` (9 scenarios above, all `@wip`); add any missing generic CLI step (run binary with
args, assert exit code, assert stdout JSON field) to `component-tests/steps/compat_check_steps.go` — reuse
`component-tests/steps/cli_steps.go` generics where they already fit, do not duplicate; wire the
`provider-stub` compose service. Do **not** implement `internal/compat-check/` logic here — scenarios stay
RED (exit `3` `NOT_IMPLEMENTED` from the placeholder head) until module tickets 03–08 land.

**Verify:** `./component-tests/scripts/run-tests.sh` runs and reports **RED by business reason** (each
scenario fails because the placeholder head returns `NOT_IMPLEMENTED`, not because of a missing step-def
or a harness error) — RED-reason must be inspectable in the run output.

**Acceptance:** `component-tests/features/compat-check.feature` contains exactly 9 scenarios (CS1–CS9), all
tagged `@wip`; `run-tests.sh` executes them and every one is RED for the correct business reason (not a
harness/step-def error); the `provider-stub` compose service exists and answers all three provider-side
shapes (unreachable/500, delayed, malformed). Removing `@wip` is **NOT** this ticket's job — that is
`@fagan`'s slice-acceptance step, after all module tickets are GREEN.

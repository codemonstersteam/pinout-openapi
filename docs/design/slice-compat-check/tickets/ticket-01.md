---
id: 01
type: scaffold
slice: slice-compat-check
blocked_by: []
inputs: [.agent/planner/target, harness/scaffold.sh, harness/target-profiles.json]
outputs: [.dockerignore, .gitignore, README.md, cmd/app/main.go, cmd/app/main_test.go, component-tests/.gitignore, component-tests/Dockerfile.runtime, component-tests/compose/envs/healthy.env, component-tests/compose/tool.Dockerfile, component-tests/docker-compose.test.yml, component-tests/features/smoke.feature, component-tests/go.mod, component-tests/go.sum, component-tests/scripts/run-tests.sh, component-tests/steps/cli_steps.go, component-tests/steps/main_test.go, component-tests/steps/world.go, config.yaml, go.mod, go.sum, internal/shared/config/config.go, internal/shared/config/config_test.go, internal/shared/report/writer.go]
skills: [service-scaffold]
---

### TICKET S-compat-check.01 — scaffold: clone `template-go-cli` → runnable placeholder

**io:** — (not applicable to `type: scaffold`)   →  skills: `service-scaffold`

**Context:**
- Target shape marker: `.agent/planner/target` = `cli` → `harness/target-profiles.json` `profiles.cli.template`
  = `template-go-cli` (NOT `template-go-api` — this slice is a CLI, never an HTTP service).
- Run `sh harness/scaffold.sh pinout-openapi` (service-name = repo/service name; scaffold.sh resolves the
  template path itself from the `cli` shape marker — do not pass a template path explicitly and do not
  hand-assemble a runner).
- The frozen contract already exists at `api-specification/{config.schema.json,report.schema.json,exit-codes.md}`
  — `scaffold.sh` preserves it (does not overwrite `api-specification/`) because it is non-empty.
- **`outputs` above = EXACTLY the template's tracked files** (as `git archive HEAD` ships them from
  `template-go-cli`), with the go-module renamed `template-go-cli` → `pinout-openapi` in every `.go` file and
  `go.mod`. Do **not** invent a `cmd/<slug>/main.go` — the binary entrypoint stays `cmd/app/main.go`
  (`scaffold.sh` renames the go-module, never the `cmd/` directory). `internal/example/` is the generic
  placeholder slice the template ships (NOT this slice's code) — later module tickets add
  `internal/compat-check/`; do not delete `internal/example/` in this ticket (the wiring ticket retires it
  once `internal/compat-check/` is wired in).

**Subagent instruction:** run `sh harness/scaffold.sh pinout-openapi` from the repo root. Do not edit template
files by hand. Do not write any `internal/compat-check/` code — that belongs to later module tickets.

**Verify:** `go build ./...` at repo root must exit 0 (scaffold.sh already runs this internally and fails the
ticket on red — the two follow-up checks are `go vet ./...` and `go run ./cmd/app version` → exit 0).

**Acceptance:** repo builds green (`go build ./...` exit 0); `go run ./cmd/app version` prints a version
string and exits 0; `api-specification/` untouched (still the frozen `config.schema.json`/`report.schema.json`/
`exit-codes.md`); no slice-named `cmd/` directory was invented.

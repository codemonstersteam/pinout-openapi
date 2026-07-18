---
id: 04
type: module
slice: slice-01-validate
blocked_by: [01, 02, 03]
inputs: [docs/design/slice-01-validate/contracts.md, docs/design/slice-01-validate/module-tree.md]
outputs: [internal/validate/cli/parse.go]
io: none
skills: []
---

### TICKET 04 — slice-01-validate/cli.Parse: ingress door (argv → Invocation)

**io:** none → skills: [] (ingress adapter — design per the **`cli-io`** discipline, cobra; the driving
adapter of the hexagon is `io: none`, never `http-io`). **Not unit-tested** (parse-map; asserted by
component scenarios).

**Context (only this module):**
- contract (`contracts.md` §cli.Parse): `Parse(args: []string) -> Result[Invocation, Error]`.
  Input = process argv. Deps = —. Parse `validate <config.yaml>` + flags (`--json`, `--verbose`,
  `--version`, `--help`) into ONE flat `Invocation{ConfigPath, JSONReport, Verbose}` — **path only**,
  no file I/O, no logic.
  - antecedent: `args` is the real invocation vector.
  - consequent: Ok `Invocation`; Fail `ErrConfig` (missing positional / unknown flag) → `CONFIG_ERROR`, exit 2.
- unit tests: **none** (ingress parse-map).
- component scenario(s) to green: contributes to scenario 2 (`CONFIG_ERROR`, exit 2) — greened later.

**Dependencies:** `Invocation` type + `ErrConfig` sentinel (ticket 03). cobra (from scaffold).

**Subagent instruction:** implement `cli.Parse` in `internal/validate/cli/parse.go` → `go build`/`vet` →
done. Do not touch other modules; do not drive component scenarios GREEN.

**Verify:** `go build ./internal/validate/... && go vet ./internal/validate/...`.

**Acceptance:** package `cli` builds and vets clean; no units (parse-map). Component GREEN is @fagan's step.

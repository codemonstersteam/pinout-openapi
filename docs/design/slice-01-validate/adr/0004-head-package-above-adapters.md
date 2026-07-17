# Shared domain types move to a leaf package so the head can call adapters without a cycle

module-tree.md's File layout put the shared slice domain types (`domain.go`), the sentinel
errors (`errors.go`) AND the composition-root head (`head.go`) all in one package
`internal/validate`, while every adapter subpackage (`config`, `contract`, `provider`,
`compare`, `report`, `cli`) imports `internal/validate` for those domain types. That layout is
**inherently cyclic** in Go: the head calls the adapters directly (`config.NewConfig`,
`compare.NewComparison`, `compare.CompareContracts`, `report.FoldReport` — contracts.md lists
their Deps as "—", i.e. direct calls, not injected), so `internal/validate` would import
`internal/validate/compare`, which imports `internal/validate` back → `import cycle not allowed`.

Decision: the head stays at its ticket-declared path `internal/validate/head.go` in
`package validate` (ticket-16 `outputs: [internal/validate/head.go]`; the done-marker guardrail
enforces this). The shared value types + sentinel errors moved DOWN into a new leaf package
`internal/validate/domain` (`package domain`), which imports nothing internal. Every adapter and
the head now import `internal/validate/domain` for those types. Import graph is acyclic:
`validate(head) → {domain, config, compare, report}`; each adapter → `domain`; `domain` →
nothing internal.

`ProcessValidate`'s signature, the `Deps` ports, the frozen API contract (config/report schemas,
exit codes) and the behaviour of all 15 already-green modules are unchanged — this is a
package-placement fix, not a contract change. The 15 modules changed only the import path
(`internal/validate` → `internal/validate/domain`) and the type qualifier (`validate.` →
`domain.`); every `validate.` occurrence in them was verified to be a domain-type selector
before the rename, and their unit tests re-ran green, proving behaviour preserved.

The mirror option (move the head out to its own package, leave the types in `internal/validate`)
is topologically identical and touches fewer files, but it relocates `head.go` away from the
path ticket-16 declares and the done-marker guardrail enforces — so it was rejected in favour of
keeping the head where the plan pins it. Reversing this decision (folding the domain types back
beside the head) re-introduces the cycle, hence recorded. The scaffold placeholder packages
`internal/validate/{io,logic,domain(old)}` were template cruft; the old `domain/` placeholder
(`Identifier` sample) was replaced by this real domain package.

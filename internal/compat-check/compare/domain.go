// Package compare hides the 5 compatibility rules — the pure core, the reason the tool exists
// (module-tree.md "The secret each module hides"). It never imports cobra/net/http/os/the
// kin-openapi loader — only internal/compat-check/config's Operation type and
// internal/compat-check/spec's already-parsed Spec model (ticket-05 "Dependencies").
package compare

import "sort"

// Rule identifies which of the 5 compatibility rules a Finding violates. The string value is the
// underscore form mandated by the frozen api-specification/report.schema.json findings[].rule
// enum (contracts.md: "rule is the UNDERSCORE form ... per frozen report.schema.json").
type Rule string

const (
	RulePresence              Rule = "presence"
	RuleRequestContravariance Rule = "request_contravariance"
	RuleResponseCovariance    Rule = "response_covariance"
	RuleStatusCodes           Rule = "status_codes"
	RuleContentTypes          Rule = "content_types"
)

// Verdict is the overall / per-operation compatibility result (report.schema.json "verdict" and
// "operations[].status" enums).
type Verdict string

const (
	VerdictCompatible   Verdict = "compatible"
	VerdictIncompatible Verdict = "incompatible"
)

// Finding is one mismatch that makes an operation incompatible (report.schema.json
// operations[].findings[] shape: rule, location, detail all required).
type Finding struct {
	Rule     Rule
	Location string
	Detail   string
}

// OperationResult is the per-operation compatibility outcome (report.schema.json operations[]
// shape: path, method, status, findings). Invariant (enforced by construction in
// compareOperation, never assembled independently): Status == compatible iff len(Findings) == 0.
type OperationResult struct {
	Path     string
	Method   string
	Status   Verdict
	Findings []Finding
}

// Report is the canon report.schema.json output shape: verdict, consumer.name,
// provider.{name,source}, operations[].
type Report struct {
	Verdict        Verdict
	ConsumerName   string
	ProviderName   string
	ProviderSource string
	Operations     []OperationResult
}

// sortedKeys returns m's keys in a deterministic (sorted) order — used wherever a check walks a
// kin-openapi map (Content, Schemas), so finding order — and, downstream, the assembled report —
// is stable across repeated runs on byte-identical inputs (module-tree.md "no map-iteration
// order" determinism rationale, CS1 acceptance clause).
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

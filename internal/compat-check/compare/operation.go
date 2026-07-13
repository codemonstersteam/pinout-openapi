package compare

import (
	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// Comparison unites the checked operations, the consumer + provider parsed specs, and the report
// identity strings (consumer name, provider name, resolved provider source) needed to assemble a
// complete Report — one data argument into Compare (module-tree.md "Head-pipe pseudocode":
// Compare(Comparison) -> Report takes exactly one argument, and report.schema.json requires
// consumer.name/provider.name/provider.source on every Report).
//
// DEVIATION (flagged, see ticket-05 completion notes / .agent/memory.md): contracts.md's prose
// literally lists NewComparison(checkedOps, consumer, provider) — 3 args — and
// buildReport(config, results, providerSource) as a separate call taking the whole Config. Taken
// together with the single-arg Compare(comparison) pseudocode, that is unsatisfiable without
// widening one of the two signatures: Compare has only `comparison` to work with, so buildReport's
// identity inputs must already be inside Comparison. Widening buildReport to accept a
// config.Config would also violate this ticket's explicit Dependencies line ("imports
// internal/compat-check/config ... Operation type" only — not Config/Consumer/Provider). Widening
// NewComparison's parameter list (plain strings, not config domain types) reconciles both
// constraints. ticket-06 (head.go/register.go) must pass consumerName/providerName/providerSource
// explicitly when calling NewComparison.
type Comparison struct {
	ops            []config.Operation
	consumer       spec.Spec
	provider       spec.Spec
	consumerName   string
	providerName   string
	providerSource string
}

// NewComparison unites checkedOps with the loaded consumer and provider specs, plus the report
// identity strings buildReport needs (see the DEVIATION note above).
func NewComparison(checkedOps CheckedOps, consumer, provider spec.Spec, consumerName, providerName, providerSource string) Comparison {
	return Comparison{
		ops:            checkedOps.Operations(),
		consumer:       consumer,
		provider:       provider,
		consumerName:   consumerName,
		providerName:   providerName,
		providerSource: providerSource,
	}
}

// comparisonOp is one operation's resolved view: the configured operation plus its consumer and
// (possibly absent) provider openapi3.Operation. providerOp is nil when checkPresence is the only
// rule that can fire (contracts.md "checkPresence").
type comparisonOp struct {
	op         config.Operation
	consumerOp *openapi3.Operation
	providerOp *openapi3.Operation
}

// compareOperation runs the 5 compatibility rules against one resolved operation and concatenates
// their findings. status = incompatible iff findings ≥ 1, else compatible (contracts.md
// "operation.go"). When providerOp is nil, only checkPresence can meaningfully fire — the other 4
// rules are skipped rather than dereferencing a nil provider operation.
func compareOperation(co comparisonOp) OperationResult {
	findings := checkPresence(co.consumerOp, co.providerOp)
	if co.providerOp != nil {
		findings = append(findings, checkRequestContravariance(co.consumerOp, co.providerOp)...)
		findings = append(findings, checkResponseCovariance(co.consumerOp, co.providerOp)...)
		findings = append(findings, checkStatusCodes(co.consumerOp, co.providerOp)...)
		findings = append(findings, checkContentTypes(co.consumerOp, co.providerOp)...)
	}

	status := VerdictCompatible
	if len(findings) > 0 {
		status = VerdictIncompatible
	}
	return OperationResult{
		Path:     co.op.Path().String(),
		Method:   co.op.Method().String(),
		Status:   status,
		Findings: findings,
	}
}

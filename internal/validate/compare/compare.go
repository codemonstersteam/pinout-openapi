// Package compare — slice-01-validate: CompareContracts (ticket 13, contracts.md
// §CompareContracts; module-tree.md "compare"). Pure fold over the scoped operations
// (c.ScopedOps): per operation, look up the recorded ConsumedOperation, derive the
// provider operation (DeriveProviderOperation, R1, provider package) and concatMap
// CompareOperation's violations (R2..R4, rules.go). The loop lives here, not in the
// head (ProcessValidate stays a linear ROP pipe — module-tree.md).
package compare

import (
	"fmt"
	"strings"

	"pinout-openapi/internal/validate/domain"
	"pinout-openapi/internal/validate/provider"
)

// CompareContracts — contracts.md §CompareContracts: CompareContracts(c: Comparison)
// -> ComparisonOutcome. For each c.ScopedOps entry, derives the provider operation and
// folds CompareOperation's violations (concatMap); computes uncovered_operations[] —
// operations recorded in the consumed contract but outside the configured scope
// (informational, no verdict/exit-code effect); carries Provenance through unchanged.
// Antecedent: valid Comparison. No failing consequent — Violations == [] IS the
// compatible verdict (ADR-0002), not an absence of a result.
func CompareContracts(c domain.Comparison) domain.ComparisonOutcome {
	var violations []domain.Violation

	for _, ref := range c.ScopedOps {
		consumedOp := findConsumedOperation(c.Consumed.Operations, ref)

		providerOp, present := provider.DeriveProviderOperation(c.Spec, ref)
		var providerOpArg *domain.ProviderOperation
		if present {
			providerOpArg = &providerOp
		}

		violations = append(violations, CompareOperation(consumedOp, providerOpArg)...)
	}

	return domain.ComparisonOutcome{
		Violations:   violations,
		UncoveredOps: uncoveredOperations(c.Spec, c.ScopedOps),
		Provenance:   c.Provenance,
	}
}

// findConsumedOperation — the recorded ConsumedOperation matching ref, or a zero-value
// (empty Sends/Reads) if the scope names an operation absent from the consumed
// contract (nothing was ever recorded sending/reading anything for it).
func findConsumedOperation(ops []domain.ConsumedOperation, ref domain.OperationRef) domain.ConsumedOperation {
	for _, op := range ops {
		if op.Ref == ref {
			return op
		}
	}
	return domain.ConsumedOperation{Ref: ref}
}

// uncoveredOperations — PROVIDER operations outside the configured scope
// (c.ScopedOps == cfg.operations): the provider's full operation surface
// (provider.EnumerateOperations over c.Spec) minus the operations the consumer
// depends on. Informational only (report.schema.json #/properties/uncovered_operations;
// CONTEXT.md "Uncovered operation"; contracts.md §CompareContracts), no effect on the
// verdict or the exit code. NOT the consumer's out-of-scope operations — the frozen
// design is unanimous that uncovered means provider surface the consumer does not use.
func uncoveredOperations(spec domain.ProviderSpec, scoped []domain.OperationRef) []string {
	inScope := make(map[domain.OperationRef]bool, len(scoped))
	for _, ref := range scoped {
		inScope[normalizeRef(ref)] = true
	}

	var uncovered []string
	for _, op := range provider.EnumerateOperations(spec) {
		if !inScope[normalizeRef(op)] {
			uncovered = append(uncovered, fmt.Sprintf("%s %s", op.Method, op.Path))
		}
	}
	return uncovered
}

// normalizeRef — метод в нижний регистр для сравнения scope↔провайдер по одному ключу
// (config.schema.json задаёт lower-case, EnumerateOperations нормализует так же).
func normalizeRef(ref domain.OperationRef) domain.OperationRef {
	ref.Method = strings.ToLower(ref.Method)
	return ref
}

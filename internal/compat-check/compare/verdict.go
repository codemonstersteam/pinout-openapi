package compare

// aggregateVerdict rolls per-operation results into the overall verdict: any incompatible
// operation makes the whole comparison incompatible (contracts.md "verdict.go").
func aggregateVerdict(results []OperationResult) Verdict {
	for _, r := range results {
		if r.Status == VerdictIncompatible {
			return VerdictIncompatible
		}
	}
	return VerdictCompatible
}

// buildReport assembles the report.schema.json canon shape. Invariant by construction:
// status=compatible ⇒ 0 findings; status=incompatible ⇒ ≥1 finding — guaranteed because each
// result's Status and Findings both came from the same compareOperation call, never assembled
// independently here.
func buildReport(consumerName, providerName, providerSource string, results []OperationResult) Report {
	return Report{
		Verdict:        aggregateVerdict(results),
		ConsumerName:   consumerName,
		ProviderName:   providerName,
		ProviderSource: providerSource,
		Operations:     results,
	}
}

// Compare is the pure-core entry the head calls directly (ticket-06): compareOperation per
// configured operation → aggregateVerdict → buildReport. Returns a Report for both verdicts —
// INCOMPATIBLE is a verdict carried forward, never a pipe error (module-tree.md "Key design
// decision — incompatible is a verdict, not a pipe error").
func Compare(comparison Comparison) Report {
	results := make([]OperationResult, 0, len(comparison.ops))
	for _, op := range comparison.ops {
		path := op.Path().String()
		method := op.Method().String()
		results = append(results, compareOperation(comparisonOp{
			op:         op,
			consumerOp: comparison.consumer.Operation(path, method),
			providerOp: comparison.provider.Operation(path, method),
		}))
	}
	return buildReport(comparison.consumerName, comparison.providerName, comparison.providerSource, results)
}

// Package compare — the comparison core of slice-01-validate: unites the three
// already-valid pipe outputs (Config, ConsumedContract, ProviderSpec) into one
// Comparison, then folds R1..R4 over the scoped operations (CompareContracts,
// ticket 13; CompareOperation/typesMatch, ticket 11).
package compare

import (
	"pinout-openapi/internal/validate/domain"
)

// NewComparison — the sanctioned uniting constructor (2+ entities ⇒ one node,
// contracts.md §NewComparison): builds Comparison{ScopedOps, Consumed, Spec,
// Provenance} — the scope (cfg.Operations) intersected with the consumed per-op
// data and the provider spec. Antecedent: all three inputs already valid (NewConfig,
// ContractStore.Load, SpecLoader.Load). Consequent: no failing antecedent — Result[T,
// Error] is the Go idiom (T, error), err is always nil here.
func NewComparison(cfg domain.Config, consumed domain.ConsumedContract, spec domain.ProviderSpec) (domain.Comparison, error) {
	return domain.Comparison{
		ScopedOps:  cfg.Operations,
		Consumed:   consumed,
		Spec:       spec,
		Provenance: consumed.Provenance,
	}, nil
}

package compare

import (
	"reflect"
	"testing"

	"pinout-openapi/internal/validate/domain"
)

// TestNewComparison_UnitesValidInputs — the 1 happy unit (module-tree.md formula:
// NewComparison → 1, uniting constructor over already-valid inputs, no failing
// antecedent). Asserts the united Comparison carries cfg.Operations as ScopedOps,
// the consumed contract and provider spec verbatim, and Provenance echoed from
// consumed.Provenance.
func TestNewComparison_UnitesValidInputs(t *testing.T) {
	ops := []domain.OperationRef{
		{Path: "/pets", Method: "GET"},
		{Path: "/pets/{id}", Method: "GET"},
	}
	provenance := domain.Provenance{
		Provider:        "petstore",
		ProviderVersion: "1.0.0",
		CapturedHash:    "abc123",
	}
	cfg := domain.Config{
		ConsumerName:         "pinout-cli",
		ConsumedContractPath: "consumed.json",
		Operations:           ops,
		Provider:             domain.ProviderConfig{Name: "petstore", SpecPath: "spec.yaml"},
		Settings:             domain.Settings{LogLevel: "info", Timeout: 30},
	}
	consumed := domain.ConsumedContract{
		Operations: []domain.ConsumedOperation{
			{
				Ref:   domain.OperationRef{Path: "/pets", Method: "GET"},
				Sends: map[string]string{"limit": "integer"},
				Reads: map[string]string{"id": "string"},
			},
		},
		Provenance: provenance,
	}
	spec := domain.ProviderSpec{Doc: map[string]any{"openapi": "3.0.0"}}

	got, err := NewComparison(cfg, consumed, spec)
	if err != nil {
		t.Fatalf("NewComparison() unexpected error: %v", err)
	}

	want := domain.Comparison{
		ScopedOps:  ops,
		Consumed:   consumed,
		Spec:       spec,
		Provenance: provenance,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NewComparison() = %+v, want %+v", got, want)
	}
}

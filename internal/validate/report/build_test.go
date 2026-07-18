package report

import (
	"testing"

	"pinout-openapi/internal/validate/domain"
)

// TestFoldReport_Happy — 1 happy: no violations ⇒ compatible=true, errors==[],
// schema_version + provenance echoed (contracts.md §FoldReport).
func TestFoldReport_Happy(t *testing.T) {
	outcome := domain.ComparisonOutcome{
		Violations:   nil,
		UncoveredOps: nil,
		Provenance: domain.Provenance{
			Provider:        "acme-provider",
			ProviderVersion: "1.2.3",
			CapturedHash:    "sha256:deadbeef",
		},
	}

	got := FoldReport(outcome)

	if got.SchemaVersion != "1.0" {
		t.Errorf("SchemaVersion = %q, want %q", got.SchemaVersion, "1.0")
	}
	if !got.Compatible {
		t.Errorf("Compatible = false, want true")
	}
	if len(got.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", got.Errors)
	}
	if got.Provenance != outcome.Provenance {
		t.Errorf("Provenance = %+v, want %+v", got.Provenance, outcome.Provenance)
	}
	if got.UncoveredOperations != nil {
		t.Errorf("UncoveredOperations = %v, want nil", got.UncoveredOperations)
	}
}

// TestFoldReport_NonEmptyErrorsIncompatible — branch: non-empty Violations ⇒
// compatible=false, errors[] carries the violations verbatim.
func TestFoldReport_NonEmptyErrorsIncompatible(t *testing.T) {
	outcome := domain.ComparisonOutcome{
		Violations: []domain.Violation{
			{Code: "OP_NOT_IN_PROVIDER", Message: "operation not found in provider spec"},
		},
	}

	got := FoldReport(outcome)

	if got.Compatible {
		t.Errorf("Compatible = true, want false")
	}
	if len(got.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(got.Errors))
	}
	if got.Errors[0].Code != "OP_NOT_IN_PROVIDER" {
		t.Errorf("Errors[0].Code = %q, want %q", got.Errors[0].Code, "OP_NOT_IN_PROVIDER")
	}
}

// TestFoldReport_UncoveredOperationsPopulated — branch: informational
// uncovered_operations[] carries UncoveredOps through unaffected by the verdict.
func TestFoldReport_UncoveredOperationsPopulated(t *testing.T) {
	outcome := domain.ComparisonOutcome{
		Violations:   nil,
		UncoveredOps: []string{"GET /widgets", "POST /widgets/{id}/archive"},
	}

	got := FoldReport(outcome)

	if !got.Compatible {
		t.Errorf("Compatible = false, want true")
	}
	if len(got.UncoveredOperations) != 2 {
		t.Fatalf("len(UncoveredOperations) = %d, want 2", len(got.UncoveredOperations))
	}
	if got.UncoveredOperations[0] != "GET /widgets" {
		t.Errorf("UncoveredOperations[0] = %q, want %q", got.UncoveredOperations[0], "GET /widgets")
	}
}

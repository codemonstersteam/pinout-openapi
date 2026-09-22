package report

import (
	"testing"
	"time"

	"pinout-openapi/internal/validate/domain"
)

// fixedClock — детерминированные часы для happy-веток (D10: Fold не читает системное
// время, порт инжектится).
type fixedClock struct{ at time.Time }

func (c fixedClock) Now() time.Time { return c.at }

// countingClock — якорь D10: считает вызовы Now() (ровно один Fold = ровно один вызов).
type countingClock struct{ calls int }

func (c *countingClock) Now() time.Time {
	c.calls++
	return time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)
}

// TestReporterFold_Happy — 1 happy: no violations ⇒ compatible=true, errors==[],
// константы канона 1.1 (schema_version/validator/interaction), consumer.name,
// generated_at из порта часов, provenance echoed (contracts.md §Reporter.Fold).
func TestReporterFold_Happy(t *testing.T) {
	fixed := time.Date(2026, 9, 22, 12, 34, 56, 0, time.UTC)
	outcome := domain.ComparisonOutcome{
		Violations:   nil,
		UncoveredOps: nil,
		Provenance: domain.Provenance{
			Provider:        "acme-provider",
			ProviderVersion: "1.2.3",
			CapturedHash:    "sha256:deadbeef",
		},
		ConsumerName: "wallet-ui",
	}

	got := BuildReporter(fixedClock{fixed}).Fold(outcome)

	if got.SchemaVersion != "1.1" {
		t.Errorf("SchemaVersion = %q, want %q", got.SchemaVersion, "1.1")
	}
	if got.Validator != "pinout-openapi" {
		t.Errorf("Validator = %q, want %q", got.Validator, "pinout-openapi")
	}
	if got.Interaction != "sync" {
		t.Errorf("Interaction = %q, want %q", got.Interaction, "sync")
	}
	if got.Consumer.Name != "wallet-ui" {
		t.Errorf("Consumer.Name = %q, want %q", got.Consumer.Name, "wallet-ui")
	}
	if got.Consumer.Version != "" {
		t.Errorf("Consumer.Version = %q, want empty (no source, never a placeholder)", got.Consumer.Version)
	}
	if got.GeneratedAt != "2026-09-22T12:34:56Z" {
		t.Errorf("GeneratedAt = %q, want %q (RFC3339 UTC Z, second precision)", got.GeneratedAt, "2026-09-22T12:34:56Z")
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

// TestReporterFold_NonEmptyErrorsIncompatible — branch: non-empty Violations ⇒
// compatible=false, errors[] carries the violations verbatim (subject echo included).
func TestReporterFold_NonEmptyErrorsIncompatible(t *testing.T) {
	outcome := domain.ComparisonOutcome{
		Violations: []domain.Violation{
			{Code: "OP_NOT_IN_PROVIDER", Message: "operation not found in provider spec", Subject: "GET /widgets"},
		},
	}

	got := BuildReporter(fixedClock{}).Fold(outcome)

	if got.Compatible {
		t.Errorf("Compatible = true, want false")
	}
	if len(got.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(got.Errors))
	}
	if got.Errors[0].Code != "OP_NOT_IN_PROVIDER" {
		t.Errorf("Errors[0].Code = %q, want %q", got.Errors[0].Code, "OP_NOT_IN_PROVIDER")
	}
	if got.Errors[0].Subject != "GET /widgets" {
		t.Errorf("Errors[0].Subject = %q, want %q", got.Errors[0].Subject, "GET /widgets")
	}
}

// TestReporterFold_UncoveredOperationsPopulated — branch: informational
// uncovered_operations[] carries UncoveredOps through unaffected by the verdict.
func TestReporterFold_UncoveredOperationsPopulated(t *testing.T) {
	outcome := domain.ComparisonOutcome{
		Violations:   nil,
		UncoveredOps: []string{"GET /widgets", "POST /widgets/{id}/archive"},
	}

	got := BuildReporter(fixedClock{}).Fold(outcome)

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

// TestReporterFold_SingleClockRead — якорь D10 (change 001-report-schema-1.1): один
// Fold читает часы ровно один раз — весь слайс читает часы один раз на прогон.
func TestReporterFold_SingleClockRead(t *testing.T) {
	clock := &countingClock{}
	_ = BuildReporter(clock).Fold(domain.ComparisonOutcome{})
	if clock.calls != 1 {
		t.Fatalf("clock.Now() called %d times, want exactly 1 (D10)", clock.calls)
	}
}

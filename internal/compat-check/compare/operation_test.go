package compare

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// Unit-test formula: compareOperation: 1 (all rules pass → compatible) + 1 (≥1 finding →
// incompatible) = 2. NewComparison: 1 (happy, unites 3 entities) = 1. Total 3 in this file.

func TestCompareOperation(t *testing.T) {
	t.Run("all rules pass: compatible", func(t *testing.T) {
		op := mustOperation(t, "/pets", "get")
		responses := &openapi3.Responses{}
		responses.Set("200", &openapi3.ResponseRef{Value: &openapi3.Response{}})
		consumerOp := &openapi3.Operation{Responses: responses}
		providerOp := &openapi3.Operation{Responses: responses}

		result := compareOperation(comparisonOp{op: op, consumerOp: consumerOp, providerOp: providerOp})

		if result.Status != VerdictCompatible {
			t.Fatalf("status = %q, want %q", result.Status, VerdictCompatible)
		}
		if len(result.Findings) != 0 {
			t.Fatalf("findings = %v, want none", result.Findings)
		}
		if result.Path != "/pets" || result.Method != "get" {
			t.Fatalf("path/method = %q/%q, want /pets/get", result.Path, result.Method)
		}
	})

	t.Run("presence finding: incompatible", func(t *testing.T) {
		op := mustOperation(t, "/pets", "get")
		consumerOp := &openapi3.Operation{}

		result := compareOperation(comparisonOp{op: op, consumerOp: consumerOp, providerOp: nil})

		if result.Status != VerdictIncompatible {
			t.Fatalf("status = %q, want %q", result.Status, VerdictIncompatible)
		}
		if len(result.Findings) != 1 {
			t.Fatalf("findings len = %d, want 1", len(result.Findings))
		}
		if result.Findings[0].Rule != RulePresence {
			t.Fatalf("rule = %q, want %q", result.Findings[0].Rule, RulePresence)
		}
	})
}

func TestNewComparison(t *testing.T) {
	t.Run("happy: unites checkedOps + consumer + provider specs into one Report via Compare", func(t *testing.T) {
		dir := t.TempDir()
		consumerPath := filepath.Join(dir, "consumer.yaml")
		providerPath := filepath.Join(dir, "provider.yaml")
		if err := os.WriteFile(consumerPath, []byte(minimalConsumerDoc), 0o644); err != nil {
			t.Fatalf("write consumer fixture: %v", err)
		}
		if err := os.WriteFile(providerPath, []byte(minimalConsumerDoc), 0o644); err != nil {
			t.Fatalf("write provider fixture: %v", err)
		}

		consumerSpec, err := spec.NewConsumerSpecReader().Load(consumerPath)
		if err != nil {
			t.Fatalf("load consumer: %v", err)
		}

		source, err := config.NewProviderSource("", providerPath)
		if err != nil {
			t.Fatalf("NewProviderSource: %v", err)
		}
		providerSpec, err := spec.NewProviderSpecReader(time.Second).Load(source)
		if err != nil {
			t.Fatalf("load provider: %v", err)
		}

		checkedOps, err := NewCheckedOperations([]config.Operation{mustOperation(t, "/users", "get")}, consumerSpec)
		if err != nil {
			t.Fatalf("NewCheckedOperations: %v", err)
		}

		comparison := NewComparison(checkedOps, consumerSpec, providerSpec, "consumer-x", "provider-y", providerPath)
		report := Compare(comparison)

		if report.ConsumerName != "consumer-x" {
			t.Fatalf("ConsumerName = %q, want %q", report.ConsumerName, "consumer-x")
		}
		if report.ProviderName != "provider-y" {
			t.Fatalf("ProviderName = %q, want %q", report.ProviderName, "provider-y")
		}
		if report.ProviderSource != providerPath {
			t.Fatalf("ProviderSource = %q, want %q", report.ProviderSource, providerPath)
		}
		if len(report.Operations) != 1 {
			t.Fatalf("Operations len = %d, want 1", len(report.Operations))
		}
		if report.Operations[0].Path != "/users" || report.Operations[0].Method != "get" {
			t.Fatalf("op = %q/%q, want /users/get", report.Operations[0].Path, report.Operations[0].Method)
		}
		if report.Verdict != VerdictCompatible {
			t.Fatalf("Verdict = %q, want %q", report.Verdict, VerdictCompatible)
		}
	})
}

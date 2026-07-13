package compare

import "testing"

// Unit-test formula: aggregateVerdict: 1 (all compatible) + 1 (≥1 incompatible → incompatible)
// = 2. buildReport: 1 (happy, invariant holds by construction) = 1. Total 3 in this file.

func TestAggregateVerdict(t *testing.T) {
	t.Run("all operations compatible", func(t *testing.T) {
		results := []OperationResult{
			{Path: "/a", Method: "get", Status: VerdictCompatible},
			{Path: "/b", Method: "get", Status: VerdictCompatible},
		}
		if got := aggregateVerdict(results); got != VerdictCompatible {
			t.Fatalf("aggregateVerdict = %q, want %q", got, VerdictCompatible)
		}
	})

	t.Run("one incompatible operation makes the whole comparison incompatible", func(t *testing.T) {
		results := []OperationResult{
			{Path: "/a", Method: "get", Status: VerdictCompatible},
			{Path: "/b", Method: "get", Status: VerdictIncompatible, Findings: []Finding{{Rule: RulePresence}}},
		}
		if got := aggregateVerdict(results); got != VerdictIncompatible {
			t.Fatalf("aggregateVerdict = %q, want %q", got, VerdictIncompatible)
		}
	})
}

func TestBuildReport(t *testing.T) {
	t.Run("happy: assembly + consistency invariant hold by construction", func(t *testing.T) {
		results := []OperationResult{
			{Path: "/a", Method: "get", Status: VerdictCompatible, Findings: nil},
			{Path: "/b", Method: "post", Status: VerdictIncompatible, Findings: []Finding{
				{Rule: RulePresence, Location: "operation", Detail: "absent"},
			}},
		}

		report := buildReport("consumer-x", "provider-y", "https://provider.example/openapi.yaml", results)

		if report.Verdict != VerdictIncompatible {
			t.Fatalf("Verdict = %q, want %q", report.Verdict, VerdictIncompatible)
		}
		if report.ConsumerName != "consumer-x" || report.ProviderName != "provider-y" {
			t.Fatalf("identity = %q/%q, want consumer-x/provider-y", report.ConsumerName, report.ProviderName)
		}
		if report.ProviderSource != "https://provider.example/openapi.yaml" {
			t.Fatalf("ProviderSource = %q", report.ProviderSource)
		}
		if len(report.Operations) != 2 {
			t.Fatalf("Operations len = %d, want 2", len(report.Operations))
		}

		for _, op := range report.Operations {
			switch op.Status {
			case VerdictCompatible:
				if len(op.Findings) != 0 {
					t.Fatalf("compatible op %q has findings: %v", op.Path, op.Findings)
				}
			case VerdictIncompatible:
				if len(op.Findings) == 0 {
					t.Fatalf("incompatible op %q has no findings", op.Path)
				}
			}
		}
	})
}

package compare

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Unit-test formula: checkRequestContravariance: 1 happy + 1 (extra provider-required item) = 2.

func TestCheckRequestContravariance(t *testing.T) {
	t.Run("happy: consumer sends everything provider requires", func(t *testing.T) {
		consumerOp := &openapi3.Operation{
			Parameters: openapi3.Parameters{
				{Value: &openapi3.Parameter{Name: "api_key", In: "header", Required: true}},
			},
		}
		providerOp := &openapi3.Operation{
			Parameters: openapi3.Parameters{
				{Value: &openapi3.Parameter{Name: "api_key", In: "header", Required: true}},
			},
		}

		findings := checkRequestContravariance(consumerOp, providerOp)
		if len(findings) != 0 {
			t.Fatalf("findings = %v, want none", findings)
		}
	})

	t.Run("provider requires an item the consumer omits", func(t *testing.T) {
		consumerOp := &openapi3.Operation{
			Parameters: openapi3.Parameters{
				{Value: &openapi3.Parameter{Name: "api_key", In: "header", Required: true}},
			},
		}
		providerOp := &openapi3.Operation{
			Parameters: openapi3.Parameters{
				{Value: &openapi3.Parameter{Name: "api_key", In: "header", Required: true}},
				{Value: &openapi3.Parameter{Name: "trace_id", In: "header", Required: true}},
			},
		}

		findings := checkRequestContravariance(consumerOp, providerOp)
		if len(findings) != 1 {
			t.Fatalf("findings len = %d, want 1", len(findings))
		}
		if findings[0].Rule != RuleRequestContravariance {
			t.Fatalf("rule = %q, want %q", findings[0].Rule, RuleRequestContravariance)
		}
		if findings[0].Location != "trace_id" {
			t.Fatalf("location = %q, want %q", findings[0].Location, "trace_id")
		}
	})
}

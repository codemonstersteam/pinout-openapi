package compare

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Unit-test formula: checkStatusCodes: 1 happy + 1 (consumer code ∉ provider codes) = 2.

func responsesWithStatuses(statuses ...string) *openapi3.Responses {
	r := &openapi3.Responses{}
	for _, s := range statuses {
		r.Set(s, &openapi3.ResponseRef{Value: &openapi3.Response{}})
	}
	return r
}

func TestCheckStatusCodes(t *testing.T) {
	t.Run("happy: consumer codes are a subset of provider codes", func(t *testing.T) {
		consumerOp := &openapi3.Operation{Responses: responsesWithStatuses("200")}
		providerOp := &openapi3.Operation{Responses: responsesWithStatuses("200", "404")}

		findings := checkStatusCodes(consumerOp, providerOp)
		if len(findings) != 0 {
			t.Fatalf("findings = %v, want none", findings)
		}
	})

	t.Run("consumer code absent from provider codes", func(t *testing.T) {
		consumerOp := &openapi3.Operation{Responses: responsesWithStatuses("200", "404")}
		providerOp := &openapi3.Operation{Responses: responsesWithStatuses("200")}

		findings := checkStatusCodes(consumerOp, providerOp)
		if len(findings) != 1 {
			t.Fatalf("findings len = %d, want 1", len(findings))
		}
		if findings[0].Rule != RuleStatusCodes {
			t.Fatalf("rule = %q, want %q", findings[0].Rule, RuleStatusCodes)
		}
		if findings[0].Location != "404" {
			t.Fatalf("location = %q, want %q", findings[0].Location, "404")
		}
	})
}

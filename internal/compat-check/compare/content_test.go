package compare

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Unit-test formula: checkContentTypes: 1 happy + 1 (mismatch) = 2.

func opWithRequestContentType(ct string) *openapi3.Operation {
	return &openapi3.Operation{
		RequestBody: &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{
			Content: openapi3.Content{ct: {}},
		}},
	}
}

func TestCheckContentTypes(t *testing.T) {
	t.Run("happy: content-types match", func(t *testing.T) {
		consumerOp := opWithRequestContentType("application/json")
		providerOp := opWithRequestContentType("application/json")

		findings := checkContentTypes(consumerOp, providerOp)
		if len(findings) != 0 {
			t.Fatalf("findings = %v, want none", findings)
		}
	})

	t.Run("content-type mismatch", func(t *testing.T) {
		consumerOp := opWithRequestContentType("application/xml")
		providerOp := opWithRequestContentType("application/json")

		findings := checkContentTypes(consumerOp, providerOp)
		if len(findings) != 1 {
			t.Fatalf("findings len = %d, want 1", len(findings))
		}
		if findings[0].Rule != RuleContentTypes {
			t.Fatalf("rule = %q, want %q", findings[0].Rule, RuleContentTypes)
		}
	})
}

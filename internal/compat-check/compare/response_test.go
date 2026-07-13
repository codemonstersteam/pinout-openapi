package compare

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Unit-test formula: checkResponseCovariance: 1 happy + 1 (consumer reads a field provider
// omits) = 2.

func responseWithFields(fields ...string) *openapi3.ResponseRef {
	props := openapi3.Schemas{}
	for _, f := range fields {
		props[f] = &openapi3.SchemaRef{Value: &openapi3.Schema{}}
	}
	return &openapi3.ResponseRef{Value: &openapi3.Response{
		Content: openapi3.Content{
			"application/json": {Schema: &openapi3.SchemaRef{Value: &openapi3.Schema{Properties: props}}},
		},
	}}
}

func TestCheckResponseCovariance(t *testing.T) {
	t.Run("happy: consumer reads only fields the provider provides", func(t *testing.T) {
		consumerResponses := &openapi3.Responses{}
		consumerResponses.Set("200", responseWithFields("id"))
		providerResponses := &openapi3.Responses{}
		providerResponses.Set("200", responseWithFields("id", "name"))

		consumerOp := &openapi3.Operation{Responses: consumerResponses}
		providerOp := &openapi3.Operation{Responses: providerResponses}

		findings := checkResponseCovariance(consumerOp, providerOp)
		if len(findings) != 0 {
			t.Fatalf("findings = %v, want none", findings)
		}
	})

	t.Run("consumer reads a field the provider omits", func(t *testing.T) {
		consumerResponses := &openapi3.Responses{}
		consumerResponses.Set("200", responseWithFields("id", "email"))
		providerResponses := &openapi3.Responses{}
		providerResponses.Set("200", responseWithFields("id"))

		consumerOp := &openapi3.Operation{Responses: consumerResponses}
		providerOp := &openapi3.Operation{Responses: providerResponses}

		findings := checkResponseCovariance(consumerOp, providerOp)
		if len(findings) != 1 {
			t.Fatalf("findings len = %d, want 1", len(findings))
		}
		if findings[0].Rule != RuleResponseCovariance {
			t.Fatalf("rule = %q, want %q", findings[0].Rule, RuleResponseCovariance)
		}
		if findings[0].Location != "200.email" {
			t.Fatalf("location = %q, want %q", findings[0].Location, "200.email")
		}
	})
}

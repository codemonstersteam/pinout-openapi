package compare

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// Unit-test formula: checkPresence: 1 happy + 1 (absent in provider) = 2.

func TestCheckPresence(t *testing.T) {
	t.Run("happy: provider exposes the operation", func(t *testing.T) {
		consumerOp := &openapi3.Operation{}
		providerOp := &openapi3.Operation{}

		findings := checkPresence(consumerOp, providerOp)
		if len(findings) != 0 {
			t.Fatalf("findings = %v, want none", findings)
		}
	})

	t.Run("operation absent from provider", func(t *testing.T) {
		consumerOp := &openapi3.Operation{}

		findings := checkPresence(consumerOp, nil)
		if len(findings) != 1 {
			t.Fatalf("findings len = %d, want 1", len(findings))
		}
		if findings[0].Rule != RulePresence {
			t.Fatalf("rule = %q, want %q", findings[0].Rule, RulePresence)
		}
	})
}

package compare

import "github.com/getkin/kin-openapi/openapi3"

// checkStatusCodes enforces consumer status codes ⊆ provider status codes: every status the
// consumer expects must be declared by the provider. A missing status is the violation
// (contracts.md "checkStatusCodes").
func checkStatusCodes(consumerOp, providerOp *openapi3.Operation) []Finding {
	if consumerOp == nil || providerOp == nil || consumerOp.Responses == nil {
		return nil
	}

	var findings []Finding
	for _, status := range consumerOp.Responses.Keys() {
		if providerOp.Responses != nil && providerOp.Responses.Value(status) != nil {
			continue
		}
		findings = append(findings, Finding{
			Rule:     RuleStatusCodes,
			Location: status,
			Detail:   "consumer expects status " + status + " the provider does not declare",
		})
	}
	return findings
}

package compare

import "github.com/getkin/kin-openapi/openapi3"

// checkPresence reports whether the provider exposes the same path+method the consumer expects.
// Direction: consumer expectations must be satisfiable by the provider master (contracts.md
// "checkPresence"). providerOp == nil means the operation is absent from the provider spec.
func checkPresence(consumerOp, providerOp *openapi3.Operation) []Finding {
	if providerOp != nil {
		return nil
	}
	return []Finding{{
		Rule:     RulePresence,
		Location: "operation",
		Detail:   "operation is absent from the provider spec",
	}}
}

package compare

import "github.com/getkin/kin-openapi/openapi3"

// checkRequestContravariance enforces provider.required(request) ⊆ consumer.sent(request): every
// request item (parameter or body property) the provider requires must be among what the
// consumer actually sends. An extra provider-required item is the violation (contracts.md
// "checkRequestContravariance").
func checkRequestContravariance(consumerOp, providerOp *openapi3.Operation) []Finding {
	if consumerOp == nil || providerOp == nil {
		return nil
	}
	sent := requestItemsSent(consumerOp)

	var findings []Finding
	for _, item := range requestItemsRequired(providerOp) {
		if sent[item] {
			continue
		}
		findings = append(findings, Finding{
			Rule:     RuleRequestContravariance,
			Location: item,
			Detail:   "provider requires request item \"" + item + "\" the consumer does not send",
		})
	}
	return findings
}

// requestItemsRequired collects the request items (required parameter names + request-body
// schema required property names) op requires.
func requestItemsRequired(op *openapi3.Operation) []string {
	items := requestParameterNames(op, true)
	if schema := requestBodySchema(op); schema != nil {
		items = append(items, schema.Required...)
	}
	return items
}

// requestItemsSent collects every request item (parameter or body property, regardless of
// required-ness) op sends.
func requestItemsSent(op *openapi3.Operation) map[string]bool {
	sent := make(map[string]bool)
	for _, name := range requestParameterNames(op, false) {
		sent[name] = true
	}
	if schema := requestBodySchema(op); schema != nil {
		for _, name := range sortedKeys(schema.Properties) {
			sent[name] = true
		}
	}
	return sent
}

// requestParameterNames returns op's parameter names; onlyRequired restricts to Required==true.
func requestParameterNames(op *openapi3.Operation, onlyRequired bool) []string {
	var names []string
	for _, ref := range op.Parameters {
		if ref == nil || ref.Value == nil {
			continue
		}
		if onlyRequired && !ref.Value.Required {
			continue
		}
		names = append(names, ref.Value.Name)
	}
	return names
}

// requestBodySchema returns the first (sorted by content-type) media type's schema of op's
// request body, or nil when the operation declares no request body.
func requestBodySchema(op *openapi3.Operation) *openapi3.Schema {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil
	}
	content := op.RequestBody.Value.Content
	for _, ct := range sortedKeys(content) {
		mt := content[ct]
		if mt != nil && mt.Schema != nil && mt.Schema.Value != nil {
			return mt.Schema.Value
		}
	}
	return nil
}

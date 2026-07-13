package compare

import "github.com/getkin/kin-openapi/openapi3"

// checkResponseCovariance enforces consumer.read(response) ⊆ provider.provided(response): every
// response field the consumer reads (for a status both sides declare) must be provided by the
// provider. A missing field is the violation (contracts.md "checkResponseCovariance"). A status
// present on the consumer side but absent on the provider side is checkStatusCodes' concern, not
// this rule's — skipped here.
func checkResponseCovariance(consumerOp, providerOp *openapi3.Operation) []Finding {
	if consumerOp == nil || providerOp == nil || consumerOp.Responses == nil || providerOp.Responses == nil {
		return nil
	}

	var findings []Finding
	for _, status := range consumerOp.Responses.Keys() {
		consumerResp := consumerOp.Responses.Value(status)
		providerResp := providerOp.Responses.Value(status)
		if consumerResp == nil || consumerResp.Value == nil || providerResp == nil || providerResp.Value == nil {
			continue
		}

		consumerFields := responseFields(consumerResp.Value)
		providerFields := responseFields(providerResp.Value)
		for _, field := range sortedKeys(consumerFields) {
			if providerFields[field] {
				continue
			}
			findings = append(findings, Finding{
				Rule:     RuleResponseCovariance,
				Location: status + "." + field,
				Detail:   "consumer reads response field \"" + field + "\" for status " + status + " the provider does not provide",
			})
		}
	}
	return findings
}

// responseFields collects the property names of resp's content schemas (across all declared
// content-types).
func responseFields(resp *openapi3.Response) map[string]bool {
	fields := make(map[string]bool)
	for _, ct := range sortedKeys(resp.Content) {
		mt := resp.Content[ct]
		if mt == nil || mt.Schema == nil || mt.Schema.Value == nil {
			continue
		}
		for _, name := range sortedKeys(mt.Schema.Value.Properties) {
			fields[name] = true
		}
	}
	return fields
}

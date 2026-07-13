package compare

import "github.com/getkin/kin-openapi/openapi3"

// checkContentTypes enforces content-type agreement: every content-type the consumer sends
// (request body) or reads (response body, per status both sides declare) must be offered by the
// provider. A mismatch is the violation (contracts.md "checkContentTypes"). A status present on
// the consumer side but absent on the provider side is checkStatusCodes' concern — skipped here.
func checkContentTypes(consumerOp, providerOp *openapi3.Operation) []Finding {
	if consumerOp == nil || providerOp == nil {
		return nil
	}

	findings := contentTypeMismatches(requestContent(consumerOp), requestContent(providerOp), "request")

	if consumerOp.Responses != nil && providerOp.Responses != nil {
		for _, status := range consumerOp.Responses.Keys() {
			consumerResp := consumerOp.Responses.Value(status)
			providerResp := providerOp.Responses.Value(status)
			if consumerResp == nil || consumerResp.Value == nil || providerResp == nil || providerResp.Value == nil {
				continue
			}
			findings = append(findings, contentTypeMismatches(consumerResp.Value.Content, providerResp.Value.Content, "response "+status)...)
		}
	}
	return findings
}

// requestContent returns op's request-body content map, or nil when no request body is declared.
func requestContent(op *openapi3.Operation) openapi3.Content {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return nil
	}
	return op.RequestBody.Value.Content
}

// contentTypeMismatches reports each consumer content-type absent from provider, tagging the
// finding's location with where (e.g. "request", "response 200").
func contentTypeMismatches(consumer, provider openapi3.Content, where string) []Finding {
	var findings []Finding
	for _, ct := range sortedKeys(consumer) {
		if _, ok := provider[ct]; ok {
			continue
		}
		findings = append(findings, Finding{
			Rule:     RuleContentTypes,
			Location: where + ": " + ct,
			Detail:   "consumer expects content-type \"" + ct + "\" for " + where + " the provider does not declare",
		})
	}
	return findings
}

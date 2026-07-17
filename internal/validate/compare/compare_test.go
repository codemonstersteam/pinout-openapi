package compare

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/validate/domain"
)

// Unit-test formula (module-tree.md, row CompareContracts): 1 happy + 2 branches —
// multi-operation fold accumulation, uncovered_operations[] detection.

const compareTestSpecYAML = `
openapi: 3.0.3
info:
  title: test provider
  version: "1"
paths:
  /widgets:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
  /orders:
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [amount]
              properties:
                amount:
                  type: number
      responses:
        "201":
          description: created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
`

func mustLoadCompareSpec(t *testing.T) domain.ProviderSpec {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromData([]byte(compareTestSpecYAML))
	if err != nil {
		t.Fatalf("load test spec: %v", err)
	}
	return domain.ProviderSpec{Doc: doc}
}

// TestCompareContracts_Happy — 1 happy: a single scoped, forward-compatible operation
// folds into a compatible outcome (no violations); Provenance carried through
// unchanged. uncovered_operations[] reports the PROVIDER surface outside the scope —
// the spec exposes /widgets:get (scoped) and /orders:post (not scoped), so the latter
// is surfaced informationally without affecting the compatible verdict.
func TestCompareContracts_Happy(t *testing.T) {
	spec := mustLoadCompareSpec(t)
	ref := domain.OperationRef{Path: "/widgets", Method: "get"}
	provenance := domain.Provenance{Provider: "acme-provider", ProviderVersion: "1.2.3", CapturedHash: "sha256:deadbeef"}

	c := domain.Comparison{
		ScopedOps: []domain.OperationRef{ref},
		Consumed: domain.ConsumedContract{
			Operations: []domain.ConsumedOperation{
				{Ref: ref, Sends: map[string]string{}, Reads: map[string]string{"id": "string"}},
			},
			Provenance: provenance,
		},
		Spec:       spec,
		Provenance: provenance,
	}

	got := CompareContracts(c)

	if len(got.Violations) != 0 {
		t.Fatalf("Violations = %v, want none", got.Violations)
	}
	if len(got.UncoveredOps) != 1 || got.UncoveredOps[0] != "post /orders" {
		t.Fatalf("UncoveredOps = %v, want [\"post /orders\"] (provider op outside scope)", got.UncoveredOps)
	}
	if got.Provenance != provenance {
		t.Errorf("Provenance = %+v, want %+v", got.Provenance, provenance)
	}
}

// TestCompareContracts_MultiOperationFoldAccumulation — branch: violations from
// several scoped operations are concatMap'd (folded) into one Violations slice, not
// dropped/overwritten by later iterations.
func TestCompareContracts_MultiOperationFoldAccumulation(t *testing.T) {
	spec := mustLoadCompareSpec(t)
	widgetsRef := domain.OperationRef{Path: "/widgets", Method: "get"}
	ordersRef := domain.OperationRef{Path: "/orders", Method: "post"}

	c := domain.Comparison{
		ScopedOps: []domain.OperationRef{widgetsRef, ordersRef},
		Consumed: domain.ConsumedContract{
			Operations: []domain.ConsumedOperation{
				// reads a field the provider does not offer -> READS_FIELD_NOT_PROVIDED
				{Ref: widgetsRef, Sends: map[string]string{}, Reads: map[string]string{"name": "string"}},
				// does not send the body field the provider requires -> MISSING_REQUIRED_REQUEST_FIELD
				{Ref: ordersRef, Sends: map[string]string{}, Reads: map[string]string{}},
			},
		},
		Spec: spec,
	}

	got := CompareContracts(c)

	if len(got.Violations) != 2 {
		t.Fatalf("Violations = %v, want 2 (accumulated across both operations)", got.Violations)
	}

	var sawReadsNotProvided, sawMissingRequired bool
	for _, v := range got.Violations {
		switch v.Code {
		case domain.CodeReadsFieldNotProvided:
			sawReadsNotProvided = true
		case domain.CodeMissingRequiredRequestField:
			sawMissingRequired = true
		}
	}
	if !sawReadsNotProvided {
		t.Errorf("missing %s violation from /widgets", domain.CodeReadsFieldNotProvided)
	}
	if !sawMissingRequired {
		t.Errorf("missing %s violation from /orders", domain.CodeMissingRequiredRequestField)
	}
}

// TestCompareContracts_UncoveredOperationsDetection — branch: uncovered_operations[] is
// the PROVIDER surface outside the configured scope, sourced from c.Spec — NOT from the
// consumed contract (the frozen semantics, CONTEXT.md:52 / contracts.md §CompareContracts /
// use-case.md / ticket-13). Discriminator: the consumed contract names ONLY the scoped
// op (nothing out of scope on the consumer side), yet the provider spec exposes
// /orders:post beyond the scope — so a consumer-sourced reading would (wrongly) yield an
// empty list, while the correct provider-sourced reading surfaces "post /orders". The
// out-of-scope provider op is reported informationally and never compared (no violation).
func TestCompareContracts_UncoveredOperationsDetection(t *testing.T) {
	spec := mustLoadCompareSpec(t)
	widgetsRef := domain.OperationRef{Path: "/widgets", Method: "get"}

	c := domain.Comparison{
		ScopedOps: []domain.OperationRef{widgetsRef}, // /orders (in the spec) is NOT scoped
		Consumed: domain.ConsumedContract{
			Operations: []domain.ConsumedOperation{
				// consumer records ONLY the scoped op — nothing out of scope here.
				{Ref: widgetsRef, Sends: map[string]string{}, Reads: map[string]string{"id": "string"}},
			},
		},
		Spec: spec,
	}

	got := CompareContracts(c)

	if len(got.Violations) != 0 {
		t.Fatalf("Violations = %v, want none (out-of-scope provider op must not be compared)", got.Violations)
	}
	if len(got.UncoveredOps) != 1 {
		t.Fatalf("UncoveredOps = %v, want 1 entry (provider /orders:post outside scope)", got.UncoveredOps)
	}
	if want := "post /orders"; got.UncoveredOps[0] != want {
		t.Errorf("UncoveredOps[0] = %q, want %q", got.UncoveredOps[0], want)
	}
}

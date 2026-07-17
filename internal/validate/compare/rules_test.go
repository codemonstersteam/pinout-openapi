package compare

import (
	"testing"

	"pinout-openapi/internal/validate/domain"
)

// Unit-test formula (module-tree.md, row CompareOperation): 1 happy + 5 branches —
// R2 required body field not sent, R2 required param (path/query/header) not sent,
// R3 read field not provided, R4 type mismatch (request), R4 type mismatch (response).
// R1 (OP_NOT_IN_PROVIDER) is DeriveProviderOperation's unit boundary (ticket 10), not
// CompareOperation's (ADR-0002 / module-tree.md "Unit-test formula" note).

func TestCompareOperation_Happy_NoViolations(t *testing.T) {
	consumedOp := domain.ConsumedOperation{
		Ref:   domain.OperationRef{Path: "/orders", Method: "post"},
		Sends: map[string]string{"amount": "number"},
		Reads: map[string]string{"id": "string"},
	}
	providerOp := &domain.ProviderOperation{
		Requires: map[string]string{"amount": "number"},
		Provides: map[string]string{"id": "string"},
	}

	got := CompareOperation(consumedOp, providerOp)

	if len(got) != 0 {
		t.Fatalf("want no violations for a compatible pair, got %v", got)
	}
}

func TestCompareOperation_R2_MissingRequiredBodyField(t *testing.T) {
	consumedOp := domain.ConsumedOperation{
		Ref:   domain.OperationRef{Path: "/orders", Method: "post"},
		Sends: map[string]string{}, // does not send the body field the provider requires
		Reads: map[string]string{},
	}
	providerOp := &domain.ProviderOperation{
		Requires: map[string]string{"amount": "number"}, // required request body field
		Provides: map[string]string{},
	}

	got := CompareOperation(consumedOp, providerOp)

	assertSingleViolation(t, got, domain.CodeMissingRequiredRequestField, "amount")
}

func TestCompareOperation_R2_MissingRequiredParamField(t *testing.T) {
	consumedOp := domain.ConsumedOperation{
		Ref:   domain.OperationRef{Path: "/orders/{id}", Method: "get"},
		Sends: map[string]string{}, // does not send the path/query/header param the provider requires
		Reads: map[string]string{},
	}
	providerOp := &domain.ProviderOperation{
		Requires: map[string]string{"X-Idempotency-Key": "string"}, // required header param
		Provides: map[string]string{},
	}

	got := CompareOperation(consumedOp, providerOp)

	assertSingleViolation(t, got, domain.CodeMissingRequiredRequestField, "X-Idempotency-Key")
}

func TestCompareOperation_R3_ReadFieldNotProvided(t *testing.T) {
	consumedOp := domain.ConsumedOperation{
		Ref:   domain.OperationRef{Path: "/orders/{id}", Method: "get"},
		Sends: map[string]string{},
		Reads: map[string]string{"currency": "string"}, // reads a field the provider no longer provides
	}
	providerOp := &domain.ProviderOperation{
		Requires: map[string]string{},
		Provides: map[string]string{},
	}

	got := CompareOperation(consumedOp, providerOp)

	assertSingleViolation(t, got, domain.CodeReadsFieldNotProvided, "currency")
}

func TestCompareOperation_R4_TypeMismatchRequest(t *testing.T) {
	consumedOp := domain.ConsumedOperation{
		Ref:   domain.OperationRef{Path: "/orders", Method: "post"},
		Sends: map[string]string{"amount": "string"}, // sends a string
		Reads: map[string]string{},
	}
	providerOp := &domain.ProviderOperation{
		Requires: map[string]string{"amount": "number"}, // provider expects a number
		Provides: map[string]string{},
	}

	got := CompareOperation(consumedOp, providerOp)

	assertSingleViolation(t, got, domain.CodeTypeMismatch, "amount")
}

func TestCompareOperation_R4_TypeMismatchResponse(t *testing.T) {
	consumedOp := domain.ConsumedOperation{
		Ref:   domain.OperationRef{Path: "/orders/{id}", Method: "get"},
		Sends: map[string]string{},
		Reads: map[string]string{"balance": "string"}, // reads it as a string
	}
	providerOp := &domain.ProviderOperation{
		Requires: map[string]string{},
		Provides: map[string]string{"balance": "number"}, // provider provides a number
	}

	got := CompareOperation(consumedOp, providerOp)

	assertSingleViolation(t, got, domain.CodeTypeMismatch, "balance")
}

func assertSingleViolation(t *testing.T, got []domain.Violation, wantCode, wantDetails string) {
	t.Helper()

	if len(got) != 1 {
		t.Fatalf("want exactly 1 violation, got %d: %v", len(got), got)
	}
	if got[0].Code != wantCode {
		t.Errorf("want code %q, got %q", wantCode, got[0].Code)
	}
	if got[0].Details != wantDetails {
		t.Errorf("want details %q, got %q", wantDetails, got[0].Details)
	}
}

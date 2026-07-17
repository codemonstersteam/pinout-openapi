package provider

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"

	"pinout-openapi/internal/validate/domain"
)

const testSpecYAML = `
openapi: 3.0.3
info:
  title: test provider
  version: "1"
paths:
  /widgets/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: verbose
          in: query
          required: false
          schema:
            type: boolean
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
                  name:
                    type: string
                  count:
                    type: integer
    post:
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [name]
              properties:
                name:
                  type: string
                note:
                  type: string
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

func mustLoadSpec(t *testing.T) domain.ProviderSpec {
	t.Helper()
	doc, err := openapi3.NewLoader().LoadFromData([]byte(testSpecYAML))
	if err != nil {
		t.Fatalf("load test spec: %v", err)
	}
	return domain.ProviderSpec{Doc: doc}
}

// happy: operation present — Requires derived over body + parameters (path/query/
// header; optional ones excluded), Provides derived from the 2xx response body
// properties (module-tree.md formula: 1 happy unit for DeriveProviderOperation).
func TestDeriveProviderOperation_Happy(t *testing.T) {
	spec := mustLoadSpec(t)

	t.Run("parameters", func(t *testing.T) {
		op, ok := DeriveProviderOperation(spec, domain.OperationRef{Path: "/widgets/{id}", Method: "get"})
		if !ok {
			t.Fatalf("expected operation present, got NotPresent")
		}

		wantRequires := map[string]string{"id": "string"}
		if len(op.Requires) != len(wantRequires) {
			t.Fatalf("Requires = %v, want %v", op.Requires, wantRequires)
		}
		for k, v := range wantRequires {
			if op.Requires[k] != v {
				t.Errorf("Requires[%q] = %q, want %q", k, op.Requires[k], v)
			}
		}

		wantProvides := map[string]string{"id": "string", "name": "string", "count": "integer"}
		if len(op.Provides) != len(wantProvides) {
			t.Fatalf("Provides = %v, want %v", op.Provides, wantProvides)
		}
		for k, v := range wantProvides {
			if op.Provides[k] != v {
				t.Errorf("Provides[%q] = %q, want %q", k, op.Provides[k], v)
			}
		}
	})

	t.Run("body", func(t *testing.T) {
		op, ok := DeriveProviderOperation(spec, domain.OperationRef{Path: "/widgets/{id}", Method: "post"})
		if !ok {
			t.Fatalf("expected operation present, got NotPresent")
		}
		if got, want := op.Requires["name"], "string"; got != want {
			t.Errorf("Requires[name] = %q, want %q", got, want)
		}
		if _, present := op.Requires["note"]; present {
			t.Errorf("Requires must not contain optional body field %q", "note")
		}
		if got, want := op.Provides["id"], "string"; got != want {
			t.Errorf("Provides[id] = %q, want %q", got, want)
		}
	})
}

// branch: operation absent in provider spec -> R1 NotPresent (ok=false), whether the
// path itself is unknown or the path exists but the method is undeclared.
func TestDeriveProviderOperation_NotPresent(t *testing.T) {
	spec := mustLoadSpec(t)

	t.Run("unknown path", func(t *testing.T) {
		op, ok := DeriveProviderOperation(spec, domain.OperationRef{Path: "/does-not-exist", Method: "get"})
		if ok {
			t.Fatalf("expected NotPresent, got ok=true op=%+v", op)
		}
		if len(op.Requires) != 0 || len(op.Provides) != 0 {
			t.Errorf("expected zero-value ProviderOperation on NotPresent, got %+v", op)
		}
	})

	t.Run("known path, undeclared method", func(t *testing.T) {
		_, ok := DeriveProviderOperation(spec, domain.OperationRef{Path: "/widgets/{id}", Method: "delete"})
		if ok {
			t.Fatalf("expected NotPresent for undeclared method, got ok=true")
		}
	})
}

// EnumerateOperations lists every (path+method) in the provider spec, methods
// lower-cased (config convention) and the slice sorted deterministically — the source
// set for uncovered_operations[] (contracts.md §CompareContracts).
func TestEnumerateOperations(t *testing.T) {
	t.Run("all operations, lower-cased and sorted", func(t *testing.T) {
		got := EnumerateOperations(mustLoadSpec(t))
		want := []domain.OperationRef{
			{Path: "/widgets/{id}", Method: "get"},
			{Path: "/widgets/{id}", Method: "post"},
		}
		if len(got) != len(want) {
			t.Fatalf("EnumerateOperations = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("op[%d] = %+v, want %+v", i, got[i], want[i])
			}
		}
	})

	t.Run("empty on nil / non-openapi doc", func(t *testing.T) {
		if got := EnumerateOperations(domain.ProviderSpec{Doc: nil}); len(got) != 0 {
			t.Errorf("nil doc: got %v, want empty", got)
		}
		if got := EnumerateOperations(domain.ProviderSpec{Doc: "not a spec"}); len(got) != 0 {
			t.Errorf("non-openapi doc: got %v, want empty", got)
		}
	})
}

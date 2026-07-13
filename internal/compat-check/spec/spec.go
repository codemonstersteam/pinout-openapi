// Package spec hides how an OpenAPI 3.x document is loaded, $ref-resolved and queried
// (module-tree.md "The secret each module hides"). It honest-reuses kin-openapi for parsing —
// this package never reimplements OpenAPI parsing/validation.
package spec

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Spec wraps a parsed, $ref-resolved kin-openapi document. Read-only query surface — the only
// producers are ConsumerSpecReader.Load and ProviderSpecReader.Load in this package. Types +
// queries only, not unit-tested (thin wrapper over an honest-reused parser — module-tree.md
// node -> file map).
type Spec struct {
	doc *openapi3.T
}

// newSpec wraps an already-loaded, already-validated kin-openapi document. Unexported: a Spec
// can only be produced by this package's two readers.
func newSpec(doc *openapi3.T) Spec {
	return Spec{doc: doc}
}

// HasOperation reports whether path+method (method matched case-insensitively) resolves to an
// operation in this spec.
func (s Spec) HasOperation(path, method string) bool {
	return s.Operation(path, method) != nil
}

// Operation returns the resolved kin-openapi operation for path+method (method matched
// case-insensitively — config.Method is already normalized lower-case, kin-openapi keys are
// upper-case per net/http convention), or nil when path+method is absent from this spec.
func (s Spec) Operation(path, method string) *openapi3.Operation {
	if s.doc == nil || s.doc.Paths == nil {
		return nil
	}
	item := s.doc.Paths.Find(path)
	if item == nil {
		return nil
	}
	return item.GetOperation(strings.ToUpper(method))
}

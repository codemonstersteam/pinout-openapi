package compare

import (
	"errors"

	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// ErrOpNotInConsumer is raised when a configured operation is absent from the consumer spec —
// misconfiguration (UC Ext. 3a), not an incompatibility. error.code OP_NOT_IN_CONSUMER, exit 2
// (contracts.md "Error model").
var ErrOpNotInConsumer = errors.New("operation not present in consumer spec")

// CheckedOps is the consumer pre-check's output: every configured operation confirmed present in
// the consumer spec. Unexported field — only NewCheckedOperations can produce one
// (valid-by-construction, module-tree.md).
type CheckedOps struct {
	ops []config.Operation
}

// Operations returns the confirmed-present operations.
func (c CheckedOps) Operations() []config.Operation { return c.ops }

// NewCheckedOperations confirms every configured path+method exists in the consumer spec before
// any compatibility rule runs (contracts.md "NewCheckedOperations (consumer pre-check)").
func NewCheckedOperations(ops []config.Operation, consumer spec.Spec) (CheckedOps, error) {
	for _, op := range ops {
		if !consumer.HasOperation(op.Path().String(), op.Method().String()) {
			return CheckedOps{}, ErrOpNotInConsumer
		}
	}
	return CheckedOps{ops: ops}, nil
}

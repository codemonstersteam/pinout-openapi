package compare

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// Unit-test formula (module-tree.md "Unit-test formula", ticket-05 "18 by formula"):
// NewCheckedOperations: 1 happy + 1 (op absent from consumer) = 2.

const minimalConsumerDoc = `openapi: 3.0.3
info:
  title: consumer-fixture
  version: "1.0.0"
paths:
  /users:
    get:
      responses:
        "200":
          description: OK
`

func mustLoadConsumerSpec(t *testing.T, doc string) spec.Spec {
	t.Helper()
	path := filepath.Join(t.TempDir(), "consumer.yaml")
	if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	s, err := spec.NewConsumerSpecReader().Load(path)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return s
}

func mustOperation(t *testing.T, rawPath, rawMethod string) config.Operation {
	t.Helper()
	p, err := config.NewOperationPath(rawPath)
	if err != nil {
		t.Fatalf("NewOperationPath: %v", err)
	}
	m, err := config.NewMethod(rawMethod)
	if err != nil {
		t.Fatalf("NewMethod: %v", err)
	}
	op, err := config.NewOperation(p, m)
	if err != nil {
		t.Fatalf("NewOperation: %v", err)
	}
	return op
}

func TestNewCheckedOperations(t *testing.T) {
	t.Run("happy: every configured operation present in consumer spec", func(t *testing.T) {
		consumer := mustLoadConsumerSpec(t, minimalConsumerDoc)
		ops := []config.Operation{mustOperation(t, "/users", "get")}

		checked, err := NewCheckedOperations(ops, consumer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(checked.Operations()) != 1 {
			t.Fatalf("Operations() len = %d, want 1", len(checked.Operations()))
		}
	})

	t.Run("op absent from consumer spec rejected", func(t *testing.T) {
		consumer := mustLoadConsumerSpec(t, minimalConsumerDoc)
		ops := []config.Operation{mustOperation(t, "/missing", "post")}

		_, err := NewCheckedOperations(ops, consumer)
		if !errors.Is(err, ErrOpNotInConsumer) {
			t.Fatalf("err = %v, want ErrOpNotInConsumer", err)
		}
	})
}

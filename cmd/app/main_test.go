// Wiring smoke test (ticket 17): the stable `version` command only — the endpoint's own
// domain behavior (validate <config.yaml>) is proven by component scenarios
// (component-tests/features/validate.feature), not a unit test (program-implementation:
// the head/adapter/main are never unit-tested).
package main

import (
	"bytes"
	"strings"
	"testing"
)

// version — стабильная команда, exit 0 + непустой вывод. Дублирует контракт smoke.
func TestVersionCmd(t *testing.T) {
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "version") {
		t.Fatalf("ожидали строку версии, получили %q", out.String())
	}
}

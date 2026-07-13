package main

import (
	"bytes"
	"errors"
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

// run <config> — реальная труба (ticket-09): для несуществующего пути ConfigReader.Read
// заворачивает ErrConfigInvalid → error.code CONFIG_INVALID → cli.Exit(ExitConfigOrUsage=2)
// (api-specification/exit-codes.md). Заменяет плейсхолдер-тест NOT_IMPLEMENTED/exit 3,
// поскольку internal/example и его NOT_IMPLEMENTED-труба этим тикетом удалены. cli.Exit
// (ticket-08) пишет error.code-диагностику напрямую в os.Stderr, а не в cmd.ErrOrStderr() —
// здесь проверяется только exit-код, контракт диагностики уже покрыт компонентными сценариями.
func TestRunCmd_ConfigUnreadable(t *testing.T) {
	root := newRootCmd()
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs([]string{"run", "cfg.yaml"})
	err := root.Execute()
	if err == nil {
		t.Fatal("ожидали exitError, получили nil")
	}
	var ee *exitError
	if !errors.As(err, &ee) || ee.code != 2 {
		t.Fatalf("ожидали exit code 2, получили %v", err)
	}
}

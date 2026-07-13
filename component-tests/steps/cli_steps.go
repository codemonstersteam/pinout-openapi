package steps

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cucumber/godog"
)

// registerCLISteps — универсальные CLI-степы (чёрный ящик): запуск бинаря с
// аргументами и проверки кода выхода / вывода. Достаточно для smoke и большинства
// компонентных сценариев CLI. Доменные степы кладут рядом в новых *_steps.go.
func (w *World) registerCLISteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^запускаем бинарь с аргументами "([^"]*)"$`, w.runBinary)
	ctx.Step(`^запускаем бинарь без аргументов$`, w.runBinaryNoArgs)
	ctx.Step(`^код выхода (\d+)$`, w.exitCode)
	ctx.Step(`^вывод не пустой$`, w.outputNonEmpty)
	ctx.Step(`^вывод содержит "([^"]*)"$`, w.outputContains)
}

func (w *World) runBinaryNoArgs() error { return w.exec(nil) }

func (w *World) runBinary(argline string) error {
	return w.exec(strings.Fields(argline))
}

func (w *World) exec(args []string) error {
	c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(c, w.toolBin, args...)
	out, err := cmd.CombinedOutput()
	w.lastOut = string(out)
	w.ran = true
	var ee *exec.ExitError
	switch {
	case err == nil:
		w.lastExit = 0
	case errors.As(err, &ee):
		w.lastExit = ee.ExitCode()
	default:
		return fmt.Errorf("запуск %s: %w", w.toolBin, err)
	}
	return nil
}

func (w *World) exitCode(expected int) error {
	if !w.ran {
		return fmt.Errorf("бинарь ещё не запускали")
	}
	if w.lastExit != expected {
		return fmt.Errorf("ожидали код выхода %d, получили %d (вывод: %s)", expected, w.lastExit, w.lastOut)
	}
	return nil
}

func (w *World) outputNonEmpty() error {
	if !w.ran {
		return fmt.Errorf("бинарь ещё не запускали")
	}
	if strings.TrimSpace(w.lastOut) == "" {
		return fmt.Errorf("ожидали непустой вывод, получили пустой")
	}
	return nil
}

func (w *World) outputContains(sub string) error {
	if !strings.Contains(w.lastOut, sub) {
		return fmt.Errorf("ожидали подстроку %q в выводе, получили: %s", sub, w.lastOut)
	}
	return nil
}

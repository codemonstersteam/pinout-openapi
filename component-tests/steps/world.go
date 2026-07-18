// Package steps — godog-степы компонентных тестов CLI (чёрный ящик: запуск бинаря,
// проверка кода выхода и вывода). World — состояние сценария, создаётся заново на
// каждый Scenario. Доменных зависимостей нет — добавляй фикстуры/степы в *_steps.go.
package steps

import (
	"context"
	"os"

	"github.com/cucumber/godog"
)

type World struct {
	toolBin    string // путь к застейдженному бинарю (TOOL_BIN)
	lastExit   int    // код выхода последнего запуска
	lastOut    string // stdout+stderr последнего запуска
	ran        bool
	configPath string // путь к fixture-конфигу для доменных (validate) сценариев
}

func newWorld() *World {
	return &World{
		toolBin: getenv("TOOL_BIN", "/bin-share/tool"),
	}
}

func (w *World) resetState() {
	w.lastExit = 0
	w.lastOut = ""
	w.ran = false
	w.configPath = ""
}

func (w *World) beforeScenario(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	w.resetState()
	return ctx, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

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
	toolBin  string // путь к застейдженному бинарю (TOOL_BIN)
	lastExit int    // код выхода последнего запуска
	lastOut  string // stdout+stderr последнего запуска
	ran      bool

	// Домен compat-check (compat_check_steps.go): состояние выбранной фикстуры.
	fixtureDir             string // /fixtures/<name>, выставляет useFixture
	configPath             string // /fixtures/<name>/contract-tests.yaml
	providerSpecPath       string // локальный provider.yaml фикстуры, если есть (для "не изменилась")
	providerSpecHashBefore string // sha256 provider.yaml на момент выбора фикстуры
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
	w.fixtureDir = ""
	w.configPath = ""
	w.providerSpecPath = ""
	w.providerSpecHashBefore = ""
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

package steps

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output:    colors.Colored(os.Stdout),
	Format:    "pretty",
	Randomize: -1, // детерминированный порядок
}

func init() { godog.BindCommandLineFlags("godog.", &opts) }

// TestFeatures гоняет все .feature из ../features. Раннер-контейнер запускает
// `go test -v -count=1 ./steps/...` после того, как бинарь застейджен в общий том.
func TestFeatures(t *testing.T) {
	opts.Paths = []string{"../features"}
	status := godog.TestSuite{
		Name:                 "component-tests",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
		Options:              &opts,
	}.Run()
	if status != 0 {
		t.Fail()
	}
}

// InitializeTestSuite — обвязка (не бизнес-степ): ждёт, пока real-protocol
// HTTP-стаб provider-stub (CS6/CS7/CS8) поднимется и начнёт слушать порт, ПЕРЕД
// первым сценарием. Это намеренно НЕ compose healthcheck/depends_on: gating
// tester через `condition: service_healthy` добавляет задержку старта, которая
// гонится с `--abort-on-container-exit` (tool — one-shot и выходит почти сразу
// после старта) и валит весь стек по wiring-причине ДО того, как tester вообще
// запустится — ложный RED, не бизнес-причина. Поэтому готовность ждёт сам
// Go-процесс тестов, а не docker compose.
func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		if err := waitForProviderStub(); err != nil {
			panic(fmt.Errorf("обвязка: %w", err))
		}
		// Провизионим директорию happy-path отчёта (CS1 пишет
		// /tmp/reports/cs1-report.json). Инструмент НЕ создаёт родительские каталоги
		// сам — это контракт CS9 (запись в несуществующий каталог -> REPORT_WRITE_ERROR,
		// exit 3), поэтому writable-локацию для позитивного сценария готовит обвязка,
		// а не SUT. /no-such-dir (CS9) намеренно НЕ создаётся.
		if err := os.MkdirAll("/tmp/reports", 0o755); err != nil {
			panic(fmt.Errorf("обвязка: не удалось создать /tmp/reports: %w", err))
		}
	})
}

func waitForProviderStub() error {
	url := getenv("PROVIDER_STUB_HEALTH_URL", "http://provider-stub:8080/healthz")
	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(15 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		lastErr = err
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("provider-stub не ответил за отведённое время на %s: %w", url, lastErr)
}

// InitializeScenario регистрирует степы и lifecycle-хуки. Добавляй свои
// register*Steps(ctx) по мере роста набора.
func InitializeScenario(ctx *godog.ScenarioContext) {
	w := newWorld()
	ctx.Before(w.beforeScenario)
	w.registerCLISteps(ctx)
	w.registerCompatCheckSteps(ctx)
}

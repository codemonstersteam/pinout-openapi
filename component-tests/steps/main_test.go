package steps

import (
	"os"
	"testing"

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
		Name:                "component-tests",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()
	if status != 0 {
		t.Fail()
	}
}

// InitializeScenario регистрирует степы и lifecycle-хуки. Добавляй свои
// register*Steps(ctx) по мере роста набора.
func InitializeScenario(ctx *godog.ScenarioContext) {
	w := newWorld()
	ctx.Before(w.beforeScenario)
	w.registerCLISteps(ctx)
	w.registerValidateSteps(ctx)
}

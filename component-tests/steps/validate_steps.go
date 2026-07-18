package steps

import (
	"encoding/json"
	"fmt"

	"github.com/cucumber/godog"
)

// fixturesDir — per-slice fixture root, bind-mounted read-only into the tester container
// (docker-compose.test.yml: ./fixtures:/fixtures:ro). Isolated per slice: only
// component-tests/fixtures/validate/** — see reference.md "Fixtures — isolated per slice".
const fixturesDir = "/fixtures/validate"

// validateReport — the wire shape of api-specification/report.schema.json, just enough of
// it for black-box assertions (exit code + stdout JSON). Mirrors the frozen contract; not a
// copy of any internal type (component tests know nothing of internals).
type validateReport struct {
	SchemaVersion string `json:"schema_version"`
	Compatible    bool   `json:"compatible"`
	Provenance    struct {
		Provider        string `json:"provider"`
		ProviderVersion string `json:"provider_version"`
		CapturedHash    string `json:"captured_hash"`
	} `json:"provenance"`
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"errors"`
	UncoveredOperations []string `json:"uncovered_operations"`
}

// registerValidateSteps — step-defs for slice-01-validate's component scenarios
// (component-tests/features/validate.feature), realized 1:1 from contracts.md §"Component
// scenarios (DESIGN half)". Mechanical glue only (subprocess + JSON-of-stdout assertions);
// no domain logic here — the black box is the staged binary, not this package.
func (w *World) registerValidateSteps(ctx *godog.ScenarioContext) {
	// Given — select the per-scenario fixture (good-validate / bad-validate-<branch>).
	ctx.Step(`^a config whose consumed-contract and provider spec are reachable and compatible$`, w.givenGoodConfig)
	ctx.Step(`^a config file that cannot be read or is schema-invalid$`, w.givenBadConfigFixture)
	ctx.Step(`^a config pointing at a consumed-contract path that does not exist$`, w.givenFileNotFoundFixture)
	ctx.Step(`^a config pointing at a provider spec that is not valid OpenAPI$`, w.givenParseErrorFixture)
	ctx.Step(`^a config with spec_url pointing at a stub that returns 503$`, w.givenHTTPErrorFixture)
	ctx.Step(`^a config with spec_url pointing at a stub that stalls past settings\.timeout$`, w.givenTimeoutFixture)

	// When — run the staged binary against the selected config fixture.
	ctx.Step("^I run `pinout-openapi validate config\\.yaml`$", w.runValidate)

	// Then — exit code (English phrasing alongside cli_steps.go's Russian one; same method).
	ctx.Step(`^the exit code is (\d+)$`, w.exitCode)

	// Then — stdout report assertions.
	ctx.Step(`^stdout is a schema-valid report with compatible=true and errors==\[\]$`, w.stdoutCompatibleNoErrors)
	ctx.Step(`^uncovered provider operations are listed in uncovered_operations\[\]$`, w.stdoutHasUncoveredOperations)
	ctx.Step(`^stdout report errors\[0\]\.code == "([^"]*)"$`, w.stdoutFirstErrorCode)
}

func (w *World) givenGoodConfig()          { w.configPath = fixturesDir + "/good/config.yaml" }
func (w *World) givenBadConfigFixture()    { w.configPath = fixturesDir + "/bad-config/config.yaml" }
func (w *World) givenFileNotFoundFixture() { w.configPath = fixturesDir + "/bad-file-not-found/config.yaml" }
func (w *World) givenParseErrorFixture()   { w.configPath = fixturesDir + "/bad-parse-error/config.yaml" }
func (w *World) givenHTTPErrorFixture()    { w.configPath = fixturesDir + "/bad-http-error/config.yaml" }
func (w *World) givenTimeoutFixture()      { w.configPath = fixturesDir + "/bad-timeout/config.yaml" }

// runValidate invokes the staged binary as `<tool> validate <configPath>` — the black-box
// equivalent of the Cockburn `pinout-openapi validate <config.yaml>` invocation.
func (w *World) runValidate() error {
	if w.configPath == "" {
		return fmt.Errorf("no config fixture selected — a Given step must run first")
	}
	return w.exec([]string{"validate", w.configPath})
}

// parseReport decodes stdout+stderr (CombinedOutput, per exec in cli_steps.go) as the
// report JSON. A decode failure is a legitimate scenario failure — RED by business reason
// while the slice is unimplemented (placeholder/absent module prints no report), not a
// harness bug.
func (w *World) parseReport() (validateReport, error) {
	var rep validateReport
	if err := json.Unmarshal([]byte(w.lastOut), &rep); err != nil {
		return rep, fmt.Errorf("stdout is not a valid JSON report (%v); raw output: %s", err, w.lastOut)
	}
	return rep, nil
}

func (w *World) stdoutCompatibleNoErrors() error {
	rep, err := w.parseReport()
	if err != nil {
		return err
	}
	if !rep.Compatible {
		return fmt.Errorf("expected report.compatible=true, got false (errors=%+v)", rep.Errors)
	}
	if len(rep.Errors) != 0 {
		return fmt.Errorf("expected report.errors==[], got %+v", rep.Errors)
	}
	return nil
}

func (w *World) stdoutHasUncoveredOperations() error {
	rep, err := w.parseReport()
	if err != nil {
		return err
	}
	if len(rep.UncoveredOperations) == 0 {
		return fmt.Errorf("expected a non-empty report.uncovered_operations[], got %+v", rep.UncoveredOperations)
	}
	return nil
}

func (w *World) stdoutFirstErrorCode(wantCode string) error {
	rep, err := w.parseReport()
	if err != nil {
		return err
	}
	if len(rep.Errors) == 0 {
		return fmt.Errorf("expected report.errors[0].code=%s, got empty errors[]", wantCode)
	}
	if got := rep.Errors[0].Code; got != wantCode {
		return fmt.Errorf("expected report.errors[0].code=%s, got %s", wantCode, got)
	}
	return nil
}

package steps

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/cucumber/godog"
)

// fixturesDir — per-slice fixture root, bind-mounted read-only into the tester container
// (docker-compose.test.yml: ./fixtures:/fixtures:ro). Isolated per slice: only
// component-tests/fixtures/validate/** — see reference.md "Fixtures — isolated per slice".
const fixturesDir = "/fixtures/validate"

// validateReport — the wire shape of api-specification/report.schema.json (canon 1.1), just
// enough of it for black-box assertions (exit code + stdout JSON). Mirrors the frozen contract;
// not a copy of any internal type (component tests know nothing of internals).
type validateReport struct {
	SchemaVersion string `json:"schema_version"`
	Validator     string `json:"validator"`
	Interaction   string `json:"interaction"`
	Consumer      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"consumer"`
	GeneratedAt string `json:"generated_at"`
	Compatible  bool   `json:"compatible"`
	Provenance  struct {
		Provider        string `json:"provider"`
		ProviderVersion string `json:"provider_version"`
		CapturedHash    string `json:"captured_hash"`
	} `json:"provenance"`
	Errors []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Subject string `json:"subject"`
	} `json:"errors"`
	UncoveredOperations []string `json:"uncovered_operations"`
}

// generatedAtPattern — canon 1.1: RFC 3339, UTC (Z), second precision, no fractional
// seconds, no offset (docs/report-format.md §1).
var generatedAtPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)

// registerValidateSteps — step-defs for slice-01-validate's component scenarios
// (component-tests/features/validate.feature), realized 1:1 from contracts.md §"Component
// scenarios (DESIGN half)". Mechanical glue only (subprocess + JSON-of-stdout assertions);
// no domain logic here — the black box is the staged binary, not this package.
func (w *World) registerValidateSteps(ctx *godog.ScenarioContext) {
	// Given — select the per-scenario fixture (good-validate / bad-validate-<branch>).
	ctx.Step(`^a config whose consumed-contract and provider spec are reachable and compatible$`, w.givenGoodConfig)
	ctx.Step(`^a config whose scope includes an operation the provider does not expose$`, w.givenIncompatibleFixture)
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
	ctx.Step(`^stdout is a canon 1\.1 report \(schema_version, validator, interaction, consumer\.name, generated_at\)$`, w.stdoutCanon11)
	ctx.Step(`^uncovered provider operations are listed in uncovered_operations\[\]$`, w.stdoutHasUncoveredOperations)
	ctx.Step(`^stdout report errors\[0\]\.code == "([^"]*)"$`, w.stdoutFirstErrorCode)
	ctx.Step(`^stdout report compatible=false with errors\[0\]\.subject naming the operation$`, w.stdoutIncompatibleWithSubject)
}

func (w *World) givenGoodConfig()          { w.configPath = fixturesDir + "/good/config.yaml" }
func (w *World) givenIncompatibleFixture() { w.configPath = fixturesDir + "/incompatible/config.yaml" }
func (w *World) givenBadConfigFixture()    { w.configPath = fixturesDir + "/bad-config/config.yaml" }
func (w *World) givenFileNotFoundFixture() {
	w.configPath = fixturesDir + "/bad-file-not-found/config.yaml"
}
func (w *World) givenParseErrorFixture() { w.configPath = fixturesDir + "/bad-parse-error/config.yaml" }
func (w *World) givenHTTPErrorFixture()  { w.configPath = fixturesDir + "/bad-http-error/config.yaml" }
func (w *World) givenTimeoutFixture()    { w.configPath = fixturesDir + "/bad-timeout/config.yaml" }

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

// stdoutCanon11 — canon 1.1 identity assertions (change 001-report-schema-1.1):
// schema_version, build-time constants validator/interaction, non-empty consumer.name
// (from the config), generated_at in RFC3339 UTC Z second precision.
func (w *World) stdoutCanon11() error {
	rep, err := w.parseReport()
	if err != nil {
		return err
	}
	if rep.SchemaVersion != "1.1" {
		return fmt.Errorf("expected report.schema_version=1.1, got %q", rep.SchemaVersion)
	}
	if rep.Validator != "pinout-openapi" {
		return fmt.Errorf("expected report.validator=pinout-openapi, got %q", rep.Validator)
	}
	if rep.Interaction != "sync" {
		return fmt.Errorf("expected report.interaction=sync, got %q", rep.Interaction)
	}
	if rep.Consumer.Name == "" {
		return fmt.Errorf("expected non-empty report.consumer.name (config.consumer.name), got empty")
	}
	if !generatedAtPattern.MatchString(rep.GeneratedAt) {
		return fmt.Errorf("expected report.generated_at RFC3339 UTC Z (second precision), got %q", rep.GeneratedAt)
	}
	return nil
}

// stdoutIncompatibleWithSubject — the incompatible verdict on the machine channel:
// compatible=false and errors[0].subject non-empty (canon 1.1: the netlist edge key,
// `METHOD /path`, uppercase method).
func (w *World) stdoutIncompatibleWithSubject() error {
	rep, err := w.parseReport()
	if err != nil {
		return err
	}
	if rep.Compatible {
		return fmt.Errorf("expected report.compatible=false, got true")
	}
	if len(rep.Errors) == 0 {
		return fmt.Errorf("expected non-empty report.errors[], got empty")
	}
	subject := rep.Errors[0].Subject
	if subject == "" {
		return fmt.Errorf("expected non-empty report.errors[0].subject, got empty")
	}
	if !subjectPattern.MatchString(subject) {
		return fmt.Errorf("expected report.errors[0].subject as `METHOD /path` (uppercase method), got %q", subject)
	}
	return nil
}

// subjectPattern — `METHOD /path`: uppercase HTTP method, one space, path starting with /.
var subjectPattern = regexp.MustCompile(`^[A-Z]+ /\S.*$`)

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

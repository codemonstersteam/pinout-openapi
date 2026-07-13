// Package config hides the contract-tests.yaml format and how it is read + validated. It is the
// slice's one boundary where an illegal config state can be rejected — every exported constructor
// range-checks its field and unexported struct fields make a naked literal outside this package
// impossible (valid-by-construction, module-tree.md "Valid-by-construction").
package config

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// ErrConfigInvalid is the sentinel raised by NewConfig (unparseable YAML / any child VO
// rejection) and by ConfigReader.Read (missing/unreadable file) — both map to error.code
// CONFIG_INVALID, exit 2 (contracts.md "Error model").
var ErrConfigInvalid = errors.New("config invalid")

var validMethods = map[string]bool{
	"get": true, "put": true, "post": true, "delete": true,
	"options": true, "head": true, "patch": true, "trace": true,
}

var validLogLevels = map[string]bool{
	"debug": true, "info": true, "warn": true, "error": true,
}

const (
	defaultLogLevel       = "info"
	defaultSaveJSONReport = true
	defaultJSONReportFile = "compatibility_report.json"
	defaultTimeout        = 30
)

// NewMethod validates and normalizes an OpenAPI HTTP verb. Case-insensitive on input, normalized
// to lower-case (Postel's law per config.schema.json).
func NewMethod(raw string) (Method, error) {
	lower := strings.ToLower(raw)
	if !validMethods[lower] {
		return Method{}, ErrConfigInvalid
	}
	return Method{verb: lower}, nil
}

// NewOperationPath validates a non-empty OpenAPI path key with a leading '/'.
func NewOperationPath(raw string) (OperationPath, error) {
	if !strings.HasPrefix(raw, "/") {
		return OperationPath{}, ErrConfigInvalid
	}
	return OperationPath{path: raw}, nil
}

// NewOperation composes two already-validated value-objects. It cannot fail — validity was
// established by NewOperationPath/NewMethod — but returns an error to keep the uniform
// NewX(...) -> (X, error) constructor shape used across this package.
func NewOperation(path OperationPath, method Method) (Operation, error) {
	return Operation{path: path, method: method}, nil
}

// NewProviderSource validates that exactly one of url/path is present and resolves the kind.
func NewProviderSource(url, path string) (ProviderSource, error) {
	hasURL := strings.TrimSpace(url) != ""
	hasPath := strings.TrimSpace(path) != ""

	switch {
	case hasURL && hasPath:
		return ProviderSource{}, ErrConfigInvalid
	case !hasURL && !hasPath:
		return ProviderSource{}, ErrConfigInvalid
	case hasURL:
		return ProviderSource{kind: ProviderSourceHTTP, url: url}, nil
	default:
		return ProviderSource{kind: ProviderSourceFile, path: path}, nil
	}
}

// NewSettings applies committed defaults and range-checks the optional settings block.
// rawSaveJSONReport/rawTimeout are pointers so "omitted" (nil, defaulted) is distinguishable
// from an explicit zero value (false / not applicable here since 0 is out of [1,600] range).
func NewSettings(rawLogLevel string, rawSaveJSONReport *bool, rawJSONReportFile string, rawTimeout *int) (Settings, error) {
	logLevel := rawLogLevel
	if logLevel == "" {
		logLevel = defaultLogLevel
	}
	if !validLogLevels[logLevel] {
		return Settings{}, ErrConfigInvalid
	}

	saveJSONReport := defaultSaveJSONReport
	if rawSaveJSONReport != nil {
		saveJSONReport = *rawSaveJSONReport
	}

	timeout := defaultTimeout
	if rawTimeout != nil {
		timeout = *rawTimeout
	}
	if timeout < 1 || timeout > 600 {
		return Settings{}, ErrConfigInvalid
	}

	// json_report_file has a committed default (config.schema.json: "default":
	// "compatibility_report.json"; "the app substitutes the default only as a documented
	// fallback"). An omitted/blank value is NOT a config error — the default is applied so the
	// save branch always has a target. Only save_json_report (default true) governs whether the
	// file is written; the path itself is never a rejection cause.
	reportFile := rawJSONReportFile
	if strings.TrimSpace(reportFile) == "" {
		reportFile = defaultJSONReportFile
	}

	return Settings{
		logLevel:       logLevel,
		saveJSONReport: saveJSONReport,
		jsonReportFile: reportFile,
		timeout:        timeout,
	}, nil
}

// NewConsumer validates the consumer's name and that at least one operation is configured.
// specPath is carried through unvalidated here (config.schema.json's minLength:1 on spec_path
// is enforced by the "required" YAML shape check in NewConfig, not by this constructor — see
// module-tree.md unit-test formula: NewConsumer's only distinguishable branches are name/operations).
func NewConsumer(name string, specPath string, operations []Operation) (Consumer, error) {
	if strings.TrimSpace(name) == "" {
		return Consumer{}, ErrConfigInvalid
	}
	if len(operations) == 0 {
		return Consumer{}, ErrConfigInvalid
	}
	return Consumer{name: name, specPath: specPath, operations: operations}, nil
}

// NewProvider validates the provider's name.
func NewProvider(name string, source ProviderSource) (Provider, error) {
	if strings.TrimSpace(name) == "" {
		return Provider{}, ErrConfigInvalid
	}
	return Provider{name: name, source: source}, nil
}

// rawYAML mirrors config.schema.json's shape for gopkg.in/yaml.v3 unmarshalling.
type rawYAML struct {
	Consumer *rawConsumerYAML `yaml:"consumer"`
	Provider *rawProviderYAML `yaml:"provider"`
	Settings *rawSettingsYAML `yaml:"settings"`
}

type rawConsumerYAML struct {
	SpecPath   string             `yaml:"spec_path"`
	Name       string             `yaml:"name"`
	Operations []rawOperationYAML `yaml:"operations"`
}

type rawOperationYAML struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

type rawProviderYAML struct {
	SpecURL  string `yaml:"spec_url"`
	SpecPath string `yaml:"spec_path"`
	Name     string `yaml:"name"`
}

type rawSettingsYAML struct {
	LogLevel       string `yaml:"log_level"`
	SaveJSONReport *bool  `yaml:"save_json_report"`
	JSONReportFile string `yaml:"json_report_file"`
	Timeout        *int   `yaml:"timeout"`
}

// NewConfig unmarshals raw YAML and assembles the validated value-objects into a Config.
// Any unparseable YAML or child VO rejection collapses to ErrConfigInvalid — the caller (the
// cli/ door, via head.go) never sees which field failed, only that the config is invalid.
func NewConfig(raw RawConfig) (Config, error) {
	var y rawYAML
	if err := yaml.Unmarshal(raw, &y); err != nil {
		return Config{}, ErrConfigInvalid
	}
	if y.Consumer == nil || y.Provider == nil {
		return Config{}, ErrConfigInvalid
	}

	ops := make([]Operation, 0, len(y.Consumer.Operations))
	for _, rawOp := range y.Consumer.Operations {
		path, err := NewOperationPath(rawOp.Path)
		if err != nil {
			return Config{}, ErrConfigInvalid
		}
		method, err := NewMethod(rawOp.Method)
		if err != nil {
			return Config{}, ErrConfigInvalid
		}
		op, err := NewOperation(path, method)
		if err != nil {
			return Config{}, ErrConfigInvalid
		}
		ops = append(ops, op)
	}

	consumer, err := NewConsumer(y.Consumer.Name, y.Consumer.SpecPath, ops)
	if err != nil {
		return Config{}, ErrConfigInvalid
	}

	source, err := NewProviderSource(y.Provider.SpecURL, y.Provider.SpecPath)
	if err != nil {
		return Config{}, ErrConfigInvalid
	}
	provider, err := NewProvider(y.Provider.Name, source)
	if err != nil {
		return Config{}, ErrConfigInvalid
	}

	var rawLogLevel, rawJSONReportFile string
	var rawSaveJSONReport *bool
	var rawTimeout *int
	if y.Settings != nil {
		rawLogLevel = y.Settings.LogLevel
		rawSaveJSONReport = y.Settings.SaveJSONReport
		rawJSONReportFile = y.Settings.JSONReportFile
		rawTimeout = y.Settings.Timeout
	}
	settings, err := NewSettings(rawLogLevel, rawSaveJSONReport, rawJSONReportFile, rawTimeout)
	if err != nil {
		return Config{}, ErrConfigInvalid
	}

	return Config{consumer: consumer, provider: provider, settings: settings}, nil
}

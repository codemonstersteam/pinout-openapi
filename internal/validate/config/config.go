// Package config — подпакет слайса slice-01-validate: секрет "формат конфиг-файла
// (YAML) и правила валидности конфига (schema invariants)" (module-tree.md). Этот
// файл — NewConfig (тикет 07): valid-by-construction конструктор Config из
// RawConfig. ConfigStore.Load (тикет 06, store.go, тот же пакет) — отдельный
// I/O-модуль, сюда не входит.
package config

import (
	"fmt"
	"strings"

	"pinout-openapi/internal/validate/domain"
)

// validMethods — enum config.schema.json#/.../operations/items/properties/method.
var validMethods = map[string]bool{
	"get":     true,
	"post":    true,
	"put":     true,
	"patch":   true,
	"delete":  true,
	"head":    true,
	"options": true,
}

// validLogLevels — enum config.schema.json#/.../settings/properties/log_level.
var validLogLevels = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
}

// Дефолты settings (config.schema.json#/.../settings, поле "default").
const (
	defaultLogLevel       = "info"
	defaultSaveJSONReport = true
	defaultJSONReportFile = "compatibility_report.json"
	defaultTimeout        = 30
)

// NewConfig — valid-by-construction конструктор (contracts.md §NewConfig): проверяет
// каждое поле RawConfig против инвариантов config.schema.json; невалидные состояния
// невозможны за этой границей (неэкспортируемые поля Config). Deps = —.
//
// Antecedent: RawConfig раскодирован. Consequent: Ok Config (непустые имена;
// operations ≥ 1, каждый path соответствует ^/ и method ∈ enum; provider —
// ровно один источник spec_path/spec_url; settings продефолчены и в диапазоне) либо
// Err ErrConfig.
func NewConfig(raw domain.RawConfig) (domain.Config, error) {
	if raw.Consumer.Name == "" {
		return domain.Config{}, fmt.Errorf("%w: consumer.name must not be empty", domain.ErrConfig)
	}
	if raw.Provider.Name == "" {
		return domain.Config{}, fmt.Errorf("%w: provider.name must not be empty", domain.ErrConfig)
	}

	if len(raw.Consumer.Operations) == 0 {
		return domain.Config{}, fmt.Errorf("%w: consumer.operations must have at least 1 item", domain.ErrConfig)
	}
	for _, op := range raw.Consumer.Operations {
		if !strings.HasPrefix(op.Path, "/") {
			return domain.Config{}, fmt.Errorf("%w: operation path %q must match ^/", domain.ErrConfig, op.Path)
		}
		if !validMethods[op.Method] {
			return domain.Config{}, fmt.Errorf("%w: operation method %q is not a valid HTTP method", domain.ErrConfig, op.Method)
		}
	}

	hasPath := raw.Provider.SpecPath != ""
	hasURL := raw.Provider.SpecURL != ""
	switch {
	case hasPath && hasURL:
		return domain.Config{}, fmt.Errorf("%w: provider must set exactly one of spec_path/spec_url, both are set", domain.ErrConfig)
	case !hasPath && !hasURL:
		return domain.Config{}, fmt.Errorf("%w: provider must set exactly one of spec_path/spec_url, neither is set", domain.ErrConfig)
	}

	settings, err := newSettings(raw)
	if err != nil {
		return domain.Config{}, err
	}

	return domain.Config{
		ConsumerName:         raw.Consumer.Name,
		ConsumedContractPath: raw.Consumer.ConsumedContractPath,
		Operations:           raw.Consumer.Operations,
		Provider:             raw.Provider,
		Settings:             settings,
	}, nil
}

// newSettings — дефолты + range-проверка Settings (config.schema.json#/.../settings).
func newSettings(raw domain.RawConfig) (domain.Settings, error) {
	logLevel := defaultLogLevel
	if raw.Settings.LogLevel != nil {
		logLevel = *raw.Settings.LogLevel
		if !validLogLevels[logLevel] {
			return domain.Settings{}, fmt.Errorf("%w: settings.log_level %q is not a valid log level", domain.ErrConfig, logLevel)
		}
	}

	saveJSONReport := defaultSaveJSONReport
	if raw.Settings.SaveJSONReport != nil {
		saveJSONReport = *raw.Settings.SaveJSONReport
	}

	jsonReportFile := defaultJSONReportFile
	if raw.Settings.JSONReportFile != nil {
		jsonReportFile = *raw.Settings.JSONReportFile
	}

	timeout := defaultTimeout
	if raw.Settings.Timeout != nil {
		timeout = *raw.Settings.Timeout
		if timeout <= 0 {
			return domain.Settings{}, fmt.Errorf("%w: settings.timeout must be > 0, got %d", domain.ErrConfig, timeout)
		}
	}

	ignoreWarnings := false
	if raw.Settings.IgnoreWarnings != nil {
		ignoreWarnings = *raw.Settings.IgnoreWarnings
	}

	return domain.Settings{
		LogLevel:       logLevel,
		SaveJSONReport: saveJSONReport,
		JSONReportFile: jsonReportFile,
		Timeout:        timeout,
		IgnoreWarnings: ignoreWarnings,
	}, nil
}

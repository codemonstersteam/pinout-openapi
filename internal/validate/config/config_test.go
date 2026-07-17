package config

import (
	"errors"
	"testing"

	"pinout-openapi/internal/validate/domain"
)

// baseRawConfig — минимально валидный RawConfig; каждый тест-кейс мутирует одну
// антецедентную ветку (module-tree.md "Unit-test formula" §NewConfig, 9 units).
func baseRawConfig() domain.RawConfig {
	var raw domain.RawConfig
	raw.Consumer.Name = "consumer-a"
	raw.Consumer.ConsumedContractPath = "testdata/consumed-contract.json"
	raw.Consumer.Operations = []domain.OperationRef{{Path: "/users", Method: "get"}}
	raw.Provider.Name = "provider-a"
	raw.Provider.SpecPath = "testdata/provider.yaml"
	return raw
}

func TestNewConfig(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		cfg, err := NewConfig(baseRawConfig())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ConsumerName != "consumer-a" {
			t.Errorf("ConsumerName = %q, want %q", cfg.ConsumerName, "consumer-a")
		}
		if cfg.ConsumedContractPath != "testdata/consumed-contract.json" {
			t.Errorf("ConsumedContractPath = %q, want %q", cfg.ConsumedContractPath, "testdata/consumed-contract.json")
		}
		if len(cfg.Operations) != 1 || cfg.Operations[0].Path != "/users" || cfg.Operations[0].Method != "get" {
			t.Errorf("Operations = %+v, want [{/users get}]", cfg.Operations)
		}
		if cfg.Provider.SpecPath != "testdata/provider.yaml" {
			t.Errorf("Provider.SpecPath = %q, want %q", cfg.Provider.SpecPath, "testdata/provider.yaml")
		}
		if cfg.Settings.Timeout != defaultTimeout {
			t.Errorf("Settings.Timeout = %d, want default %d", cfg.Settings.Timeout, defaultTimeout)
		}
		if cfg.Settings.LogLevel != defaultLogLevel {
			t.Errorf("Settings.LogLevel = %q, want default %q", cfg.Settings.LogLevel, defaultLogLevel)
		}
		if cfg.Settings.SaveJSONReport != defaultSaveJSONReport {
			t.Errorf("Settings.SaveJSONReport = %v, want default %v", cfg.Settings.SaveJSONReport, defaultSaveJSONReport)
		}
		if cfg.Settings.JSONReportFile != defaultJSONReportFile {
			t.Errorf("Settings.JSONReportFile = %q, want default %q", cfg.Settings.JSONReportFile, defaultJSONReportFile)
		}
		if cfg.Settings.IgnoreWarnings != false {
			t.Errorf("Settings.IgnoreWarnings = %v, want default false", cfg.Settings.IgnoreWarnings)
		}
	})

	t.Run("empty name (consumer or provider)", func(t *testing.T) {
		emptyConsumer := baseRawConfig()
		emptyConsumer.Consumer.Name = ""
		if _, err := NewConfig(emptyConsumer); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("empty consumer.name: err = %v, want ErrConfig", err)
		}

		emptyProvider := baseRawConfig()
		emptyProvider.Provider.Name = ""
		if _, err := NewConfig(emptyProvider); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("empty provider.name: err = %v, want ErrConfig", err)
		}
	})

	t.Run("empty operations", func(t *testing.T) {
		raw := baseRawConfig()
		raw.Consumer.Operations = nil
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})

	t.Run("bad path pattern (not ^/)", func(t *testing.T) {
		raw := baseRawConfig()
		raw.Consumer.Operations = []domain.OperationRef{{Path: "users", Method: "get"}}
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})

	t.Run("bad method enum", func(t *testing.T) {
		raw := baseRawConfig()
		raw.Consumer.Operations = []domain.OperationRef{{Path: "/users", Method: "trace"}}
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})

	t.Run("spec_* both-set", func(t *testing.T) {
		raw := baseRawConfig()
		raw.Provider.SpecURL = "https://example.com/spec.yaml"
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})

	t.Run("spec_* neither-set", func(t *testing.T) {
		raw := baseRawConfig()
		raw.Provider.SpecPath = ""
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})

	t.Run("timeout <= 0", func(t *testing.T) {
		raw := baseRawConfig()
		zero := 0
		raw.Settings.Timeout = &zero
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})

	t.Run("bad log_level enum", func(t *testing.T) {
		raw := baseRawConfig()
		bad := "verbose"
		raw.Settings.LogLevel = &bad
		if _, err := NewConfig(raw); !errors.Is(err, domain.ErrConfig) {
			t.Errorf("err = %v, want ErrConfig", err)
		}
	})
}

package config

import (
	"errors"
	"testing"
)

// Unit-test formula (module-tree.md "Unit-test formula"): N = 1 happy + Σ distinguishable
// antecedent branches. 21 cases total across the 8 constructors in this file:
//   NewMethod(2) + NewOperationPath(2) + NewOperation(1) + NewProviderSource(4) +
//   NewSettings(5) + NewConsumer(3) + NewProvider(2) + NewConfig(2) = 21

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

// --- NewMethod (2) ---

func TestNewMethod(t *testing.T) {
	t.Run("happy: valid verb normalized to lower-case", func(t *testing.T) {
		m, err := NewMethod("GET")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := m.String(); got != "get" {
			t.Fatalf("String() = %q, want %q", got, "get")
		}
	})

	t.Run("invalid verb rejected", func(t *testing.T) {
		_, err := NewMethod("fetch")
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})
}

// --- NewOperationPath (2) ---

func TestNewOperationPath(t *testing.T) {
	t.Run("happy: leading slash", func(t *testing.T) {
		p, err := NewOperationPath("/users")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := p.String(); got != "/users" {
			t.Fatalf("String() = %q, want %q", got, "/users")
		}
	})

	t.Run("no leading slash rejected", func(t *testing.T) {
		_, err := NewOperationPath("users")
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})
}

// --- NewOperation (1) ---

func TestNewOperation(t *testing.T) {
	t.Run("happy: composes validated VOs", func(t *testing.T) {
		path, err := NewOperationPath("/users")
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		method, err := NewMethod("get")
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		op, err := NewOperation(path, method)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if op.Path() != path || op.Method() != method {
			t.Fatalf("Operation did not compose the given VOs: got %+v", op)
		}
	})
}

// --- NewProviderSource (4) ---

func TestNewProviderSource(t *testing.T) {
	t.Run("happy: url present", func(t *testing.T) {
		s, err := NewProviderSource("https://example.com/spec.yaml", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Kind() != ProviderSourceHTTP || s.URL() != "https://example.com/spec.yaml" {
			t.Fatalf("got %+v, want kind=http url set", s)
		}
	})

	t.Run("happy: path present", func(t *testing.T) {
		s, err := NewProviderSource("", "/specs/provider.yaml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Kind() != ProviderSourceFile || s.Path() != "/specs/provider.yaml" {
			t.Fatalf("got %+v, want kind=file path set", s)
		}
	})

	t.Run("both present rejected", func(t *testing.T) {
		_, err := NewProviderSource("https://example.com/spec.yaml", "/specs/provider.yaml")
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})

	t.Run("neither present rejected", func(t *testing.T) {
		_, err := NewProviderSource("", "")
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})
}

// --- NewSettings (5) ---

func TestNewSettings(t *testing.T) {
	t.Run("happy: defaults and explicit values", func(t *testing.T) {
		s, err := NewSettings("debug", boolPtr(true), "report.json", intPtr(60))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.LogLevel() != "debug" || !s.SaveJSONReport() || s.JSONReportFile() != "report.json" || s.Timeout() != 60 {
			t.Fatalf("got %+v, want debug/true/report.json/60", s)
		}
	})

	t.Run("bad log_level rejected", func(t *testing.T) {
		_, err := NewSettings("verbose", nil, "", nil)
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})

	t.Run("timeout below 1 rejected", func(t *testing.T) {
		_, err := NewSettings("", nil, "", intPtr(0))
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})

	t.Run("timeout above 600 rejected", func(t *testing.T) {
		_, err := NewSettings("", nil, "", intPtr(601))
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})

	t.Run("save_json_report true with missing json_report_file defaults, not rejected", func(t *testing.T) {
		// Frozen contract (config.schema.json): json_report_file has a committed default
		// ("compatibility_report.json") the app substitutes as a documented fallback. An
		// omitted path is therefore NOT a config error — it defaults. (Component oracle: CS2/
		// CS4-CS8 fixtures omit json_report_file and must reach their real stage, not CONFIG_INVALID.)
		s, err := NewSettings("", boolPtr(true), "", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !s.SaveJSONReport() || s.JSONReportFile() != "compatibility_report.json" {
			t.Fatalf("got save=%v file=%q, want true/compatibility_report.json", s.SaveJSONReport(), s.JSONReportFile())
		}
	})
}

// --- NewConsumer (3) ---

func validOperation(t *testing.T) Operation {
	t.Helper()
	path, err := NewOperationPath("/users")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	method, err := NewMethod("get")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	op, err := NewOperation(path, method)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return op
}

func TestNewConsumer(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		ops := []Operation{validOperation(t)}
		c, err := NewConsumer("checkout-service", "/specs/consumer.yaml", ops)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name() != "checkout-service" || c.SpecPath() != "/specs/consumer.yaml" || len(c.Operations()) != 1 {
			t.Fatalf("got %+v, want fields to round-trip", c)
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		ops := []Operation{validOperation(t)}
		_, err := NewConsumer("   ", "/specs/consumer.yaml", ops)
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})

	t.Run("empty operations rejected", func(t *testing.T) {
		_, err := NewConsumer("checkout-service", "/specs/consumer.yaml", nil)
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})
}

// --- NewProvider (2) ---

func TestNewProvider(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		source, err := NewProviderSource("https://example.com/spec.yaml", "")
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		p, err := NewProvider("checkout-provider", source)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name() != "checkout-provider" || p.Source() != source {
			t.Fatalf("got %+v, want fields to round-trip", p)
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		source, err := NewProviderSource("https://example.com/spec.yaml", "")
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		_, err = NewProvider("", source)
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})
}

// --- NewConfig (2) ---

const validConfigYAML = `
consumer:
  spec_path: /specs/consumer.yaml
  name: checkout-service
  operations:
    - path: /users
      method: GET
provider:
  spec_url: https://example.com/provider.yaml
  name: checkout-provider
settings:
  log_level: debug
  save_json_report: true
  json_report_file: report.json
  timeout: 45
`

func TestNewConfig(t *testing.T) {
	t.Run("happy: full config assembled", func(t *testing.T) {
		cfg, err := NewConfig(RawConfig(validConfigYAML))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Consumer().Name() != "checkout-service" {
			t.Fatalf("Consumer().Name() = %q, want checkout-service", cfg.Consumer().Name())
		}
		if cfg.Provider().Name() != "checkout-provider" {
			t.Fatalf("Provider().Name() = %q, want checkout-provider", cfg.Provider().Name())
		}
		if cfg.Settings().Timeout() != 45 {
			t.Fatalf("Settings().Timeout() = %d, want 45", cfg.Settings().Timeout())
		}
		if len(cfg.Consumer().Operations()) != 1 {
			t.Fatalf("Operations() len = %d, want 1", len(cfg.Consumer().Operations()))
		}
	})

	t.Run("unparseable YAML rejected", func(t *testing.T) {
		_, err := NewConfig(RawConfig("consumer: [unterminated flow sequence"))
		if !errors.Is(err, ErrConfigInvalid) {
			t.Fatalf("err = %v, want ErrConfigInvalid", err)
		}
	})
}

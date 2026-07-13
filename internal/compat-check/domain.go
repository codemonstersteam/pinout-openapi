// Package compatcheck is the package root of slice compat-check (module-tree.md, contracts.md):
// the head pipe (head.go), its DTOs + the four driven port interfaces (this file), and the
// sentinel-to-error.code map (errors.go). Sub-packages config/, spec/, compare/ hold the pure
// core + I/O adapters; cli/ (ticket 08) is the driving door; report/ (ticket 07) is the fourth
// I/O adapter. register.go (ticket 09) wires concrete adapters into the ports declared here — it
// does not redeclare them.
package compatcheck

import (
	"time"

	"pinout-openapi/internal/compat-check/compare"
	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// Request is the slice's one external input (module-tree.md "one external input"): the path to
// contract-tests.yaml, produced by cli.Parse (ticket 08).
type Request struct {
	ConfigPath string
}

// Outcome is the head's success value: the verdict the door (cli/, ticket 08) maps to exit 0|1
// (contracts.md "Error model"). Produced by ReportWriter.Write's success branch — the report is
// already persisted by the time an Outcome exists.
type Outcome struct {
	Verdict compare.Verdict
}

// ConfigReader is the port over contract-tests.yaml bytes (contracts.md "ConfigReader"). The
// concrete config.ConfigReader (ticket 03) satisfies it structurally — no import of this package
// needed on that side (Go structural typing, no import cycle).
type ConfigReader interface {
	Read(path string) (config.RawConfig, error)
}

// ConsumerSpecReader is the port over the consumer OpenAPI document (contracts.md
// "ConsumerSpecReader"). spec.ConsumerSpecReader (ticket 04) satisfies it structurally.
type ConsumerSpecReader interface {
	Load(path string) (spec.Spec, error)
}

// ProviderSpecReader is the port over the provider master OpenAPI document (contracts.md
// "ProviderSpecReader": Load(source) -> Result<Spec, Error>, unchanged here — no drift).
// spec.ProviderSpecReader (ticket 04) satisfies it structurally.
type ProviderSpecReader interface {
	Load(source config.ProviderSource) (spec.Spec, error)
}

// ProviderSpecReaderFactory builds a ProviderSpecReader bound to one run's resolved fetch
// timeout.
//
// DEVIATION (flagged — see head.go's ProcessCompatCheck doc + .agent/memory.md): contracts.md's
// Head section describes Deps as "four port interfaces", and the concrete spec.ProviderSpecReader
// needs config.Settings.Timeout() at construction (NewProviderSpecReader(timeout), ticket 04) —
// but that timeout value is only known after step 2 of the pipe (NewConfig), while Deps is
// assembled once at process start-up (register.go's NewDeps(), ticket 09), before any config is
// read. A static ProviderSpecReader field cannot carry the right timeout for a given run.
// Widening ProviderSpecReader.Load's own signature to also accept a timeout would drift from
// contracts.md's literal Load(source) signature — rejected. Making Deps.Provider a small factory
// instead resolves the two-phase construction without touching the port's own operation: head.go
// calls deps.Provider(timeout) once cfg.Settings().Timeout() is known (pipe step 5), then
// Load(source) on the result. ProviderSpecReader itself is untouched — only how Deps supplies an
// instance of it differs from the literal contracts.md prose.
type ProviderSpecReaderFactory func(timeout time.Duration) ProviderSpecReader

// ReportWriter is the port that persists the canon report (contracts.md "ReportWriter"). The
// concrete report.ReportWriter (ticket 07) satisfies it structurally, importing this package only
// for the Outcome return type and the ErrReportWrite sentinel (errors.go) — not for this
// interface.
type ReportWriter interface {
	Write(report compare.Report, settings config.Settings) (Outcome, error)
}

// Deps is the head's single dependency container: four autonomous I/O ports, never a raw
// *http.Client/*os.File (module-tree.md "Ports (hexagon)"). register.go (ticket 09) wires
// concrete adapters into these fields; head.go depends only on the interfaces (and the one
// factory, Provider) declared here.
type Deps struct {
	Config   ConfigReader
	Consumer ConsumerSpecReader
	Provider ProviderSpecReaderFactory
	Report   ReportWriter
}

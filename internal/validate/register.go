// Package validate — wiring (ticket 17): builds the concrete Deps that ProcessValidate
// (head.go) needs, from the real internal/validate/** adapters. No domain logic here — this
// file only constructs objects (contracts.md §ProcessValidate "Dependencies (deps)").
package validate

import (
	"log/slog"
	"os"
	"time"

	"pinout-openapi/internal/validate/config"
	"pinout-openapi/internal/validate/contract"
	"pinout-openapi/internal/validate/domain"
	"pinout-openapi/internal/validate/provider"
	"pinout-openapi/internal/validate/report"
)

// systemClock — the production Clock (contracts.md §ProcessValidate Deps "orthogonal
// tools"): a thin wrapper over time.Now, carried for downstream/future use.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// NewDeps builds the concrete Deps for ProcessValidate: the real filesystem/HTTP adapters
// (ConfigStore, ContractStore, ReportWriter) plus Clock and *slog.Logger (contracts.md
// §ProcessValidate "Dependencies (deps)"). verbose raises the log level (cli.Parse's
// --verbose flag, known before Deps is built); settings.log_level itself is only known once
// ConfigStore.Load+NewConfig run inside ProcessValidate, downstream of this construction, so
// the non-verbose default mirrors the config schema's own default ("info").
//
// BuildSpecLoader (ADR-0005): the SpecLoader is NOT built here — settings.timeout is only
// known after ConfigStore.Load+NewConfig run inside the pipe, downstream of this
// construction. Instead this wires a factory closure that the head calls once per
// invocation, after NewConfig, bounded by the real cfg.Settings.Timeout.
func NewDeps(verbose bool) Deps {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

	return Deps{
		ConfigStore:   config.NewConfigStore(),
		ContractStore: contract.NewStore(),
		BuildSpecLoader: func(s domain.Settings) SpecLoader {
			return provider.NewSpecLoader(s.Timeout)
		},
		ReportWriter: report.NewFileReportWriter(),
		Clock:        systemClock{},
		Logger:       logger,
	}
}

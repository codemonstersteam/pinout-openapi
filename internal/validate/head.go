// Package validate — slice-01-validate composition root (contracts.md §ProcessValidate,
// module-tree.md "Head-pipe pseudocode"). ProcessValidate is a linear ROP pipe wiring
// already-tested modules together; it has NO branching and NO logic of its own — a failing
// step short-circuits and its sentinel error (internal/validate/domain) rises untransformed.
// Only the cli adapter (ticket 05, not composed here) maps a sentinel to (exit code,
// error.code). NOT unit-tested (pipe of already-tested parts, ADR-0002) — exit codes are
// asserted by component scenarios once wired (ticket 17, register.go).
//
// Layout note (ADR-0004): the shared slice types + sentinel errors live in the LEAF package
// internal/validate/domain, imported by every adapter (config/compare/report/…) AND by this
// head. The head must call the adapters directly (contracts.md lists their Deps as "—"), so
// it cannot itself sit in a package the adapters import — that is why the domain moved to a
// leaf while the head stays here in package validate (head.go at its declared path). See
// ADR-0004: the import graph is head(validate) → {domain, config, compare, report}; each
// adapter → domain; domain → nothing internal. Acyclic.
package validate

import (
	"log/slog"
	"time"

	"pinout-openapi/internal/validate/compare"
	"pinout-openapi/internal/validate/config"
	"pinout-openapi/internal/validate/domain"
	"pinout-openapi/internal/validate/report"
)

// ConfigStore — the port ProcessValidate needs of ticket 06's I/O object
// (contracts.md §ConfigStore.Load). Defined here, not imported from internal/validate/config,
// so this package builds standalone; config.ConfigStore satisfies it structurally.
type ConfigStore interface {
	Load(path string) (domain.RawConfig, error)
}

// ContractStore — the port ProcessValidate needs of ticket 08's I/O object
// (contracts.md §ContractStore.Load). internal/validate/contract.Store satisfies it
// structurally.
type ContractStore interface {
	Load(path string) (domain.ConsumedContract, error)
}

// SpecLoader — the port ProcessValidate needs of ticket 09's I/O object
// (contracts.md §SpecLoader.Load). internal/validate/provider.SpecLoader satisfies it
// structurally.
type SpecLoader interface {
	Load(p domain.ProviderConfig) (domain.ProviderSpec, error)
}

// ReportWriter — the port ProcessValidate needs of ticket 15's I/O object
// (contracts.md §ReportWriter.Write). internal/validate/report.ReportWriter satisfies it
// structurally (it is already declared as an interface there).
type ReportWriter interface {
	Write(s domain.Settings, r domain.Report) (domain.Report, error)
}

// Clock — orthogonal tool listed in Deps by contracts.md §ProcessValidate (autonomous
// I/O objects + orthogonal tools, no raw *os.File/*http.Client). Not called by the linear
// pipe itself (module-tree.md pseudocode has no time step) — carried for downstream/future
// use (e.g. logging timestamps) without widening this port beyond what a caller needs.
type Clock interface {
	Now() time.Time
}

// Deps — the composition root's ports (contracts.md §ProcessValidate "Dependencies"):
// autonomous I/O objects + orthogonal tools. The wiring ticket (17, register.go) supplies
// the concrete construction; this package only needs the ports to build standalone.
//
// BuildSpecLoader (ADR-0005): a wired-once factory, not a ready SpecLoader. The frozen design
// required the SpecLoader to be both bounded by settings.timeout and built once as part of a
// fully-constructed Deps — but settings.timeout is only known after ConfigStore.Load+NewConfig
// run inside the pipe. So the SpecLoader is constructed late, once per invocation, inside
// ProcessValidate after NewConfig, bounded by the real cfg.Settings.Timeout (see the
// BuildSpecLoader call in ProcessValidate below).
type Deps struct {
	ConfigStore     ConfigStore
	ContractStore   ContractStore
	BuildSpecLoader func(domain.Settings) SpecLoader
	ReportWriter    ReportWriter
	Clock           Clock
	Logger          *slog.Logger
}

// ProcessValidate — the head: module-tree.md's linear ROP pipe, no branching of its own.
//
//	d.ConfigStore.Load        -> RawConfig         [ErrConfig]
//	config.NewConfig          -> Config            [ErrConfig]
//	d.ContractStore.Load      -> ConsumedContract  [ErrFileNotFound, ErrParse]
//	d.BuildSpecLoader(cfg.Settings) -> SpecLoader   -- LATE construction, pure/total, no error (ADR-0005)
//	loader.Load               -> ProviderSpec      [ErrFileNotFound, ErrParse, ErrHTTP, ErrTimeout]
//	compare.NewComparison -> Comparison
//	compare.CompareContracts -> ComparisonOutcome
//	report.FoldReport     -> Report
//	d.ReportWriter.Write  -> Report
//
// Antecedent: a valid Invocation. Consequent: Ok Report (compatible ⇔ errors == []); Fail
// any child's sentinel, short-circuited and risen untransformed (contracts.md §ProcessValidate).
func ProcessValidate(inv domain.Invocation, d Deps) (domain.Report, error) {
	raw, err := d.ConfigStore.Load(inv.ConfigPath)
	if err != nil {
		return domain.Report{}, err
	}

	cfg, err := config.NewConfig(raw)
	if err != nil {
		return domain.Report{}, err
	}

	consumed, err := d.ContractStore.Load(cfg.ConsumedContractPath)
	if err != nil {
		return domain.Report{}, err
	}

	loader := d.BuildSpecLoader(cfg.Settings)

	spec, err := loader.Load(cfg.Provider)
	if err != nil {
		return domain.Report{}, err
	}

	comparison, err := compare.NewComparison(cfg, consumed, spec)
	if err != nil {
		return domain.Report{}, err
	}

	outcome := compare.CompareContracts(comparison)
	rep := report.FoldReport(outcome)

	written, err := d.ReportWriter.Write(cfg.Settings, rep)
	if err != nil {
		return domain.Report{}, err
	}

	return written, nil
}

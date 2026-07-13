package compatcheck

import (
	"time"

	"pinout-openapi/internal/compat-check/compare"
	"pinout-openapi/internal/compat-check/config"
)

// ProcessCompatCheck is the slice's head: the linear ROP pipe of module-tree.md's "Head-pipe
// pseudocode" — no branching of its own. A failing step short-circuits; the underlying ErrXxx
// sentinel rises untransformed (errors.go re-exports each one at this package's root so the door,
// cli/ ticket 08, can errors.Is against it without importing config/spec/compare directly).
// INCOMPATIBLE is a verdict carried to the door, never a short-circuit here (module-tree.md "Key
// design decision") — the report is written, and Outcome returned, for both verdicts.
//
// Step 5 (deps.Provider(timeout).Load(...)) resolves the two-phase provider-reader-timeout
// construction problem — see domain.go's ProviderSpecReaderFactory doc for the full rationale.
func ProcessCompatCheck(req Request, deps Deps) (Outcome, error) {
	raw, err := deps.Config.Read(req.ConfigPath)
	if err != nil {
		return Outcome{}, err
	}

	cfg, err := config.NewConfig(raw)
	if err != nil {
		return Outcome{}, err
	}

	consumerSpec, err := deps.Consumer.Load(cfg.Consumer().SpecPath())
	if err != nil {
		return Outcome{}, err
	}

	checkedOps, err := compare.NewCheckedOperations(cfg.Consumer().Operations(), consumerSpec)
	if err != nil {
		return Outcome{}, err
	}

	timeout := time.Duration(cfg.Settings().Timeout()) * time.Second
	providerSpec, err := deps.Provider(timeout).Load(cfg.Provider().Source())
	if err != nil {
		return Outcome{}, err
	}

	comparison := compare.NewComparison(
		checkedOps, consumerSpec, providerSpec,
		cfg.Consumer().Name(), cfg.Provider().Name(), providerSourceString(cfg.Provider().Source()),
	)

	report := compare.Compare(comparison)

	return deps.Report.Write(report, cfg.Settings())
}

// providerSourceString resolves the human-readable provider-source identity buildReport needs
// (report.schema.json's provider.source field): the URL when the source is fetched over HTTP, the
// filesystem path otherwise (config.ProviderSource.Kind()).
func providerSourceString(source config.ProviderSource) string {
	if source.Kind() == config.ProviderSourceHTTP {
		return source.URL()
	}
	return source.Path()
}

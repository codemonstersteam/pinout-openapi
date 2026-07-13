package compatcheck

import (
	"errors"

	"pinout-openapi/internal/compat-check/compare"
	"pinout-openapi/internal/compat-check/config"
	"pinout-openapi/internal/compat-check/spec"
)

// Re-exported sentinels — contracts.md "Error model". Each is the SAME error value the owning
// sub-package raises (a plain var alias, not a new errors.New), so the head's ROP pipe can let
// deps.Config.Read / deps.Consumer.Load / compare.NewCheckedOperations / deps.Provider(...).Load
// errors rise untransformed (module-tree.md "ROP short-circuit") while the door (cli/, ticket 08)
// — which does not import config/spec/compare directly per its own Dependencies line — can still
// errors.Is against these package-root names and get an identical match.
var (
	ErrConfigInvalid       = config.ErrConfigInvalid
	ErrSpecUnreadable      = spec.ErrSpecUnreadable
	ErrSpecParse           = spec.ErrSpecParse
	ErrOpNotInConsumer     = compare.ErrOpNotInConsumer
	ErrProviderUnreachable = spec.ErrProviderUnreachable
	ErrProviderTimeout     = spec.ErrProviderTimeout
	ErrProviderParse       = spec.ErrProviderParse
)

// ErrReportWrite is the ReportWriter port's failure sentinel (contracts.md "ReportWriter"). Owned
// here — not re-exported from elsewhere — because report/ (ticket 07) does not exist yet as a
// package when this ticket lands; report/writer.go imports this package root for the Outcome type
// and this sentinel, per ticket 07's own Dependencies line.
var ErrReportWrite = errors.New("report write error")

// ErrorCode maps a sentinel error risen by the head's pipe to its error.code string (contracts.md
// "Error model" — the CONFIG_INVALID..REPORT_WRITE_ERROR rows only; the two verdict rows
// (compatible -> no error.code, INCOMPATIBLE) are not sentinel-driven and are not this function's
// concern). The exit-code arithmetic itself (error.code/verdict -> 0|1|2|3) is the door's job
// (cli/, ticket 08) — this function only names the failure, it does not pick an exit code.
// Returns "" when err does not match any known sentinel.
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrConfigInvalid):
		return "CONFIG_INVALID"
	case errors.Is(err, ErrSpecUnreadable):
		return "SPEC_UNREADABLE"
	case errors.Is(err, ErrSpecParse):
		return "SPEC_PARSE_ERROR"
	case errors.Is(err, ErrOpNotInConsumer):
		return "OP_NOT_IN_CONSUMER"
	case errors.Is(err, ErrProviderUnreachable):
		return "PROVIDER_UNREACHABLE"
	case errors.Is(err, ErrProviderTimeout):
		return "PROVIDER_TIMEOUT"
	case errors.Is(err, ErrProviderParse):
		return "PROVIDER_PARSE_ERROR"
	case errors.Is(err, ErrReportWrite):
		return "REPORT_WRITE_ERROR"
	default:
		return ""
	}
}

package cli

import (
	"fmt"
	"os"

	compatcheck "pinout-openapi/internal/compat-check"
)

// The four exit classes (api-specification/exit-codes.md "Exit convention"; sysexits-aligned,
// not bloated). This grid is the ONLY place error.code/verdict -> exit code happens
// (contracts.md "Error model": "mapped ONLY at the cli/ door") — the head and every I/O object
// let their sentinel rise untransformed and never compute an exit code themselves.
const (
	ExitOK            = 0 // Ok — verdict compatible, report already on stdout
	ExitIncompatible  = 1 // domain failure — a valid "no", not an error
	ExitConfigOrUsage = 2 // config/usage: CONFIG_INVALID, SPEC_UNREADABLE, SPEC_PARSE_ERROR, OP_NOT_IN_CONSUMER
	ExitEnvironment   = 3 // environment/spec-I/O: PROVIDER_UNREACHABLE, PROVIDER_TIMEOUT, PROVIDER_PARSE_ERROR, REPORT_WRITE_ERROR
)

// exitCodeByErrorCode is the frozen error.code -> exit mapping, verbatim from
// api-specification/exit-codes.md rows 2-9 / contracts.md "Error model". Kept as a table (not a
// chain of if/else) so the 8-sentinel completeness is visible at a glance and mechanically
// checkable against the frozen doc.
var exitCodeByErrorCode = map[string]int{
	"CONFIG_INVALID":       ExitConfigOrUsage,
	"SPEC_UNREADABLE":      ExitConfigOrUsage,
	"SPEC_PARSE_ERROR":     ExitConfigOrUsage,
	"OP_NOT_IN_CONSUMER":   ExitConfigOrUsage,
	"PROVIDER_UNREACHABLE": ExitEnvironment,
	"PROVIDER_TIMEOUT":     ExitEnvironment,
	"PROVIDER_PARSE_ERROR": ExitEnvironment,
	"REPORT_WRITE_ERROR":   ExitEnvironment,
}

// verdictIncompatible mirrors compare.VerdictIncompatible's underlying string value
// ("incompatible", report.schema.json "verdict" enum / contracts.md "Error model" row 1). Kept as
// a plain string instead of importing the compare package: this ticket's Dependencies line
// scopes cli/ to the compat-check package root only (Request/Outcome/the error sentinels), not
// config/spec/compare/report — and compare.Verdict's sole exported surface needed here is its
// string representation, which compatcheck.Outcome.Verdict already carries without naming the
// defining package.
const verdictIncompatible = "incompatible"

// Exit maps the head's Result<Outcome, Error> to the process boundary (contracts.md
// "cli.Parse / cli.Exit (the door)"; api-specification/exit-codes.md). The report itself is
// already on stdout by the time Exit runs — report.ReportWriter.Write (ticket 07) writes the
// machine-readable stdout line unconditionally on its own success branch — so Exit never
// re-serializes it; Exit only decides the exit code and, on a non-zero *error* exit, writes the
// error.code diagnostic line to stderr. verdict=incompatible (exit 1) is a valid domain "no", not
// an error on the Result's failure track, so no error.code diagnostic is emitted for it
// (contracts.md's Gherkin reconciliation: CS2's only Then-steps are "exit 1" + the stdout report,
// no stderr assertion). No domain logic here — this door only maps codes.
func Exit(outcome compatcheck.Outcome, err error) int {
	if err == nil {
		if string(outcome.Verdict) == verdictIncompatible {
			return ExitIncompatible
		}
		return ExitOK
	}

	code := compatcheck.ErrorCode(err)
	fmt.Fprintf(os.Stderr, "error: code=%s: %v\n", code, err)

	if exitCode, known := exitCodeByErrorCode[code]; known {
		return exitCode
	}
	// Defensive default: every sentinel the head's pipe can raise is covered by
	// exitCodeByErrorCode above (contracts.md's Error model, 8 rows). An error reaching here
	// with an unrecognized code is an unexpected/unclassified failure, not caller input — treat
	// it as environment class (3), never as a silent success.
	return ExitEnvironment
}

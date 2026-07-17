// Command tool — entry point for pinout-openapi (slice-01-validate). Wiring only (ticket
// 17, module-tree.md "main (in cmd/pinout-openapi/main.go) does three things only" — the
// binary stays at cmd/app/ per this ticket, the scaffold's original layout, and per the
// component-test harness which already builds ./cmd/app (component-tests/compose/tool.Dockerfile)):
// cli.Parse(argv) -> build Deps (register.go) -> validate.ProcessValidate(inv, deps) ->
// code := cli.ResolveExitCode(res) -> os.Exit(code). The report JSON is always the stdout
// machine channel; diagnostics go to stderr. No domain logic lives here.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"pinout-openapi/internal/validate"
	"pinout-openapi/internal/validate/cli"
	"pinout-openapi/internal/validate/domain"
)

const (
	appName = "pinout-openapi"
	version = "0.1.0"
)

// newRootCmd assembles the root cobra command. Kept in a function so tests (and main) can
// get a fresh command tree without global state.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           appName,
		Short:         "pinout-openapi — forward-compatibility validation of a consumer/provider contract pair",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	root.AddCommand(newValidateCmd())
	return root
}

// newVersionCmd — stable command: prints the version, exit 0 (component-tests/features/
// smoke.feature's structural check that the staged binary runs).
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", appName, version)
			return nil
		},
	}
}

// newValidateCmd — the live endpoint: `validate <config.yaml> [--json] [--verbose]
// [--version] [--help]`. Flag parsing is disabled on this cobra node on purpose — cli.Parse
// is the ingress door and owns every flag definition (contracts.md §cli.Parse); this node
// only routes the "validate" subcommand and hands the raw trailing args to it.
func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "validate <config.yaml>",
		Short:              "Validate a consumer/provider contract pair for forward compatibility",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, args)
		},
	}
}

// runValidate wires and drives one invocation: cli.Parse -> build Deps -> ProcessValidate
// -> ResolveExitCode -> print report/exit. No branching of its own beyond what the exit-code
// grid (contracts.md §"Error model") already dictates.
func runValidate(cmd *cobra.Command, args []string) error {
	inv, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return &exitError{code: 2}
	}

	// --version / --help recognized by cli.Parse itself (Invocation with empty ConfigPath,
	// no error) — printing/exiting on them is main's job, not the ingress door's.
	if inv.ConfigPath == "" {
		fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", appName, version)
		return nil
	}

	deps := validate.NewDeps(inv.Verbose)
	rep, procErr := validate.ProcessValidate(inv, deps)
	code := cli.ResolveExitCode(rep, procErr)

	switch {
	case code == 2:
		// CONFIG_ERROR — a config-side breach; no report on the machine channel
		// (contracts.md component scenario 2: "no report"), diagnostics to stderr.
		fmt.Fprintln(cmd.ErrOrStderr(), procErr)
	case procErr != nil:
		// FILE_NOT_FOUND / PARSE_ERROR / HTTP_ERROR / TIMEOUT_ERROR (exit 3): the head
		// short-circuited before FoldReport ever ran, so there is no Report to echo —
		// synthesize the minimal schema-valid envelope carrying the mapped error.code
		// (contracts.md §"Error model") so the machine channel always gets one.
		if writeErr := printReportJSON(cmd.OutOrStdout(), errorReport(procErr)); writeErr != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), writeErr)
		}
	default:
		// success: exit 0 (compatible) or exit 1 (verdict violations) — full report.
		if writeErr := printReportJSON(cmd.OutOrStdout(), rep); writeErr != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), writeErr)
		}
	}

	if code != 0 {
		return &exitError{code: code}
	}
	return nil
}

// errorReport wraps a short-circuited sentinel error into the frozen Report shape
// (report.schema.json) for the exit-3 adapter branches — 1:1 with contracts.md's "Error
// model" table (the same sentinel -> error.code mapping cli.ResolveExitCode already applies
// to the exit code, extended here to the report's errors[0].code string).
func errorReport(err error) domain.Report {
	return domain.Report{
		SchemaVersion: "1.0",
		Compatible:    false,
		Errors:        []domain.Violation{{Code: errorCode(err), Message: err.Error()}},
	}
}

// errorCode maps a rising sentinel (domain/errors.go) to its report.schema.json error.code
// string. ErrConfig has no case — its branch never prints a report (see runValidate).
func errorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrFileNotFound):
		return "FILE_NOT_FOUND"
	case errors.Is(err, domain.ErrParse):
		return "PARSE_ERROR"
	case errors.Is(err, domain.ErrHTTP):
		return "HTTP_ERROR"
	case errors.Is(err, domain.ErrTimeout):
		return "TIMEOUT_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// printReportJSON writes r as JSON to w — the stdout machine channel (contracts.md
// §cli.ResolveExitCode: "main prints the report JSON to stdout and diagnostics to stderr").
func printReportJSON(w io.Writer, r domain.Report) error {
	return json.NewEncoder(w).Encode(r)
}

// exitError carries an exit code through cobra out to main().
type exitError struct{ code int }

func (e *exitError) Error() string { return fmt.Sprintf("exit code %d", e.code) }

func main() {
	if err := newRootCmd().Execute(); err != nil {
		var ee *exitError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

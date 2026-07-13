// Package cli is the slice's ONE driving (primary) adapter — the door (module-tree.md "Driving
// adapter"; cli-io skill). It carries no domain logic: command.go only assembles the raw CLI
// invocation into the head's one external input DTO; exit.go only maps the head's Result back to
// the process boundary (exit code + streams). The cobra root command itself (Use/Args/RunE) and
// the Parse -> ProcessCompatCheck -> Exit orchestration are assembled in cmd/app/main.go by the
// wiring ticket (09) — this package exposes the pure pieces that wiring composes, and does not
// build compatcheck.Deps or call compatcheck.ProcessCompatCheck itself.
package cli

import (
	"fmt"
	"strings"

	compatcheck "pinout-openapi/internal/compat-check"
)

// Parse turns the CLI's raw positional arguments into the slice's one external input DTO
// (contracts.md "cli.Parse / cli.Exit (the door)"; module-tree.md "One external input
// (pinout-openapi <contract-tests.yaml>)"). The only accepted shape is exactly one positional
// argument — the path to contract-tests.yaml. Ticket-08 deliberately does not invent a default
// path beyond what contracts.md/use-case.md commit to: the config path is a required positional
// arg, no flags beyond it. No domain validation happens here (that is config.NewConfig, ticket
// 03, invoked deeper in the pipe) — this door only assembles the DTO.
func Parse(args []string) (compatcheck.Request, error) {
	if len(args) != 1 {
		return compatcheck.Request{}, fmt.Errorf(
			"usage: pinout-openapi <contract-tests.yaml>: expected exactly one config path argument, got %d",
			len(args),
		)
	}

	path := strings.TrimSpace(args[0])
	if path == "" {
		return compatcheck.Request{}, fmt.Errorf(
			"usage: pinout-openapi <contract-tests.yaml>: config path must not be empty",
		)
	}

	return compatcheck.Request{ConfigPath: path}, nil
}

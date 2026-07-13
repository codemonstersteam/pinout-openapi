// Command tool — точка входа pinout-openapi: сравнение подмножества consumer OpenAPI со
// spec провайдера (compat-check). Subcommand `run <contract-tests.yaml>` проводит
// args → cli.Parse → compatcheck.ProcessCompatCheck (реальное ядро slice-compat-check) →
// cli.Exit. Единственный внешний вход слайса — путь к contract-tests.yaml (contracts.md
// "cli.Parse / cli.Exit (the door)").
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	compatcheck "pinout-openapi/internal/compat-check"
	"pinout-openapi/internal/compat-check/cli"
	"pinout-openapi/internal/compat-check/report"
)

const (
	version = "0.1.0"
	appName = "pinout-openapi"
)

// newRootCmd собирает корневую cobra-команду. Вынесено в функцию, чтобы тесты
// могли получить свежее дерево команд без глобального состояния.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           appName,
		Short:         "pinout-openapi — compat-check consumer/provider OpenAPI спецификаций",
		Long:          "pinout-openapi run <contract-tests.yaml> сверяет подмножество операций consumer-спеки с master-спекой провайдера.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	root.AddCommand(newRunCmd())
	return root
}

// newVersionCmd — стабильная команда: печатает версию, exit 0. Используется smoke-тестом
// (структурная проверка «бинарь запускается и печатает что-то»), доменом не затрагивается.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Показать версию",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", appName, version)
			return nil
		},
	}
}

// newRunCmd — реальная ingress-труба слайса compat-check: args → cli.Parse (Request) →
// compatcheck.ProcessCompatCheck (ROP-конвейер ядра, register.NewDeps() + report.NewWriter()
// как единственная точка сборки Deps.Report — см. register.go про cycle-deviation) →
// cli.Exit (единственное место маппинга error.code/verdict → exit-код, exit-codes.md).
func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run <contract-tests.yaml>",
		Short: "Сравнить consumer-спеку с provider-спекой по contract-tests.yaml",
		Long:  "Проводит args → cli.Parse → compatcheck.ProcessCompatCheck → cli.Exit (frozen api-specification/exit-codes.md).",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := cli.Parse(args)
			if err != nil {
				// Usage error (wrong arg count / empty path) — not a head Outcome/sentinel,
				// so it never goes through cli.Exit (which expects a head result). Per
				// api-specification/exit-codes.md exit 2 is "bad invocation / usage".
				fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
				return &exitError{code: cli.ExitConfigOrUsage}
			}

			deps := compatcheck.NewDeps()
			deps.Report = report.NewWriter()

			outcome, err := compatcheck.ProcessCompatCheck(req, deps)
			code := cli.Exit(outcome, err)
			if code == 0 {
				return nil
			}
			return &exitError{code: code}
		},
	}
}

// exitError несёт exit-код через cobra наружу в main().
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

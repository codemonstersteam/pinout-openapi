// Package cli — ingress-дверь слайса slice-01-validate: превращает argv в Invocation
// (cli.Parse, ticket 04) и итоговый Result[Report,Error] в exit-код (cli.ResolveExitCode,
// отдельный тикет). Легит-проводка формы вызова (cobra), БЕЗ доменной логики и файлового
// I/O — см. docs/design/slice-01-validate/{contracts,module-tree}.md.
package cli

import (
	"github.com/spf13/cobra"

	"pinout-openapi/internal/validate/domain"
)

// Parse — ingress-дверь: argv → Invocation. `validate <config.yaml> [--json] [--verbose]
// [--version] [--help]`. Только форма (позиционный аргумент + известные флаги), без
// чтения файла и без доменных инвариантов — те лежат в config.NewConfig и downstream.
//
// Успех: Invocation{ConfigPath, JSONReport, Verbose} — путь только, файл читает
// ConfigStore ниже по трубе. `--version`/`--help` — распознанные флаги без файла
// (Invocation.ConfigPath пуст); что печатать/выполнять по ним — решает `main`
// (cmd/pinout-openapi/main.go), не эта дверь.
//
// Провал: ErrConfig — позиционный `<config.yaml>` отсутствует ИЛИ встречен
// нераспознанный флаг.
func Parse(args []string) (domain.Invocation, error) {
	var inv domain.Invocation

	cmd := &cobra.Command{
		Use:           "validate",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().BoolVar(&inv.JSONReport, "json", false, "печатать отчёт как JSON")
	cmd.Flags().BoolVar(&inv.Verbose, "verbose", false, "подробный лог")
	cmd.Flags().Bool("version", false, "печатать версию")
	cmd.Flags().BoolP("help", "h", false, "показать справку")

	if err := cmd.ParseFlags(args); err != nil {
		return domain.Invocation{}, domain.ErrConfig
	}

	showVersion, _ := cmd.Flags().GetBool("version")
	showHelp, _ := cmd.Flags().GetBool("help")
	if showVersion || showHelp {
		return domain.Invocation{JSONReport: inv.JSONReport, Verbose: inv.Verbose}, nil
	}

	positional := cmd.Flags().Args()
	if len(positional) != 1 || positional[0] == "" {
		return domain.Invocation{}, domain.ErrConfig
	}
	inv.ConfigPath = positional[0]

	return inv, nil
}

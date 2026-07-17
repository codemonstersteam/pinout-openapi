// Package cli — ingress/egress-слой слайса. Этот файл — единственное место, где
// Result[Report, Error] головы (ProcessValidate) отображается в exit code (cli-io
// grid, contracts.md §cli.ResolveExitCode / §"Error model"). Печать report JSON на
// stdout и диагностики на stderr делает main (wiring, ticket 17) — не этот файл.
package cli

import (
	"errors"

	"pinout-openapi/internal/validate/domain"
)

// ResolveExitCode отображает Result[Report, Error] головы в один exit code ∈
// {0,1,2,3} (report.schema.json x-exit-codes). Result[T, Error] — Go-идиома (T,
// error): err == nil ⇒ report валиден, err != nil ⇒ report игнорируется.
//
// Грид (contracts.md §cli.ResolveExitCode):
//
//	err == nil, report.Compatible  → 0
//	err == nil, !report.Compatible → 1
//	errors.Is(err, ErrConfig)      → 2
//	errors.Is(err, ErrFileNotFound | ErrParse | ErrHTTP | ErrTimeout) → 3
func ResolveExitCode(report domain.Report, err error) int {
	if err == nil {
		if report.Compatible {
			return 0
		}
		return 1
	}
	if errors.Is(err, domain.ErrConfig) {
		return 2
	}
	return 3
}

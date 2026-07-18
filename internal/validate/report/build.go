// Package report — слайс slice-01-validate: свёртывание ComparisonOutcome в выходной
// Report DTO (форма report.schema.json). Чистая логика, io: none, без падающего
// антецедента (contracts.md §FoldReport, module-tree.md).
package report

import "pinout-openapi/internal/validate/domain"

// FoldReport — свёртывание ComparisonOutcome → Report. schema_version зафиксирован
// ("1.0"); compatible ⇔ errors == [] (report.schema.json инвариант); errors — эхо
// Violations (никогда nil — required-поле схемы, array); provenance — сквозной эхо;
// uncovered_operations — информационно, из UncoveredOps. Нет пути отказа: outcome уже
// валиден по построению (CompareContracts, ticket 13).
func FoldReport(o domain.ComparisonOutcome) domain.Report {
	errs := o.Violations
	if errs == nil {
		errs = []domain.Violation{}
	}
	return domain.Report{
		SchemaVersion:       "1.0",
		Compatible:          len(o.Violations) == 0,
		Provenance:          o.Provenance,
		Errors:              errs,
		UncoveredOperations: o.UncoveredOps,
	}
}

// Package report — слайс slice-01-validate: свёртывание ComparisonOutcome в выходной
// Report DTO (форма report.schema.json, канон 1.1). Чистая логика, io: none, без
// падающего антецедента (contracts.md §BuildReporter/§Reporter.Fold, module-tree.md;
// change 001-report-schema-1.1: узел FoldReport разделён на фабрику + одноаргументный
// шаг — правило «один вход данных», ADR-001 async-близнеца).
package report

import (
	"time"

	"pinout-openapi/internal/validate/domain"
)

// Константы сборки отчёта (report.schema.json: const-поля канона 1.1).
const (
	reportSchemaVersion = "1.1"
	reportValidator     = "pinout-openapi"
	reportInteraction   = "sync"
)

// Reporter — коллаборатор, связывающий порт часов ДО трубы (D10: generated_at —
// единственная неаддитивная часть канона; время инжектится сверху, ядро системные
// часы не читает). Фабрика и продукт — два узла (Парнас: фабрика прячет «откуда
// время», Fold — «какова форма отчёта»).
type Reporter struct {
	clock domain.Clock
}

// BuildReporter связывает порт часов в репортера. Вызывается в bind-блоке головы
// (ProcessValidate), до первого шага трубы.
func BuildReporter(clock domain.Clock) Reporter {
	return Reporter{clock: clock}
}

// Fold — свёртывание ComparisonOutcome → Report одним аргументом данных.
// schema_version/validator/interaction — константы; compatible ⇔ errors == []
// (инвариант схемы); errors — эхо Violations (никогда nil — required-поле схемы,
// array); provenance и consumer.name — сквозные эхи; uncovered_operations —
// информационно. generated_at = clock.Now() в RFC3339 UTC Z секундной точности —
// ЕДИНСТВЕННОЕ чтение часов в слайсе (D10; проверяется countingClock-якорем).
// Нет пути отказа: outcome уже валиден по построению (CompareContracts, ticket 13).
func (r Reporter) Fold(o domain.ComparisonOutcome) domain.Report {
	errs := o.Violations
	if errs == nil {
		errs = []domain.Violation{}
	}
	return domain.Report{
		SchemaVersion:       reportSchemaVersion,
		Validator:           reportValidator,
		Interaction:         reportInteraction,
		Consumer:            domain.ReportConsumer{Name: o.ConsumerName},
		GeneratedAt:         r.clock.Now().UTC().Format(time.RFC3339),
		Compatible:          len(o.Violations) == 0,
		Provenance:          o.Provenance,
		Errors:              errs,
		UncoveredOperations: o.UncoveredOps,
	}
}

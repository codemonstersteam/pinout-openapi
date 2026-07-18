package domain

import "errors"

// Сентинелы слайса — поднимаются НЕТРАНСФОРМИРОВАННО по трубе ProcessValidate
// (короткое замыкание ROP); отображает их в (exit code, error.code) ТОЛЬКО
// cli.ResolveExitCode (contracts.md "Error model"). Ни один промежуточный модуль их не
// оборачивает и не подменяет.
var (
	// ErrConfig — конфиг не найден/нечитаем/невалиден по схеме, либо у provider не
	// ровно один источник (spec_path/spec_url). error.code=CONFIG_ERROR, exit 2.
	ErrConfig = errors.New("validate: config error")

	// ErrFileNotFound — consumed-contract или provider spec_path не найден/нечитаем.
	// error.code=FILE_NOT_FOUND, exit 3.
	ErrFileNotFound = errors.New("validate: file not found")

	// ErrParse — consumed-contract или provider spec нечитаем как валидный
	// JSON/YAML/OpenAPI. error.code=PARSE_ERROR, exit 3.
	ErrParse = errors.New("validate: parse error")

	// ErrHTTP — provider spec_url недостижим / не-2xx. error.code=HTTP_ERROR, exit 3.
	ErrHTTP = errors.New("validate: http error")

	// ErrTimeout — fetch provider spec_url превысил settings.timeout секунд.
	// error.code=TIMEOUT_ERROR, exit 3.
	ErrTimeout = errors.New("validate: timeout error")
)

// Verdict-коды — НЕ ошибки трубы, а честный доменный ответ CompareContracts на
// success-пути (folded в ComparisonOutcome.Violations / Report.Errors), exit 1
// (report.schema.json#/properties/errors/items/properties/code enum;
// x-exit-codes.table[1]). Соответствие правилам R1..R4 (contracts.md §CompareOperation):
const (
	// CodeOpNotInProvider — R1: операция consumer.operations отсутствует в provider-спеке.
	CodeOpNotInProvider = "OP_NOT_IN_PROVIDER"

	// CodeMissingRequiredRequestField — R2: requires(provider) ⊄ sends(consumer) —
	// обязательное поле/параметр запроса провайдера консьюмер не отправляет.
	CodeMissingRequiredRequestField = "MISSING_REQUIRED_REQUEST_FIELD"

	// CodeReadsFieldNotProvided — R3: reads(consumer) ⊄ provides(provider) — поле,
	// которое читает консьюмер, провайдер больше не отдаёт.
	CodeReadsFieldNotProvided = "READS_FIELD_NOT_PROVIDED"

	// CodeTypeMismatch — R4: тип общего поля/параметра не совпадает
	// (provider ≠ consumer).
	CodeTypeMismatch = "TYPE_MISMATCH"
)

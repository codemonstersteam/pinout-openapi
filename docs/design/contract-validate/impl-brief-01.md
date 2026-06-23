# Implementation brief — срез 01 `validate-contract`

Самодостаточное задание на реализацию для агента-исполнителя (skill `program-implementation`).
Источник дизайна: [`slices/01-validate-contract.md`](./slices/01-validate-contract.md), [`messages.md`](./messages.md), [`contracts-graph.md`](./contracts-graph.md), [`ref-handling.md`](./ref-handling.md), формат отчёта [`../../report-format.md`](../../report-format.md). Покрытие/границы: [`../../risk-coverage.md`](../../risk-coverage.md).

> **GATE:** писать код только после handoff-аппрува оператора в [`backlog.md`](./backlog.md). До отметки — этот файл готовится, но реализация не стартует.

## 0. Инварианты (нарушение = стоп и пересмотр)

- **Бизнес-логика ≠ I/O.** В `domain/` и `compare/` запрещены `os`, `net/http`, `time.Now`, чтение файлов. Всё I/O — за `ConfigStore`/`SpecClient`/`ReportWriter`/`Clock`.
- **Только конструкторы, без guard-функций.** Невалидный вход → структура не собирается: `NewT(raw) (T, *ValidationError)`.
- **Head — труба.** Линейная последовательность с ранним возвратом по ошибке; никакого бизнес-`if` внутри.
- **Юнит-тесты только для логики/конструкторов** (N по таблице слайса 01). Head, ингресс, I/O — компонентными.
- **Несовместимость ≠ ошибка.** `ValidateOperations` всегда возвращает `Report` (не ошибку); несовместимость живёт в `Report.Verdicts[].Compatible=false`.

## 1. Bootstrap

```
go mod init github.com/codemonstersteam/pinout-openapi   # имя сверить с remote
```
Зависимости (закрепить версии в go.mod — **сначала шаг разведки `ref-handling.md` §1**, он фиксирует версию kin-openapi):
- `github.com/getkin/kin-openapi/openapi3` — парсер OpenAPI 3.x;
- `github.com/spf13/cobra` — CLI;
- `gopkg.in/yaml.v3` — чтение `contract-tests.yaml`.

## 2. Раскладка пакетов (целевая)

```
cmd/                         # ИНГРЕСС
  root.go                    # cobra root
  validate.go                # команда validate: вызывает head, маппит исход → exit code
internal/
  domain/                    # ЛОГИКА: чистые типы + конструкторы (без I/O)
    result.go                # ValidationError, ErrorCode (const)
    operationkey.go          # OperationKey, NewOperationKey
    config.go                # Config, SpecRef, Settings, NewConfig
    contract.go              # ContractValidate, NewContractValidate
    report.go                # Report, OperationVerdict
  compare/                   # ЛОГИКА: чистое сравнение
    operations.go            # ValidateOperations, findProviderOperation
    request.go               # compareRequest
    response.go              # compareResponses, compareStatusCodes, compareContentTypes
    schema.go                # compareSchemas (+ детект цикла памятью пар)
  ioadapters/                # I/O (объекты, реализуют интерфейсы из head)
    configstore.go           # ConfigStore
    specclient.go            # SpecClient (kin-openapi, резолв $ref)
    reportwriter.go          # ReportWriter + ReportDTO + ToReportDTO
    clock.go                 # Clock
  head/
    validatecontract.go      # ValidateContract (труба) + интерфейсы Deps
testdata/                    # фикстуры спек (см. разведку ref-handling)
```

## 3. Контракты (сигнатуры, не реализация)

Идиома ошибок — Go-стиль `(value, *ValidationError)` (символично ROP; сверить со стилем `pinout-asyncapi`). `nil`-ошибка = успех.

```go
// domain/result.go
type ErrorCode string
const ( CONFIG_NOT_FOUND ErrorCode = "CONFIG_NOT_FOUND"; /* … все коды из messages.md */ )
type ValidationError struct { Code ErrorCode; Message, Location string; Context map[string]any }

// domain — конструкторы
func NewOperationKey(rawPath, rawMethod string) (OperationKey, *ValidationError)   // path "/…"; method ∈ глаголов
func NewConfig(raw RawConfig) (Config, *ValidationError)                           // 1 источник/сторону; operations непуст
func NewContractValidate(c Config, consumer, provider *openapi3.T) (ContractValidate, *ValidationError)
                                                                                  // antecedent: каждая op есть у потребителя → иначе CONSUMER_OPERATION_NOT_FOUND

// compare — чистые функции
func ValidateOperations(cv ContractValidate) Report                                // всегда Report
//   findProviderOperation → OPERATION_NOT_FOUND
//   compareRequest → REQUEST_INCOMPATIBLE ; compareResponses → RESPONSE_INCOMPATIBLE
//   compareStatusCodes → STATUS_MISMATCH ; compareContentTypes → CONTENT_TYPE_MISMATCH
//   compareSchemas(consumer, provider *openapi3.Schema, visited map[pair]bool) — рекурсия по .Value, детект цикла

// head — интерфейсы I/O (объявлены здесь, реализованы в ioadapters)
type ConfigStore  interface{ Load(path string) (RawConfig, *ValidationError) }
type SpecClient   interface{ Fetch(ref SpecRef) (*openapi3.T, *ValidationError) } // резолвит $ref; внешний → SPEC_PARSE_ERROR
type ReportWriter interface{ Write(r Report, c Config) *ValidationError }          // строит ReportDTO + Clock.Now
type Clock        interface{ Now() time.Time }
type Deps struct{ Config ConfigStore; Spec SpecClient; Writer ReportWriter; Clock Clock }
func ValidateContract(path string, d Deps) (Report, *ValidationError)
```

## 4. Порядок реализации (bottom-up, каждый шаг — компилируется + тесты зелёные)

1. `domain/result.go` — `ValidationError`, все `ErrorCode`.
2. `NewOperationKey` + юнит (N=3).
3. `NewConfig` (+`SpecRef`, `Settings`, `RawConfig`) + юнит (N=5).
4. `NewContractValidate` + юнит (N=2: happy + `CONSUMER_OPERATION_NOT_FOUND`).
5. `compare/schema.go` `compareSchemas` (+ цикл-память) + юнит (N=4) — **после разведки `ref-handling.md`**.
6. `compareRequest/Responses/StatusCodes/ContentTypes` + юнит (N=2 каждый).
7. `ValidateOperations` + `findProviderOperation` (N=2) — собирает подфункции в `Report`.
8. `ioadapters`: `ConfigStore`, `SpecClient` (kin-openapi `NewLoader`, `IsExternalRefsAllowed=false`), `ReportWriter`+`ToReportDTO`+`ReportDTO` (N=1 для маппинга), `Clock`.
9. `head/validatecontract.go` — труба (см. псевдокод ниже).
10. `cmd/` — cobra `validate`; маппинг исхода → exit 0/1/2/3 (таблица в `contracts-graph.md`).
11. `testdata/` фикстуры + раннер `component-tests/scripts/run-tests.sh` под 7 сценариев `validate.feature`.

Труба head (целевая форма):
```
Load(path) → NewConfig → Fetch(consumer) → Fetch(provider)
  → NewContractValidate → ValidateOperations → Writer.Write → return Report
// первый *ValidationError ≠ nil — ранний возврат
```

## 5. Definition of Done

- `go build ./...`, `gofmt -l` (пусто), `go vet ./...`, `go test ./...` — зелёные.
- Юнит-тесты ровно по таблице слайса 01 (логика/конструкторы); head/ингресс/I/O юнитами не покрыты.
- `component-tests/scripts/run-tests.sh` — все **7** сценариев зелёные (офлайн, через `spec_path` + `testdata/`).
- JSON-отчёт побайтово соответствует канону `report-format.md` (`schema_version`, `validator`, `interaction`, `consumer/provider {name,spec_ref,version}`, `compatible`, `verdicts[]`, `generated_at`).
- exit codes 0/1/2/3 соответствуют таблице `contracts-graph.md`.
- README / `api-specification` актуальны (skills `documentation`, `doc-quality-review`).
- Таблица разведки `$ref` заполнена в `ref-handling.md`; решение по `NOT_VERIFIED` зафиксировано (отложено/включено).

## 6. Стоп-условия для исполнителя (эскалация планировщику)

- Стиль ошибок `pinout-asyncapi` несовместим с `(value, *ValidationError)` → согласовать единый.
- Разведка показала частые внешние `$ref` → включать `NOT_VERIFIED` (правка модели) — НЕ делать молча, вернуть на план.
- Реальные спеки требуют `allOf/oneOf/anyOf` для верного вердикта → это out-of-scope, эскалация (не расширять `compareSchemas` самовольно).
- Спека поставщика за приватным git требует токен → дизайн `SpecClient` это не покрывает, эскалация.

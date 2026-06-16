# Срез 01 — validate-contract

Вход: CLI `validate <config>`. Выход: exit code + JSON-отчёт.

## Дерево модулей

```
ингресс: cmd validate (cobra)            # парсит аргумент, без логики
  └─ head: ValidateContract (труба, ROP) # линейная последовательность, без ветвления
       ├─ ConfigStore.Load(path)         → I/O: чтение YAML
       ├─ NewConfig(raw)                  → логика: конструктор
       ├─ SpecClient.Fetch(consumerRef)  → I/O: файл/URL
       ├─ SpecClient.Fetch(providerRef)  → I/O: файл/URL (master)
       ├─ NewContractValidate(...)        → логика: конструктор
       ├─ ValidateOperations(cv)          → логика: чистая функция сравнения
       └─ ReportWriter.Write(report)      → I/O: stdout + файл
```

## Контракты модулей

**head — ValidateContract**
- Signature: `ValidateContract(path: string, deps: Deps) → Result<Report>`
- Does: оркестрирует трубу; каждый шаг через `Result`, первый `Err` прерывает.
- Deps: `ConfigStore`, `SpecClient`, `ReportWriter` (реальные объекты, не сырые `*http.Client`/`os`).
- Проверяется компонентными тестами, не юнитами.

**ConfigStore.Load (I/O)** — `Load(path) → Result<rawConfig>` · `CONFIG_NOT_FOUND`, `CONFIG_INVALID`.
**SpecClient.Fetch (I/O)** — `Fetch(SpecRef) → Result<OpenAPIDoc>` · `SPEC_NOT_FOUND`, `SPEC_UNREACHABLE`, `SPEC_PARSE_ERROR`. Инкапсулирует HTTP и парсер OpenAPI (kin-openapi).
**ReportWriter.Write (I/O)** — `Write(Report) → Result<void>`. stdout + файл по настройкам.

**NewConfig (логика, конструктор)**
- `NewConfig(raw) → Result<Config>`
- Antecedent: на сторону задан ровно один источник (path|url); `operations` непустой; каждый `OperationKey` валиден.
- Consequent: success — собранный `Config`; failure — `CONFIG_INVALID` с локализацией.

**NewOperationKey (логика, конструктор)**
- Antecedent: `path` начинается с `/`; `method` ∈ HTTP-глаголов.
- Failure — `CONFIG_INVALID`.

**NewContractValidate (логика, конструктор)** — собирает вход для сравнения; success всегда при валидных доках.

**ValidateOperations (логика, чистая функция)**
- `ValidateOperations(cv: ContractValidate) → Report`
- Для каждой операции применяет семантику совместимости (см. `messages.md` §Семантика): наличие → request → response → коды → content-type.
- Подфункции (чистые): `findProviderOperation`, `compareRequest`, `compareResponses`, `compareStatusCodes`, `compareContentTypes`, `compareSchemas` (рекурсивно).

## Юнит-тесты (только логика; N = 1 happy + Σ ветвей antecedent)

| Модуль | Ветви antecedent | N |
|---|---|---|
| NewOperationKey | path без `/`; method не из списка | 3 |
| NewConfig | ноль источников; два источника; пустой operations; невалидный OperationKey | 5 |
| ValidateOperations / compareSchemas | OPERATION_NOT_FOUND; REQUEST_INCOMPATIBLE; RESPONSE_INCOMPATIBLE; STATUS_MISMATCH; CONTENT_TYPE_MISMATCH; вложенные схемы; `$ref` | ~8 |

Head, ингресс и I/O-модули юнитами **не** тестируются — только компонентными.

## Маппинг Gherkin → узел графа

| Сценарий (`component-tests/validate.feature`) | Then → узел |
|---|---|
| Совместимая пара → exit 0 | ValidateOperations (Compatible) + ReportWriter |
| Операции нет у поставщика → exit 1 | ValidateOperations: OPERATION_NOT_FOUND |
| Несовместим request → exit 1 | compareRequest: REQUEST_INCOMPATIBLE |
| Несовместим response → exit 1 | compareResponses: RESPONSE_INCOMPATIBLE |
| Нет/битый конфиг → exit 2 | ConfigStore.Load / NewConfig |
| Спека поставщика недоступна → exit 3 | SpecClient.Fetch: SPEC_UNREACHABLE |

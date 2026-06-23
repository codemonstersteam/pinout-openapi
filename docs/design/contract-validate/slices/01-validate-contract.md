# Срез 01 — validate-contract

Вход: CLI `validate <config>`. Выход: exit code + JSON-отчёт. Концептуальный дизайн (C4 Container/Component) — [`../c4.md`](../c4.md).

## Дерево модулей

```
ингресс: cmd validate (cobra)            # парсит аргумент, без логики
  └─ head: ValidateContract (труба, ROP) # линейная последовательность, без ветвления
       ├─ ConfigStore.Load(path)         → I/O: чтение YAML
       ├─ NewConfig(raw)                  → логика: конструктор
       ├─ SpecClient.Fetch(consumerRef)  → I/O: файл/URL, парсинг + резолв $ref
       ├─ SpecClient.Fetch(providerRef)  → I/O: файл/URL (master), резолв $ref
       ├─ NewContractValidate(...)        → логика: конструктор
       ├─ ValidateOperations(cv)          → логика: чистая функция сравнения → Report
       └─ ReportWriter.Write(report,cfg)  → I/O: ToReportDTO (+ Clock) → stdout + файл
```

## Контракты модулей

**head — ValidateContract**
- Signature: `ValidateContract(path: string, deps: Deps) → Result<Report>`
- Does: оркестрирует трубу; каждый шаг через `Result`, первый `Err` прерывает.
- Deps: `ConfigStore`, `SpecClient`, `ReportWriter`, `Clock` (реальные объекты, не сырые `*http.Client`/`os`/`time.Now`).
- Проверяется компонентными тестами, не юнитами.

**ConfigStore.Load (I/O)** — `Load(path) → Result<rawConfig>` · `CONFIG_NOT_FOUND`, `CONFIG_INVALID`.
**SpecClient.Fetch (I/O)** — `Fetch(SpecRef) → Result<OpenAPIDoc>` · `SPEC_NOT_FOUND`, `SPEC_UNREACHABLE`, `SPEC_PARSE_ERROR`. Инкапсулирует HTTP и парсер OpenAPI (kin-openapi); **резолвит все `$ref`** — логика получает развёрнутые схемы.
**ReportWriter.Write (I/O)** — `Write(Report, Config) → Result<void>`. Строит `ToReportDTO(Report, Config, providerVersion, Clock.Now)`, пишет канонический JSON в stdout + файл по настройкам.
**Clock (I/O)** — `Now() → time` для `generated_at`; инжектится, не вызывается из логики.

**NewConfig (логика, конструктор)**
- `NewConfig(raw) → Result<Config>`
- Antecedent: на сторону задан ровно один источник (path|url); `operations` непустой; каждый `OperationKey` валиден.
- Consequent: success — собранный `Config`; failure — `CONFIG_INVALID` с локализацией.

**NewOperationKey (логика, конструктор)**
- Antecedent: `path` начинается с `/`; `method` ∈ HTTP-глаголов.
- Failure — `CONFIG_INVALID`.

**NewContractValidate (логика, конструктор)**
- `NewContractValidate(Config, consumerDoc, providerDoc) → Result<ContractValidate>`
- Antecedent: каждая `OperationKey` конфига присутствует в спеке потребителя.
- Failure — `CONSUMER_OPERATION_NOT_FOUND` (целостность конфига, exit 2); success — собранный `ContractValidate`.

**ValidateOperations (логика, чистая функция)**
- `ValidateOperations(cv: ContractValidate) → Report`
- Для каждой операции применяет семантику совместимости (см. `messages.md` §Семантика): наличие → request → response → коды → content-type.
- Подфункции (чистые): `findProviderOperation`, `compareRequest`, `compareResponses`, `compareStatusCodes`, `compareContentTypes`, `compareSchemas` (рекурсивно).

## Юнит-тесты (только логика; N = 1 happy + Σ ветвей antecedent)

| Модуль (чистая логика) | Ветви (happy + antecedent/consequent) | N |
|---|---|---|
| NewOperationKey | happy; path без `/`; method не из списка | 3 |
| NewConfig | happy; ноль источников; два источника; пустой operations; невалидный OperationKey | 5 |
| NewContractValidate | happy; операция конфига отсутствует у потребителя (`CONSUMER_OPERATION_NOT_FOUND`) | 2 |
| findProviderOperation | happy; нет у поставщика (`OPERATION_NOT_FOUND`) | 2 |
| compareRequest | happy; `REQUEST_INCOMPATIBLE` (отсутствует обязательное поле/параметр) | 2 |
| compareResponses | happy; `RESPONSE_INCOMPATIBLE` (поле потребителя нет у поставщика) | 2 |
| compareStatusCodes | happy; `STATUS_MISMATCH` | 2 |
| compareContentTypes | happy; `CONTENT_TYPE_MISMATCH` | 2 |
| compareSchemas | happy; вложенный объект; несовпадение типа; required-поле отсутствует | 4 |
| ToReportDTO (маппинг) | happy: Report+Config+version+now → канон | 1 |

`compareSchemas` работает на схемах, разрешённых `kin-openapi` (локальные `$ref` прозрачны через `.Value`); out-of-scope MVP-конструкции (`allOf/oneOf/anyOf`, enum-сужение, format) — см. `messages.md`. Рекурсия (детект цикла памятью пар) и внешние `$ref` (→ `SPEC_PARSE_ERROR`) — план [`../ref-handling.md`](../ref-handling.md).

Head, ингресс, `SpecClient`/`ConfigStore`/`ReportWriter`/`Clock` юнитами **не** тестируются — только компонентными.

## Маппинг Gherkin → узел графа

| Сценарий (`component-tests/validate.feature`) | Then → узел |
|---|---|
| Совместимая пара → exit 0 | ValidateOperations (Compatible) + ReportWriter |
| Операции из конфига нет у потребителя → exit 2 | NewContractValidate: CONSUMER_OPERATION_NOT_FOUND |
| Операции нет у поставщика → exit 1 | ValidateOperations: OPERATION_NOT_FOUND |
| Несовместим request → exit 1 | compareRequest: REQUEST_INCOMPATIBLE |
| Несовместим response → exit 1 | compareResponses: RESPONSE_INCOMPATIBLE |
| Нет/битый конфиг → exit 2 | ConfigStore.Load / NewConfig |
| Спека поставщика недоступна → exit 3 | SpecClient.Fetch: SPEC_UNREACHABLE |

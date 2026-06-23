# contracts-graph — contract-validate

Проверка целостности графа вызовов: все типы существуют, контракты сходятся, каждый Then-шаг Gherkin имеет узел.

## Граф вызовов (труба head)

```
validate(path)
  → ConfigStore.Load(path): Result<raw>
  → NewConfig(raw): Result<Config>
  → SpecClient.Fetch(Config.ConsumerSpecRef): Result<OpenAPIDoc>
  → SpecClient.Fetch(Config.ProviderSpecRef): Result<OpenAPIDoc>
  → NewContractValidate(Config, consumerDoc, providerDoc): Result<ContractValidate>
  → ValidateOperations(ContractValidate): Report
  → ReportWriter.Write(Report, Config): Result<void>   # ToReportDTO(Report,Config,version,Clock.Now) → канон JSON
  → exitCode(Report | ValidationError)
```

## Сходимость типов

| Производит | Тип | Потребляет |
|---|---|---|
| ConfigStore.Load | rawConfig | NewConfig |
| NewConfig | Config | SpecClient.Fetch, NewContractValidate |
| SpecClient.Fetch ×2 | OpenAPIDoc | NewContractValidate |
| NewContractValidate | ContractValidate | ValidateOperations |
| ValidateOperations | Report | ReportWriter.Write, exitCode |
| ToReportDTO (в ReportWriter) | ReportDTO | сериализация в канон JSON |

Все типы определены в [`messages.md`](./messages.md). ✅ висячих типов нет. `Clock` инжектится в Deps (источник `generated_at`), в граф данных не входит.

## ErrorCode → exit code

| Exit | Смысл | Коды |
|---|---|---|
| 0 | совместимо | — (`Report.Compatible == true`) |
| 1 | несовместимо | `OPERATION_NOT_FOUND`, `REQUEST_INCOMPATIBLE`, `RESPONSE_INCOMPATIBLE`, `STATUS_MISMATCH`, `CONTENT_TYPE_MISMATCH` |
| 2 | ошибка конфигурации | `CONFIG_NOT_FOUND`, `CONFIG_INVALID`, `CONSUMER_OPERATION_NOT_FOUND` |
| 3 | ошибка загрузки/парсинга спеки | `SPEC_NOT_FOUND`, `SPEC_UNREACHABLE`, `SPEC_PARSE_ERROR` |

## Соответствие Gherkin

Все 7 сценариев `validate.feature` имеют узел в графе (см. маппинг в [`slices/01-validate-contract.md`](./slices/01-validate-contract.md)). ✅ непокрытых Then-шагов нет.

## Hard checks (program-design)

- [x] Один data-аргумент на узел (зависимости инжектятся отдельно через `Deps`).
- [x] Нет сырых `*http.Client`/`os`/`time.Now` в зависимостях логики — спрятаны в `SpecClient`/`ConfigStore`/`ReportWriter`/`Clock`.
- [x] `$ref` резолвятся на IO-границе (`SpecClient.Fetch`) — `compareSchemas` работает на развёрнутых схемах.
- [x] `generated_at` берётся из `Clock` (вызывающий), не из чистой логики; `Report` (логика) → `ReportDTO` (канон) маппится в `ReportWriter`.
- [x] Нет guard-функций — только конструкторы (`NewConfig`, `NewOperationKey`, `NewContractValidate`).
- [x] Head — труба без кастомной логики (ветвление только через `Result`).
- [x] Каждый Then-шаг Gherkin имеет узел.

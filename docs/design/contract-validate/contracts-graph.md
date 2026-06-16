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
  → ReportWriter.Write(Report): Result<void>
  → exitCode(Report)
```

## Сходимость типов

| Производит | Тип | Потребляет |
|---|---|---|
| ConfigStore.Load | rawConfig | NewConfig |
| NewConfig | Config | SpecClient.Fetch, NewContractValidate |
| SpecClient.Fetch ×2 | OpenAPIDoc | NewContractValidate |
| NewContractValidate | ContractValidate | ValidateOperations |
| ValidateOperations | Report | ReportWriter.Write, exitCode |

Все типы определены в [`messages.md`](./messages.md). ✅ висячих типов нет.

## Соответствие Gherkin

Все 6 сценариев `validate.feature` имеют узел в графе (см. маппинг в [`slices/01-validate-contract.md`](./slices/01-validate-contract.md)). ✅ непокрытых Then-шагов нет.

## Hard checks (program-design)

- [x] Один data-аргумент на узел (зависимости инжектятся отдельно через `Deps`).
- [x] Нет сырых `*http.Client`/`os` в зависимостях логики — спрятаны в `SpecClient`/`ConfigStore`/`ReportWriter`.
- [x] Нет guard-функций — только конструкторы (`NewConfig`, `NewOperationKey`, `NewContractValidate`).
- [x] Head — труба без кастомной логики (ветвление только через `Result`).
- [x] Каждый Then-шаг Gherkin имеет узел.

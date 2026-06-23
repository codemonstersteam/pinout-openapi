# messages — contract-validate

Доменные типы и `Result<T, Error>`. I/O-типы (сырой YAML, сетевые байты) в домен не попадают — конвертируются в I/O-модулях.

## Result / ошибки (ROP, как в pinout-asyncapi)

```
Result<T> = Ok(T) | Err(ValidationError)

ValidationError {
  Code:     ErrorCode          # машиночитаемый код
  Message:  string             # для человека
  Location: string             # где (config / consumer.spec / provider.spec / operation)
  Context:  map[string]any     # детали диагностики
}

ErrorCode ∈ {
  CONFIG_NOT_FOUND, CONFIG_INVALID,
  SPEC_NOT_FOUND, SPEC_PARSE_ERROR, SPEC_UNREACHABLE,
  CONSUMER_OPERATION_NOT_FOUND,           # операция из конфига отсутствует в спеке потребителя
  OPERATION_NOT_FOUND, REQUEST_INCOMPATIBLE,
  RESPONSE_INCOMPATIBLE, STATUS_MISMATCH, CONTENT_TYPE_MISMATCH
}
```

## Доменные типы (собираются только конструкторами)

```
OperationKey { Path: string, Method: HTTPMethod }
  NewOperationKey(rawPath, rawMethod) → Result<OperationKey>
  # antecedent: path начинается с "/"; method ∈ {GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS}

Config {
  ConsumerSpecRef: SpecRef
  ProviderSpecRef: SpecRef          # master = прод
  ConsumerName, ProviderName: string
  Operations: []OperationKey        # непустой
  Settings: Settings
}
  NewConfig(rawConfig) → Result<Config>
  # antecedent: задан ровно один источник на сторону; Operations непустой; каждый OperationKey валиден

SpecRef = LocalPath(string) | RemoteURL(string)   # ровно один вариант

OpenAPIDoc { ... }                  # результат парсинга (kin-openapi); ВСЕ $ref резолвятся
                                    # на IO-границе (SpecClient.Fetch) — логика их не видит

ContractValidate {                  # вход в чистую логику сравнения
  Operations:   []OperationKey
  ConsumerDoc:  OpenAPIDoc
  ProviderDoc:  OpenAPIDoc
}
  NewContractValidate(Config, consumerDoc, providerDoc) → Result<ContractValidate>
  # antecedent: каждая OperationKey конфига присутствует в спеке потребителя
  #             → иначе CONSUMER_OPERATION_NOT_FOUND (ошибка целостности конфига, exit 2)

OperationContract {                 # проекция операции из OpenAPIDoc
  Key:          OperationKey
  Request:      SchemaSet           # параметры + тело
  Responses:    map[StatusCode]SchemaSet
  ContentTypes: []string
}

OperationVerdict {
  Key:        OperationKey
  Compatible: bool
  Errors:     []ValidationError     # пусто, если Compatible
}

Report {                            # ЧИСТЫЙ результат ValidateOperations — без I/O-полей
  Consumer, Provider: string        # имена сторон (из Config)
  Verdicts: []OperationVerdict
  Compatible: bool                  # AND по всем
}

ReportDTO {                         # сериализуется в канон docs/report-format.md
  SchemaVersion: "1.0"
  Validator:     "pinout-openapi"
  Interaction:   "sync"
  Consumer:      { Name, SpecRef, Version }   # SpecRef/Version — из Config/спеки
  Provider:      { Name, SpecRef, Version }   # Version = коммит/тег master-спеки
  Compatible:    bool
  Verdicts:      []{ Subject:"METHOD /path", Compatible, Errors }
  GeneratedAt:   RFC3339            # ← Clock (вызывающий), НЕ из логики
}
  ToReportDTO(Report, Config, providerVersion, now Clock) → ReportDTO
  # маппинг строится на IO-границе (ReportWriter), не в чистой логике
```

## Семантика совместимости (consequent логики сравнения)

Предусловие (гарантировано `NewContractValidate`): каждая `OperationKey` присутствует в спеке потребителя. Для каждой `OperationKey`:

1. **Наличие у поставщика.** Операция (path+method) есть у поставщика → иначе `OPERATION_NOT_FOUND`.
2. **Request.** Каждый обязательный параметр/поле, которое требует поставщик, потребитель предоставляет → иначе `REQUEST_INCOMPATIBLE`.
3. **Response.** Каждое поле, которое ожидает потребитель в ответе, присутствует в схеме ответа поставщика (provider ⊇ consumer) → иначе `RESPONSE_INCOMPATIBLE`.
4. **Коды.** Код(ы) успеха, которые обрабатывает потребитель, поставщик может вернуть → иначе `STATUS_MISMATCH`.
5. **Content-Type.** Согласован → иначе `CONTENT_TYPE_MISMATCH`.

Сравнение схем — рекурсивно по required-полям и типам (переиспользуем подход `pinout-asyncapi`).

**Out-of-scope MVP `compareSchemas`** (логируются как непроверенные, не валятся в false-negative): `allOf`/`oneOf`/`anyOf`/`discriminator`, сужение `enum`, `format` и числовые/строковые ограничения (`minimum`, `maxLength`, …), `nullable`. Покрываются компонентными тестами потребителя (слой методологии), не инструментом.

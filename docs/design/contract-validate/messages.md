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

OpenAPIDoc { ... }                  # результат парсинга (библиотека kin-openapi)

ContractValidate {                  # вход в чистую логику сравнения
  Operations:   []OperationKey
  ConsumerDoc:  OpenAPIDoc
  ProviderDoc:  OpenAPIDoc
}
  NewContractValidate(Config, consumerDoc, providerDoc) → Result<ContractValidate>

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

Report {                            # сериализуется в канон docs/report-format.md
  Consumer, Provider: string
  Verdicts: []OperationVerdict
  Compatible: bool                  # AND по всем
  ProviderSpecVersion: string       # коммит/версия master-спеки
}                                   # ↑ маппинг 1:1 на общий формат отчёта экосистемы
```

## Семантика совместимости (consequent логики сравнения)

Для каждой `OperationKey` потребителя:

1. **Наличие.** Операция (path+method) есть у поставщика → иначе `OPERATION_NOT_FOUND`.
2. **Request.** Каждый обязательный параметр/поле, которое требует поставщик, потребитель предоставляет → иначе `REQUEST_INCOMPATIBLE`.
3. **Response.** Каждое поле, которое ожидает потребитель в ответе, присутствует в схеме ответа поставщика (provider ⊇ consumer) → иначе `RESPONSE_INCOMPATIBLE`.
4. **Коды.** Код(ы) успеха, которые обрабатывает потребитель, поставщик может вернуть → иначе `STATUS_MISMATCH`.
5. **Content-Type.** Согласован → иначе `CONTENT_TYPE_MISMATCH`.

Сравнение схем — рекурсивно по required-полям и типам (переиспользуем подход `pinout-asyncapi`).

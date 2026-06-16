# Общий формат отчёта валидатора (канон экосистемы)

Канонический JSON-формат, который отдают **все** валидаторы pinout (`pinout-openapi`, `pinout-asyncapi`) и потребляет `pinout-netlist`. Определён здесь по решению оператора; async подстраивается (эпик E0), netlist потребляет (E2).

`schema_version` версионирует сам формат. Менять — только наращивая версию.

## Схема

```json
{
  "schema_version": "1.0",
  "validator": "pinout-openapi",            // или "pinout-asyncapi"
  "interaction": "sync",                     // "sync" | "async"
  "consumer": {
    "name": "mq-rest-sync-adapter",
    "spec_ref": "../api-specification/openapi.yml",
    "version": ""                            // коммит/тег спеки, если известен
  },
  "provider": {
    "name": "wallet-balance-service",
    "spec_ref": "https://.../-/raw/main/api-specification/openapi.yml",
    "version": ""                            // коммит/тег master-спеки (= прод)
  },
  "compatible": true,                        // AND по всем verdicts
  "verdicts": [
    {
      "subject": "GET /balance",            // sync: "METHOD /path"; async: имя канала/операции
      "compatible": true,
      "errors": []                           // непусто, если compatible=false
    }
  ],
  "generated_at": "2026-06-16T00:00:00Z"     // RFC3339, проставляет вызывающий
}
```

## Объект ошибки (`verdicts[].errors[]`)

Совпадает с `ValidationError` обоих валидаторов:

```json
{
  "code": "RESPONSE_INCOMPATIBLE",          // машиночитаемый код
  "message": "field 'balance' missing in provider response",
  "location": "operation GET /balance / responses/200",
  "context": { "expected": ["balance"], "got": [] }
}
```

## Правила

- `compatible` верхнего уровня = логическое И по всем `verdicts[].compatible`.
- `subject` — единственное поле, чья форма зависит от `interaction` (sync: `METHOD /path`; async: канал/операция). Остальное одинаково.
- Коды ошибок — из словаря конкретного валидатора; netlist хранит их как строки, не интерпретируя семантику протокола.
- Совместимость с async: текущий `compatibility_report.json` приводится к этому формату в рамках E0 без потери данных.

# pinout-openapi — валидатор синхронных контрактов (Go CLI)

## Проблема
В CI, до мержа, нужно детерминированно ответить: совместим ли потребитель REST-API с master-спекой
поставщика (master = прод). Сегодня это ловится только в рантайме / на ревью.

## Что делает (границы)
- **Делает:** CLI `validate <config>` — по `contract-tests.yaml` сверяет операции, на которые опирается
  потребитель, с контрактом поставщика: наличие операции (path+method), совместимость схем
  request/response, коды ответов, content-type. **Чистая функция** сравнения двух OpenAPI-спек.
- **НЕ делает:** не поднимает заглушки, не гоняет тесты, не генерирует SDK, не проверяет конформность
  сервиса своей спеке (это компонентные тесты потребителя).

## Вход — `contract-tests.yaml` (симметрично `pinout-asyncapi`; различие: `operations` вместо `channels`)

```yaml
# Конфигурация проверки совместимости синхронного контракта
contract_tests:
  # Спецификация потребителя (локальная)
  consumer:
    spec_path: "../api-specification/openapi.yml"
    name: "mq-rest-sync-adapter"
    operations:                        # операции, на которые опирается потребитель (sync ≡ channels в async)
      - { path: "/balance", method: "GET" }

  # Спецификация поставщика (master = прод; удалённая ИЛИ локальная для офлайн-тестов)
  provider:
    spec_url: "https://git.codemonsters.team/guides/wallet-balance/-/raw/main/api-specification/openapi.yml"
    # spec_path: "./testdata/provider_openapi.yml"   # альтернатива для офлайн-тестов, как в async
    name: "wallet-balance-service"

  # Настройки проверки
  settings:
    log_level: "info"                  # debug | info | warn | error
    save_json_report: true
    json_report_file: "compatibility_report.json"
    timeout: 30                        # сек, ожидание загрузки спеки поставщика
    ignore_warnings: false             # только breaking changes
```

## Выход — exit code + канонический JSON-отчёт
- **exit:** `0` совместим · `1` несовместим · `2` ошибка конфига/целостности · `3` ошибка спеки.
- **отчёт (канон экосистемы):** `schema_version` · `validator: "pinout-openapi"` · `interaction: "sync"` ·
  `consumer`/`provider` `{spec_ref, version}` · `compatible` (AND по всем `verdicts`) · `verdicts[]` ·
  `generated_at` (RFC3339). Формат симметричен `pinout-asyncapi`.

## Режимы отказа (`error.code`)
`CONFIG_NOT_FOUND` · `CONFIG_INVALID` · `SPEC_NOT_FOUND` · `SPEC_UNREACHABLE` · `SPEC_PARSE_ERROR` ·
`CONSUMER_OPERATION_NOT_FOUND` (операция конфига отсутствует у самого потребителя) ·
`REQUEST_INCOMPATIBLE` · `RESPONSE_INCOMPATIBLE` · операция потребителя отсутствует у поставщика.

## Ограничения стека
Go + `cobra` (CLI) · `kin-openapi` (парсинг OpenAPI 3.x + **резолв всех `$ref`**, honest reuse — валидацию
схем не переизобретаем) · YAML-конфиг · JSON-отчёт. Поставщик: `spec_url` ИЛИ `spec_path` (офлайн-тесты).

## Вне MVP
`allOf`/`oneOf`/`anyOf` · сужение `enum` · проверки `format` · `NOT_VERIFIED`;
внешние (кросс-файловые) `$ref` → `SPEC_PARSE_ERROR`.

## Definition of Done
- `go build ./...` и `go test ./...` зелёные (юнит-тесты — только для логики сравнения);
- **7 компонентных Gherkin-сценариев** зелёные (в Docker, чёрный ящик над CLI): совместимая пара ·
  операции нет у поставщика · несовместимый request · несовместимый response · операция конфига нет
  у потребителя · конфиг отсутствует/невалиден · спека поставщика недоступна;
- `README.md` + `api-specification/` актуальны (что делает, как запустить, конфиг, формат отчёта).

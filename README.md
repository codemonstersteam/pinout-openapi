# pinout-openapi

Концепт экосистемы: [pinout](https://github.com/codemonstersteam/pinout) ([локально](../pinout/README.md)).

Валидатор синхронных контрактов между REST-сервисами. **Чистая функция**: сравнивает OpenAPI-спецификацию потребителя с master-OpenAPI поставщика (master = прод) и выдаёт вердикт совместимости на pre-merge стадии в CI. Симметричен [`pinout-asyncapi`](https://github.com/codemonstersteam/pinout-asyncapi) по конфигурации и формату отчёта.

## Границы

- **Делает:** по `contract-tests.yaml` сверяет операции, на которые опирается потребитель, с контрактом поставщика — наличие операции (path+method), совместимость схем request/response, коды ответов, content-type.
- **Не делает:** не поднимает заглушки и не гоняет тесты (это слой методологии в репозитории потребителя); не генерирует клиентский SDK; не проверяет, что сам сервис конформен своей спеке (это компонентные тесты сервиса).

## Источник истины и обоснование

Спека поставщика в Git — источник истины. Подход и его обоснование (несущий инвариант «сервис конформен своей спеке ⇒ совместимость спек = реальная совместимость», границы, сравнение с генерацией либ) описаны в [концепте pinout](../pinout/README.md). Семантику использования контракта закрывают компонентные тесты потребителя, чьи сценарии и заглушки выведены из спеки поставщика (доработка скилла `component-tests` в [service-template](https://github.com/ubik-life/service-template/)).

## Конфигурация

`contract-tests.yaml` — симметричен `pinout-asyncapi`:

```yaml
contract_tests:
  consumer:
    spec_path: "../api-specification/openapi.yml"   # спека потребителя (локально)
    name: "mq-rest-sync-adapter"
    operations:                                       # операции, которые потребитель использует
      - { path: "/balance", method: "GET" }
  provider:
    spec_url: "https://git.codemonsters.team/guides/wallet-balance/-/raw/main/api-specification/openapi.yml"  # master = прод
    # spec_path: "./testdata/provider_openapi.yml"   # альтернатива для офлайн-тестов, как в async
    name: "wallet-balance-service"
  settings:
    log_level: "info"
    save_json_report: true
    json_report_file: "compatibility_report.json"
    timeout: 30
```

## Запуск (целевой CLI, симметрично async)

```bash
go run ./cmd validate ./contract-tests.yaml
# ✅ Контракты совместимы
# Потребитель: GET /balance
# Exit codes: 0 совместимо · 1 несовместимо · 2 ошибка конфигурации · 3 ошибка парсинга/загрузки спек
```

## Формат отчёта

JSON-отчёт — в общем формате экосистемы, определённом здесь: [`docs/report-format.md`](./docs/report-format.md). Его потребляет `pinout-netlist`; `pinout-asyncapi` подстраивается под него (эпик E0).

## Методология и проектирование

Репозиторий разрабатывается по скиллам [service-template](https://github.com/ubik-life/service-template/) — см. [`CLAUDE.md`](./CLAUDE.md). Пакет проектирования MVP — [`docs/design/contract-validate/`](./docs/design/contract-validate/). Статус и тикеты — [`docs/design/contract-validate/backlog.md`](./docs/design/contract-validate/backlog.md).

## Статус

📋 Проектирование (MVP — эпик E1 в [бэклоге экосистемы](../pinout/backlog.md)). Код ещё не реализован: сначала пакет проектирования и handoff-аппрув оператора по скиллу `program-design`.

## Лицензия

Открытая лицензия (см. `LICENSE`).

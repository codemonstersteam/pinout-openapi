# AGENTS.md — pinout-openapi

Точка входа для агента/разработчика. **Порядок чтения:** этот файл → [`CLAUDE.md`](./CLAUDE.md) → пакет проектирования [`docs/design/contract-validate/`](./docs/design/contract-validate/).

## Что это

Валидатор синхронных контрактов OpenAPI. Концепт экосистемы: [`../pinout/README.md`](../pinout/README.md). Эпик **E1** в [`../pinout/backlog.md`](../pinout/backlog.md). MVP — чистая функция «`consumer OpenAPI` vs `provider master OpenAPI`», симметрично `pinout-asyncapi`.

## Методология (скиллы, закреплённая версия)

Разработка строго по скиллам [ubik-life/service-template](https://github.com/ubik-life/service-template/). **Закреплённый коммит: `5ad8347c38ea7f07bd0620ebeef5030ca3431efd`** (зафиксирован 2026-06-17). Скиллы используем по ссылке, локальных копий не держим. Перечень скиллов и выжимка правил — в [`CLAUDE.md`](./CLAUDE.md).

## Resume here — откуда продолжить

**Состояние:** концепт и бэклог экосистемы готовы (этапы 0–1); пакет проектирования MVP готов (этап 2); **код не начат**.

**Следующий шаг (следующий митап):** продолжить проектирование/реализацию `pinout-openapi` по бэклогу.

- Срез один — `validate-contract`: [`docs/design/contract-validate/slices/01-validate-contract.md`](./docs/design/contract-validate/slices/01-validate-contract.md) (дерево модулей, контракты, юнит-N, маппинг Gherkin). Готов к реализации bottom-up.
- Граф контрактов сходится: [`contracts-graph.md`](./docs/design/contract-validate/contracts-graph.md).
- Тикет и DoD: [`backlog.md`](./docs/design/contract-validate/backlog.md).
- Решения зафиксированы: парсер `kin-openapi`; формат отчёта — канон [`docs/report-format.md`](./docs/report-format.md); поставщик `spec_url`+`spec_path`.

**Gate:** реализация (skill `program-implementation`) начинается ТОЛЬКО после handoff-аппрува оператора в [`backlog.md`](./docs/design/contract-validate/backlog.md) (строка «Оператор аппрувит — @handle, дата»).

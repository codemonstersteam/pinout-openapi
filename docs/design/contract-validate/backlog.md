# backlog — contract-validate (пакет проектирования)

Один тикет на срез + handoff-чеклист (skill `program-design` → `program-implementation`).

## Тикеты

### T01 — срез validate-contract (MVP)
Реализация по [`slices/01-validate-contract.md`](./slices/01-validate-contract.md). Bottom-up: доменные конструкторы → чистая логика сравнения → I/O-модули → head-труба → ингресс (cobra) → регистрация в `cmd`.

- [ ] `NewOperationKey`, `NewConfig`, `NewContractValidate` + юнит-тесты (N по таблице).
- [ ] `ValidateOperations` + подфункции сравнения (`compareSchemas` рекурсивно, `$ref`) + юнит-тесты.
- [ ] I/O: `ConfigStore`, `SpecClient` (HTTP/файл + парсер OpenAPI kin-openapi), `ReportWriter`.
- [ ] head `ValidateContract` (ROP-труба).
- [ ] ингресс `cmd validate` (cobra) + exit codes 0/1/2/3.
- [ ] `Report` в общем JSON-формате экосистемы (согласовать с эпиком E0 `pinout-asyncapi`).
- [ ] Компонентные тесты `component-tests/validate.feature` (6 сценариев) зелёные.
- [ ] README/`api-specification` актуальны (skill `documentation`, `doc-quality-review`).

## Definition of Done пакета

- [ ] Все 12 шагов `program-design` пройдены.
- [ ] Граф контрактов сходится ([`contracts-graph.md`](./contracts-graph.md)).
- [ ] Все Then-шаги Gherkin имеют узел.
- [ ] Юнит-тесты только для логики; head/ингресс/I/O — компонентными.
- [ ] `gofmt`, `go vet`, `go test ./...`, `component-tests/scripts/run-tests.sh` — зелёные.

## Решения оператора (зафиксированы)

1. ✅ Парсер OpenAPI — **`kin-openapi`** (honest reuse, не переизобретаем валидацию схем).
2. ✅ Общий JSON-формат отчёта определяется **здесь**, в `pinout-openapi` — канон [`docs/report-format.md`](../../report-format.md). `pinout-asyncapi` (эпик E0) подстраивается под него; `pinout-netlist` (E2) его потребляет.
3. ✅ Поставщик поддерживает **`spec_url` и `spec_path`**, как в async (офлайн-тесты). `SpecRef = LocalPath | RemoteURL` уже покрывает это (см. [`messages.md`](./messages.md)).

## Handoff

- [ ] Оператор аппрувит пакет — @<github-handle>, YYYY-MM-DD

Реализация (`program-implementation`) **не начинается** до отметки аппрува выше.

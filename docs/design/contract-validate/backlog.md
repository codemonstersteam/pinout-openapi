# backlog — contract-validate (пакет проектирования)

Один тикет на срез + handoff-чеклист (skill `program-design` → `program-implementation`).

## Тикеты

### T01 — срез validate-contract (MVP)
Реализация по [`slices/01-validate-contract.md`](./slices/01-validate-contract.md); пошаговое задание исполнителю — [`impl-brief-01.md`](./impl-brief-01.md). Bottom-up: доменные конструкторы → чистая логика сравнения → I/O-модули → head-труба → ингресс (cobra) → регистрация в `cmd`.

- [ ] `NewOperationKey`, `NewConfig`, `NewContractValidate` + юнит-тесты (N по таблице).
- [ ] `ValidateOperations` + подфункции сравнения (`findProviderOperation`, `compareRequest/Responses/StatusCodes/ContentTypes/Schemas`) на развёрнутых ($ref-free) схемах + юнит-тесты.
- [ ] Обработка `$ref` в `compareSchemas` по плану [`ref-handling.md`](./ref-handling.md): разведка → A (сравнение, local-ref прозрачны) → C (детект цикла памятью пар); внешние ref → `SPEC_PARSE_ERROR`. `NOT_VERIFIED` — отложено (триггер в плане).
- [ ] I/O: `ConfigStore`, `SpecClient` (HTTP/файл + парсер kin-openapi, **резолв `$ref`**), `ReportWriter`, `Clock`.
- [ ] head `ValidateContract` (ROP-труба).
- [ ] ингресс `cmd validate` (cobra) + exit codes 0/1/2/3 (таблица `ErrorCode→exit` в `contracts-graph.md`).
- [ ] `ToReportDTO`: `Report` → канонический JSON экосистемы (`validator/interaction/spec_ref/version/generated_at`); согласовать с E0 `pinout-asyncapi`.
- [ ] Компонентные тесты `component-tests/validate.feature` (7 сценариев) зелёные.
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

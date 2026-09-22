# Change 001 — report-schema-1.1: перевод выхода валидатора на канон `1.1`

**Вес:** minor (аддитивная эволюция замороженного контракта). **Срез:** slice-01-validate.
**Вертикаль:** SemVer-minor (`@change-intake` → evolve контракта → implementation → `@fagan`).
**Постановка:** `debt/02-report-schema-1.1.md` (тикет 2/2 долга `pinout/debt/report-canon-fork.md`).
**Источник истины формы:** канон `docs/report-format.md` v1.1 (смержен PR #8) — расхождение с ним = дефект.

## Spec-delta (что меняется в контракте)

`api-specification/report.schema.json`: `x-frozen` перештамповывается `2026-07-16 → 2026-09-22`,
`schema_version: const "1.0" → const "1.1"`. Дельта полей — аддитивная надстройка, соответствующая
канону 1.1 (владелец канона — этот репозиторий, эпик E1):

| Поле | Было (1.0, заморожено 2026-07-16) | Станет (1.1) |
|---|---|---|
| `schema_version` | `const "1.0"` | `const "1.1"` |
| `validator` | — | required, `const "pinout-openapi"` |
| `interaction` | — | required, `const "sync"` |
| `consumer` | — | required, closed-объект `{name}` (required, minLength 1); `version` optional, не эмитится (источника нет) |
| `generated_at` | — | required, RFC3339 UTC `Z` секундной точности |
| `errors[].subject` | — | свойство добавлено; для verdict-кодов (exit 1) обязательно и равно `METHOD /path` (метод UPPER, path байт-в-байт из consumed-contract); для io/parse (exit 3) опускается — поэтому в `required` items НЕ входит (канон §1: «Exit-3 errors have no operation/channel identified, so the key is omitted») |
| `compatible`, `provenance`, `errors[]` (форма), `uncovered_operations`, `x-exit-codes` | без изменений | инвариант `compatible ⇔ errors == []` цел |

`additionalProperties: false` сохраняется на корне, `consumer` и `provenance` (политика канона).

### Решение по смене `const "1.0" → const "1.1"` (не breaking)

Смена значения `schema_version` — смена версии самого канона, а не удаление/перегруппировка полей
данных. Операторское решение зафиксировано в `pinout/debt/report-canon-fork.md` (вариант A,
2026-07-25) и политике версионирования канона (§8): новое поле → minor, перегруппировка/удаление →
major. `validate-contract-diff` обязан пропустить эту дельту как аддитивную; вердикт по
const-строке фиксируется здесь как решение оператора.

### Отличие от async-близнеца (осознанное)

`pinout-asyncapi/api-specification/report.schema.json` (x-frozen 2026-07-25) объявляет
`errors[].subject` в `required` items — и его же exit-3 конверт субъект не эмитит (зафиксировано
долгом N1 в `changes/001-pipe-arity-coverage-gaps/change-delta.md` §3). Настоящая схема следует
канону владельца: subject не в `required`, обязателен для verdict-кодов. Дрейф twin-схем по этому
полю фиксируется здесь; правка замороженной схемы близнеца — вне границ этого изменения.

## Discriminating (old ≠ new, наблюдаемо на машино-канале)

- OLD: отчёт exit 1 не содержит ни идентичности сторон, ни времени, ни субъекта нарушения —
  `pinout-netlist` не может заполнить `Edge.{Consumer,Interaction,Subject}`, `VerdictRecord.At`.
- NEW: тот же прогон печатает `validator: "pinout-openapi"`, `interaction: "sync"`,
  `consumer.name` (из `config.consumer.name`), `generated_at` (RFC3339 UTC), и каждая verdict-ошибка
  несёт `subject: "GET /path"` — префикс её `location`, вычисленный один раз.
- Регрессионный различитель: на совместимой паре `compatible == true, errors == []` и инвариант
  цел; байты отчёта без `generated_at` не изменились бы — нормализация единственная (`generated_at`).

## Affected modules

| Модуль | Изменение |
|---|---|
| `internal/validate/domain` | `Report` + `Validator/Interaction/Consumer/GeneratedAt`; `Violation.Subject`; `Comparison.ConsumerName` → `ComparisonOutcome.ConsumerName`; порт `Clock` переносится сюда (из head) |
| `internal/validate/compare` | `rules.go`: subject вычисляется одним кодом рядом с `at` (UPPER-метод), проставляется во всех 5 ветках; `comparison.go`/`compare.go` несут `ConsumerName` |
| `internal/validate/report` | `FoldReport(outcome)` → `BuildReporter(clock) Reporter` + `Reporter.Fold(outcome) Report` (правило одного входа; единственный `Now()` в слайсе — D10) |
| `internal/validate` (head) | bind `reporter := report.BuildReporter(d.Clock)` до трубы; шаг `reporter.Fold(outcome)` |
| `cmd/app` | `errorReport`: `schema_version "1.1"` (конверт exit-3 остаётся минимальным, как у близнеца) |
| тесты | `report/build_test` (fixedClock, новые поля), `compare/comparison_test` (want + ConsumerName), `report` — countingClock-якорь (ровно 1 `Now()` на Fold), компонентные шаги: ассерты `schema_version=="1.1"`, `validator`, `interaction`, `consumer.name`, `generated_at` RFC3339, `errors[].subject`; новый сценарий «incompatible pair → exit 1» (дизайн-акт: формула 1+5 → 2+5) |
| `docs/design/slice-01-validate/{module-tree,contracts}.md` | канон-синк: узел `FoldReport` → два узла `BuildReporter`/`Reporter.Fold` (ADR-001 близнеца), формула сценариев, маркер `> Current as of change 001-report-schema-1.1 (lane minor)` |
| `README.md` | путь сборки `./cmd/app`; раздел «формат отчёта» → ссылка на `docs/report-format.md`; уточнение детерминизма (`generated_at` — единственная зависимость от времени, инжект портом) |

## design

`design = needed`: дельта затрагивает `module-tree.md`/`contracts.md` (узлы трубы) — после приёмки
сводится `mode=canon-sync` с маркером текущего состояния.

## Границы

Не меняются: плоская форма `errors[]` (перегруппировка в `verdicts[]` — отвергнутый вариант B),
exit-коды `0/1/2/3`, структура `provenance`, словарь кодов, `uncovered_operations`, алгоритм
`internal/validate/compare/`, `config.schema.json`, consumed-contract.

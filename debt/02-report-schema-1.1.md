# Тикет 2/2 — перевод выхода валидатора на `schema_version 1.1`

**Роль репозитория:** владелец канона (эпик E1).
**Родительский разбор:** `pinout/debt/report-canon-fork.md`.
**Зависит от:** тикет `01-report-canon-doc.md` — канон `docs/report-format.md` должен быть написан и
смержен. Он источник истины для этого тикета; расхождение с ним = дефект.

## Зачем

Отчёт валидатора не содержит идентичности сторон: нет имени потребителя, типа взаимодействия и
времени. `pinout-netlist` строит граф рёбер consumer↔provider и **физически не может** разложить
такой отчёт: ему нечем заполнить `Edge.Consumer`, `Edge.Interaction`, `Edge.Subject`,
`VerdictRecord.At`. Это функциональный провал интеграции, а не косметика.

## Дельта (аддитивно, `additionalProperties: false` сохраняется)

| поле | статус | источник значения в коде |
|---|---|---|
| `schema_version` | `const "1.0"` → `const "1.1"` | `FoldReport` |
| `validator` | required, `const "pinout-openapi"` | константа сборки |
| `interaction` | required, `const "sync"` | константа сборки |
| `consumer.name` | required | `Config.ConsumerName` — **уже есть** |
| `consumer.version` | optional | источника нет → не эмитим |
| `generated_at` | required, RFC3339 | **инжектится вызывающим** (порт часов) |
| `errors[].subject` | required, `METHOD /path` | **выделить** из вычисления `location` |

## Работы

- [ ] `api-specification/report.schema.json` → `1.1`, `x-frozen` перештамповать, `x-twin`-сверка с
      async-близнецом сохраняется;
- [ ] `internal/validate/domain/domain.go` — `Report` (+`Validator`, `Interaction`, `Consumer`,
      `GeneratedAt`), `Violation` (+`Subject`);
- [ ] `internal/validate/report/build.go` (`FoldReport`) — заполнение новых полей;
- [ ] `subject` **выделяется** из кода, который сейчас собирает префикс `location`; не дублировать
      строку руками — иначе они разъедутся;
- [ ] **шов часов** для `generated_at` — порт времени, инжектится сверху. Без него компонентные
      тесты станут недетерминированными; это единственная неаддитивная часть тикета;
- [ ] `consumer.name` протянуть из `Config` до `FoldReport` (сейчас туда не доходит — проверить
      цепочку `Comparison`/`ComparisonOutcome`);
- [ ] обновить фикстуры компонентных тестов — обновление ожидаемо, регрессией не считать;
- [ ] README: раздел «формат отчёта» → ссылка на `docs/report-format.md`.

## Границы изменения

Не меняются: плоская форма `errors[]` — перегруппировка в `verdicts[]` это отвергнутый вариант B;
exit-коды `0/1/2/3`; структура `provenance`; словарь кодов ошибок; `uncovered_operations`; алгоритм
сравнения `internal/validate/compare/`; `config.schema.json` и `consumed-contract`.

## Приёмка

- `report.schema.json` — `schema_version: "1.1"`, все поля дельты, `additionalProperties: false` цел;
- `pinout-openapi validate` печатает `1.1` с непустыми `validator` / `interaction` /
  `consumer.name` / `generated_at` / `errors[].subject`;
- инвариант `compatible ⇔ errors == []` не нарушен;
- `go build ./...`, `go test ./...`, компонентные тесты, CI — зелёные;
- выход **побайтово соответствует** `docs/report-format.md` из тикета 1/2.

# Тикет 1/2 — канон формата отчёта `docs/report-format.md` (версия 1.1)

**Приоритет:** 🔴 №1 в экосистеме. **Роль репозитория:** владелец канона (эпик E1).
**Родительский разбор:** `pinout/debt/report-canon-fork.md` (решение оператора — вариант A, 2026-07-25).
**Блокирует:** тикет 2/2 этого репо, E0 (`pinout-asyncapi`), E2 (`pinout-netlist`).
**Не зависит ни от чего.** Кормить харнесу первым.

## Зачем

`docs/report-format.md` — канон формата отчёта для всей экосистемы — удалён коммитом `2e0c232`
(«clean slate for harness run»). Пока его не было, два валидатора заморозили собственные выходные
контракты, а `pinout-netlist` проектировался под третью форму. Нужен единый письменный источник
истины, против которого правятся оба валидатора и потребитель.

**Восстанавливать файл из git запрещено.** Тот `1.0` описывает форму
(`verdicts[]{subject, compatible, errors[]}`, `provider{}`, без `provenance`), которую **не
реализовал ни один валидатор**, и носит тот же номер версии, что и фактически замороженная плоская
форма. Возврат воскресит конфликт. Канон пишется **заново от фактической формы** и версией вперёд.

## Фактическая форма (то, что реально печатается сегодня)

`api-specification/report.schema.json`, `x-frozen: 2026-07-16`, `schema_version: const "1.0"`,
`additionalProperties: false`:

```text
{schema_version, compatible,
 provenance{provider, provider_version, captured_hash},
 errors[]{code, message, location, details, context},
 uncovered_operations[]}
```

Близнец `pinout-asyncapi/api-specification/report.schema.json` (`x-frozen: 2026-07-25`) — та же
форма, отличается только enum кодов и `uncovered_channels[]` вместо `uncovered_operations[]`.

## Что написать

Канон версии **1.1** = фактическая форма + аддитивная надстройка:

| поле | статус | значение |
|---|---|---|
| `schema_version` | было `"1.0"` → стало `"1.1"` | — |
| `validator` | required | `pinout-openapi` \| `pinout-asyncapi` |
| `interaction` | required | `sync` \| `async` |
| `consumer.name` | required | имя потребителя из его конфига |
| `consumer.version` | optional | эмитится только если известна |
| `generated_at` | required, RFC3339 | проставляет вызывающий |
| `errors[].subject` | required | sync: `METHOD /path`; async: `<channel> <direction> <message>` |

Обязательные разделы документа:

- [ ] схема `1.1` целиком с описанием каждого поля;
- [ ] таблица `x-exit-codes` (`0` compatible / `1` incompatible / `2` config / `3` io-parse) —
      уже единая у обоих валидаторов, переносится как есть;
- [ ] словари кодов ошибок: sync (`OP_NOT_IN_PROVIDER`, `MISSING_REQUIRED_REQUEST_FIELD`,
      `READS_FIELD_NOT_PROVIDED`, `TYPE_MISMATCH`) и async (`R1`–`R9`), общие io/parse
      (`PARSE_ERROR`, `FILE_NOT_FOUND`, `HTTP_ERROR`, `TIMEOUT_ERROR`);
      `CONFIG_ERROR` (exit 2) в `errors[]` не попадает — фиксируется явно;
- [ ] инвариант `compatible ⇔ errors == []`;
- [ ] **`provenance` объявляется единственным источником идентичности поставщика** — она честнее
      старого `provider{}`, потому что несёт хеш захвата. Без этой записи netlist будет искать
      `provider{}` из мёртвой спеки;
- [ ] почему нет `verdicts[]`: субъектная гранулярность даётся полем `errors[].subject`, netlist
      строит рёбра по нему; перегруппировка — отвергнутый вариант B;
- [ ] политика версионирования: новое поле → minor, перегруппировка/удаление → major;
- [ ] в шапке `superseded: docs/report-format.md@2e0c232^` + одна строка, что тот `1.0` никогда не
      был реализован. Закрывает археологию для будущих агентов.

## Границы

Только `docs/report-format.md`. **Не трогать** `api-specification/`, `internal/`, тесты — это
тикет 2/2. Канон описывает целевую форму; приведение кода к ней — следующий заход.

## Приёмка

- файл существует, описывает `1.1`, самосогласован (форма в тексте = таблица дельты);
- ни одной битой ссылки внутри;
- документ достаточен, чтобы `pinout-asyncapi` и `pinout-netlist` правились **по нему одному**,
  без чтения кода этого репозитория;
- md-формат по скиллу `md-formatting`.

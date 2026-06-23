# Покрытие рисков сборки и валидации OpenAPI

Каждый риск конвейера `pinout-openapi` — от чтения спеки до вердикта совместимости — с закрывающим инструментом/проверкой и опорой в спецификации OpenAPI. Связано: словарь `ErrorCode` и `ErrorCode→exit` — [`design/contract-validate/messages.md`](./design/contract-validate/messages.md) / [`contracts-graph.md`](./design/contract-validate/contracts-graph.md); обработка `$ref` — [`design/contract-validate/ref-handling.md`](./design/contract-validate/ref-handling.md).

Разделение ответственности: структурную валидность **одного** документа закрывает `kin-openapi` (`Load` + `Validate`); совместимость **пары** — чистая логика (`compareSchemas` и подфункции); семантику использования контракта — компонентные тесты потребителя. Инструмент не дублирует первый и третий слой.

## A. Сборка спеки (чтение / парсинг / резолв)

| Риск | Чем закрывается | Опора в OpenAPI |
|---|---|---|
| Битый YAML/JSON-синтаксис | `kin-openapi` `LoadFromData/File` → `SPEC_PARSE_ERROR` | OpenAPI Document — это JSON/YAML |
| Не OpenAPI-документ (нет `info`/`paths`, кривые поля) | `doc.Validate(ctx)` | OpenAPI Object: required `openapi`, `info`, `paths` |
| Неподдерживаемая версия (Swagger 2.0 / 3.1↔3.0) | проверка поля `openapi: "3.x"` до обхода | Version String |
| Локальный `$ref` указывает в никуда (`#/components/schemas/X` нет) | резолвер `kin-openapi` → ошибка загрузки | Reference Object (`$ref`) |
| Внешний `$ref` (`./common.yaml#/X`, URL) | `IsExternalRefsAllowed=false` → `SPEC_PARSE_ERROR` (MVP, hard-fail) | Reference Object (relative/URI refs) |
| Рекурсивный `$ref` → зависание обхода | детект цикла памятью пар (`ref-handling.md`, слайс C) | Schema Object допускает рекурсивные ссылки |
| Спека недостижима / нет файла | `SpecClient.Fetch` → `SPEC_UNREACHABLE` / `SPEC_NOT_FOUND` | спека отдаётся по `spec_url`/`spec_path` (конфиг) |

## B. Валидация совместимости пары (consumer ↔ provider)

| Риск | Чем закрывается | Опора в OpenAPI |
|---|---|---|
| Шаблоны путей расходятся (`/users/{id}` vs `/users/{userId}`) | нормализация шаблона по позициям параметров | Path Templating |
| Операции из конфига нет у потребителя | `NewContractValidate` (антецедент) → `CONSUMER_OPERATION_NOT_FOUND` (exit 2) | Paths / Path Item / Operation (HTTP-глаголы) |
| Операции нет у поставщика | `findProviderOperation` → `OPERATION_NOT_FOUND` (exit 1) | Paths / Path Item / Operation |
| Поставщик требует обязательное поле/параметр, которого нет у потребителя | `compareRequest` → `REQUEST_INCOMPATIBLE` | Parameter (`required`), Request Body, Schema `required` |
| Потребитель ждёт поле, которого нет в ответе поставщика | `compareResponses` (provider ⊇ consumer) → `RESPONSE_INCOMPATIBLE` | Responses / Media Type / Schema |
| Код ответа потребителя поставщик не отдаёт | `compareStatusCodes` → `STATUS_MISMATCH` | Responses Object (status codes) |
| Рассинхрон content-type | `compareContentTypes` → `CONTENT_TYPE_MISMATCH` | Media Type Object (ключи `content`) |
| Несовместимость схем (тип/required, вложенность) | `compareSchemas` рекурсивно по `schema.Value` | Schema Object (`type`, `properties`, `required`, `items`) |

## C. Границы — закрывается НЕ инструментом (осознанно)

| Риск | Чем закрывается | Опора в OpenAPI |
|---|---|---|
| `allOf/oneOf/anyOf/discriminator`, сужение `enum`, `format`/числовые ограничения | out-of-scope `compareSchemas` → компонентные тесты потребителя (слой методологии), непроверенное логируется | Schema composition / Discriminator Object |
| Сервис не соответствует собственной спеке (спека ≠ поведение) | компонентные тесты самого сервиса (несущий инвариант экосистемы) | вне OpenAPI — слой методологии |
| Конфиг пары битый / `OperationKey` невалиден | `ConfigStore.Load`+`NewConfig` (`CONFIG_*`), `NewOperationKey` | путь начинается с `/`, метод ∈ HTTP-глаголов |

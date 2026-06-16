# api-specification

Контракт инструмента `pinout-openapi` — это CLI (`validate <config>`) и схема `contract-tests.yaml` (см. корневой [README](../README.md)).

Сюда кладётся OpenAPI-спека потребителя, валидируемого в локальных компонентных тестах (фикстуры), либо ссылки на спеки в `testdata/`. Для самого инструмента (CLI) отдельной OpenAPI-спеки нет — это не сетевой сервис.

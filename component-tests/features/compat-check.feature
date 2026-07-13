# language: ru
# Component scenarios — slice compat-check (ticket-02, wirth-tester REALIZE).
# Design source (invent nothing new): docs/design/slice-compat-check/contracts.md
# §"Component scenarios" (formula N = 1 happy + Σ 7 distinguishable I/O-adapter
# branches = 8, + CS2 as a 9th success-family verdict guard) + use-case.md.
# All 9 scenarios tagged @wip — RED by business reason (placeholder head returns
# NOT_IMPLEMENTED/exit 3) until module tickets 03-08 land. @wip is removed ONLY
# by @fagan at slice acceptance. OP_NOT_IN_CONSUMER(3a) and INCOMPATIBLE
# rule-boundaries are unit-level (contracts.md anti-gaming 7==7==7) — NOT here.
Функционал: Compat-check — совместимость consumer↔provider OpenAPI (чёрный ящик)

  CLI-компонентные тесты: static-бинарь, one-shot запуск `run <config>` против
  фикстур в /fixtures, проверка (код выхода, JSON на stdout) как чёрного ящика.
  Провайдерская сторона (CS6/CS7/CS8) обслуживается реальным HTTP-стабом
  provider-stub (compose), а не in-code моком.

  Сценарий: CS1 все настроенные операции совместимы (happy) + детерминированный отчёт
    Дано используем фикстуру "cs1-happy"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 0
    И JSON-поле "verdict" в stdout равно "compatible"
    И все операции в отчёте совместимы (status="compatible", findings пуст)
    И файл отчёта записан по пути "/tmp/reports/cs1-report.json"
    И повторный запуск на той же фикстуре и файл отчёта "/tmp/reports/cs1-report.json" побайтово идентичен первому без учёта поля "generated_at"

  Сценарий: CS2 хотя бы одна операция нарушает правило совместимости
    Дано используем фикстуру "cs2-incompatible"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 1
    И JSON-поле "verdict" в stdout равно "incompatible"
    И хотя бы одна операция в отчёте несовместима с findings, содержащими правило "presence"

  Сценарий: CS3 конфиг отсутствует / невалиден / конфликт / пустые operations
    Дано используем фикстуру "cs3-config-invalid"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 2
    И вывод содержит "CONFIG_INVALID"

  Сценарий: CS4 спецификация консьюмера отсутствует или недоступна
    Дано используем фикстуру "cs4-consumer-unreadable"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 2
    И вывод содержит "SPEC_UNREADABLE"

  Сценарий: CS5 спецификация консьюмера не является валидным OpenAPI 3.x
    Дано используем фикстуру "cs5-consumer-parse-error"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 2
    И вывод содержит "SPEC_PARSE_ERROR"

  Сценарий: CS6 спецификация провайдера недоступна / HTTP не-2xx
    Дано используем фикстуру "cs6-provider-unreachable"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 3
    И вывод содержит "PROVIDER_UNREACHABLE"

  Сценарий: CS7 получение спецификации провайдера превышает settings.timeout
    Дано используем фикстуру "cs7-provider-timeout"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 3
    И вывод содержит "PROVIDER_TIMEOUT"

  Сценарий: CS8 спецификация провайдера не является валидным OpenAPI 3.x
    Дано используем фикстуру "cs8-provider-parse-error"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 3
    И вывод содержит "PROVIDER_PARSE_ERROR"

  Сценарий: CS9 JSON-файл отчёта невозможно записать
    Дано используем фикстуру "cs9-report-write-error"
    Когда запускаем инструмент с этим конфигом
    Тогда код выхода 3
    И вывод содержит "REPORT_WRITE_ERROR"
    И провайдерская спецификация на диске не изменилась

# CLAUDE.md — pinout-openapi

Валидатор синхронных контрактов OpenAPI. Концепт: [../pinout/README.md](../pinout/README.md). Эпик E1 в [бэклоге экосистемы](../pinout/backlog.md).

## Методология: скиллы service-template (по ссылке, не копируем)

Разработка ведётся строго по скиллам репозитория [ubik-life/service-template](https://github.com/ubik-life/service-template/). Источник — upstream, локальных копий не держим. Применяем в порядке:

| Этап | Скилл | Ссылка |
|---|---|---|
| Документация | `documentation` | https://github.com/ubik-life/service-template/tree/main/skills/documentation |
| Проверка качества доков | `doc-quality-review` | https://github.com/ubik-life/service-template/tree/main/skills/doc-quality-review |
| Проектирование | `program-design` | https://github.com/ubik-life/service-template/tree/main/skills/program-design |
| Компонентные тесты | `component-tests` | https://github.com/ubik-life/service-template/tree/main/skills/component-tests |
| Реализация | `program-implementation` | https://github.com/ubik-life/service-template/tree/main/skills/program-implementation |

### Ключевые правила (выжимка, источник — скиллы выше)

- **Контракт-первый, TDD.** Сценарии (Gherkin) пишутся до кода; режимы отказа берутся из OpenAPI/README, не реверсятся из кода.
- **Vertical slice.** Один внешний вход = один срез: ингресс-адаптер → головной модуль (pipe) → логика → I/O-модули.
- **Бизнес-логика ≠ I/O.** Логика — чистые функции и конструкторы; I/O (файлы, HTTP) изолирован в объектах (`Store`, `Client`, `Writer`). Никаких сырых `*http.Client`/`os` в зависимостях модулей логики.
- **Только конструкторы, без guard-функций.** Невалидный вход → структура не собирается (`NewT(raw) → (T, error)`).
- **Головной модуль — труба.** Линейная последовательность шагов через `Result<T, Error>` (ROP, как в `pinout-asyncapi`); без ветвления внутри трубы.
- **Юнит-тесты — только для логики/конструкторов** (`N = 1 happy + Σ ветвей antecedent`). Головной модуль, ингресс и I/O проверяются компонентными тестами.

## Состояние сессии

- Этап 0–1 экосистемы: концепт и верхнеуровневый бэклог — готово (в `../pinout`).
- Этот репозиторий: создан каркас + пакет проектирования MVP `docs/design/contract-validate/`. Реализация **не начата** — ждёт handoff-аппрува оператора в `docs/design/contract-validate/backlog.md`.

## Зависимости (план)

- Парсер OpenAPI 3.x — зрелая библиотека (honest reuse), напр. `kin-openapi`; не переизобретаем валидацию схем.
- Общий формат JSON-отчёта согласован с `pinout-asyncapi` (эпик E0) для потребления `pinout-netlist`.

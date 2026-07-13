#!/usr/bin/env bash
# Запуск компонентных тестов template-go-cli ВНУТРИ Docker (изоляция; не `go test` с хоста).
# tool стейджит бинарь в общий том и выходит; tester гоняет godog против него.
set -euo pipefail
cd "$(dirname "$0")/.."
CF=(-f docker-compose.test.yml)
cleanup() { docker compose "${CF[@]}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT
echo "==> building images..."; docker compose "${CF[@]}" build
# Внимание: --exit-code-from tester НЕ равнозначен «слушать только tester» — без
# явного имени сервиса `up` наблюдает abort-on-container-exit на ВСЕХ сервисах,
# включая one-shot `tool` (стейджер, выходит 0 почти сразу после старта). Это
# гонится с временем поднятия provider-stub и валит стек ДО того, как tester
# вообще успевает стартовать (wiring-red, не бизнес-причина). Явно передаём
# `tester` — тогда compose поднимает его зависимости (tool, provider-stub)
# автоматически, но abort-триггер слушает только сам tester.
echo "==> running tests..."; docker compose "${CF[@]}" up --abort-on-container-exit --exit-code-from tester tester

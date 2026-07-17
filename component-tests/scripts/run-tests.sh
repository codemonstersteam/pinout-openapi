#!/usr/bin/env bash
# Запуск компонентных тестов template-go-cli ВНУТРИ Docker (изоляция; не `go test` с хоста).
# tool стейджит бинарь в общий том и выходит; tester гоняет godog против него.
#
# ВАЖНО (scenario 6, TIMEOUT_ERROR): provider-stub ДОЛЖЕН пережить tester — сценарий 6
# упирается в /stall (стаб спит 5s), а SUT со своим settings.timeout=1 ждёт ~1s живого
# стаба, чтобы сработал реальный дедлайн. Прежний `up --abort-on-container-exit
# --exit-code-from tester` рвал ВСЁ в момент выхода one-shot-стейджера `tool`: стаб
# останавливался раньше, чем tester доходил до сценария 6, и запрос падал в
# connection-refused → HTTP_ERROR вместо TIMEOUT_ERROR (гонка тир-дауна, не баг SUT).
# Поэтому: стаб поднимаем detached (живёт), tool стейджит бинарь (depends_on его
# ждёт), а результатом прогона служит КОД ВЫХОДА самого tester (`run --rm tester`).
set -euo pipefail
cd "$(dirname "$0")/.."
CF=(-f docker-compose.test.yml)
cleanup() { docker compose "${CF[@]}" down -v --remove-orphans >/dev/null 2>&1 || true; }
trap cleanup EXIT
echo "==> building images..."; docker compose "${CF[@]}" build
echo "==> starting provider-stub (must outlive the tester)..."; docker compose "${CF[@]}" up -d --no-build provider-stub
echo "==> running tests (tester exit code IS the suite result)..."; docker compose "${CF[@]}" run --rm tester

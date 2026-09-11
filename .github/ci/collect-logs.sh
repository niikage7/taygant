#!/usr/bin/env bash
# Сохраняет состояние и логи всех контейнеров стека в каталог $1 — оттуда их
# забирает upload-artifact. С флагом --print ещё и печатает их в лог шага
# сворачиваемыми группами, чтобы причину падения было видно без скачивания.
#
# Стек выбирается переменными docker compose из окружения
# (COMPOSE_PROJECT_NAME, COMPOSE_FILE, JWT_SECRET).
set -uo pipefail

dir=$1
print=${2:-}
mkdir -p "$dir"

docker compose ps -a >"$dir/compose-ps.txt" 2>&1
docker compose logs --no-color --timestamps >"$dir/all-services.log" 2>&1

for service in $(docker compose config --services); do
  docker compose logs --no-color --timestamps "$service" >"$dir/$service.log" 2>&1
  container=$(docker compose ps -a -q "$service" 2>/dev/null | head -n 1)
  if [ -n "$container" ]; then
    # Журнал healthcheck'а объясняет, почему контейнер так и не стал healthy.
    docker inspect --format \
      '{{.State.Status}} (exit {{.State.ExitCode}}, перезапусков {{.RestartCount}}){{if .State.Health}}{{range .State.Health.Log}}
{{.Start}} код {{.ExitCode}}: {{.Output}}{{end}}{{end}}' \
      "$container" >"$dir/$service-state.txt" 2>&1
  fi
done

if [ "$print" = "--print" ]; then
  echo "::group::📋 Состояние контейнеров"
  cat "$dir/compose-ps.txt"
  echo "::endgroup::"
  for service in $(docker compose config --services); do
    echo "::group::📜 $service — состояние и healthcheck"
    cat "$dir/$service-state.txt" 2>/dev/null || echo "контейнер не создан"
    echo "::endgroup::"
    echo "::group::📜 $service — последние 300 строк лога"
    tail -n 300 "$dir/$service.log"
    echo "::endgroup::"
  done
fi

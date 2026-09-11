#!/usr/bin/env bash
#
# Статус стека taygant: запущенные контейнеры + смоук-тест endpoints.
#
# Использование: astra/run/status.sh

set -uo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

command -v docker >/dev/null 2>&1 || { echo "docker не найден."; exit 1; }

echo "== Контейнеры =="
docker ps -a --filter name=taygant- --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}' 2>/dev/null

echo
echo "== Смоук-тест =="
ok=0; bad=0
probe() { # name url
  if http_ok "$2"; then
    echo "  OK   $1  →  $2"
    ok=$((ok + 1))
  else
    echo "  FAIL $1  →  $2"
    bad=$((bad + 1))
  fi
}

if (exec 3<>/dev/tcp/127.0.0.1/5432) 2>/dev/null; then
  echo "  OK   postgres  →  tcp://127.0.0.1:5432"
  ok=$((ok + 1))
else
  echo "  FAIL postgres  →  tcp://127.0.0.1:5432"
  bad=$((bad + 1))
fi
probe backend  "http://127.0.0.1:4000/ping"
probe frontend "http://127.0.0.1:3000/login"

echo
if [ "$bad" -eq 0 ]; then
  echo "Стек работает. Frontend: http://127.0.0.1:3000"
  exit 0
else
  echo "Проблемы: $bad из $((ok + bad)). Логи: docker logs -f taygant-<сервис>"
  exit 1
fi
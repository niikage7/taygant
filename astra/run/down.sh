#!/usr/bin/env bash
#
# Останавливает и удаляет контейнеры taygant. Данные БД сохраняются
# в $DATA_DIR/postgres и не удаляются.
#
# Полный сброс данных:  rm -rf "$APP_ROOT/data/postgres"
#
# Использование: astra/run/down.sh

set -uo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

command -v docker >/dev/null 2>&1 || { echo "docker не найден — стоп не требуется."; exit 0; }

for c in taygant-frontend taygant-backend taygant-postgres; do
  if container_exists "$c"; then
    docker rm -f "$c" >/dev/null 2>&1 && echo "остановлен: $c"
  fi
done

echo "Стек остановлен. Данные БД: $DATA_DIR/postgres"
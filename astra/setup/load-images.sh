#!/usr/bin/env bash
#
# Загружает оффлайн-архив образов (результат offline/export-images.sh)
# в rootless Docker.
#
# Использование:
#   astra/setup/load-images.sh /path/to/images.tar.gz
#
# Проверка: docker images | grep taygant

set -euo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

ARCHIVE="${1:-$APP_ROOT/images.tar.gz}"
[ -f "$ARCHIVE" ] || die "архив не найден: $ARCHIVE (соберите его на машине с интернетом: offline/export-images.sh)"

echo "Загрузка образов из $ARCHIVE …"
if [[ "$ARCHIVE" == *.gz || "$ARCHIVE" == *.tgz ]]; then
  zcat "$ARCHIVE" | docker load
else
  docker load -i "$ARCHIVE"
fi

echo
echo "Образы:"
docker images | grep -E 'REPOSITORY|taygant' || true
echo
echo "Дальше: astra/keys/sign-images.sh (подпись для ЗПС) или сразу astra/run/up.sh (без ЗПС)."
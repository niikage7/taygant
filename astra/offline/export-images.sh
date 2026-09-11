#!/usr/bin/env bash
#
# Собирает образы taygant для Astra Linux и упаковывает их в оффлайн-архив.
# Запускать НА МАШИНЕ С INTERNET, где есть Docker и есть исходники проекта.
#
# Образ фронтенда для Astra собирается отдельным тегом: rewrites Next.js
# зашиваются на этапе сборки, и для host-сети rootless-контейнеров бэкенд
# доступен по 127.0.0.1:4000 (а не по имени сервиса compose).
#
# Использование:
#   astra/offline/export-images.sh [каталог_вывода]     (по умолчанию <репозиторий>/astra-offline)
#
# Результат: <вывод>/images.tar.gz  → перенести на целевую Astra и выполнить
#            astra/setup/load-images.sh images.tar.gz

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
OUT="${1:-$ROOT/astra-offline}"

command -v docker >/dev/null 2>&1 || { echo "Docker не найден." >&2; exit 1; }

mkdir -p "$OUT"

echo "==> [1/3] Сборка образа backend (taygant-backend:astra) …"
docker build -t taygant-backend:astra "$ROOT/backend"

echo "==> [2/3] Сборка образа frontend с BACKEND_URL=http://127.0.0.1:4000 (taygant-frontend:astra) …"
docker build \
  --build-arg BACKEND_URL=http://127.0.0.1:4000 \
  -t taygant-frontend:astra \
  "$ROOT/frontend"

echo "==> [3/3] PostgreSQL (postgres:17-alpine) …"
docker pull postgres:17-alpine
docker tag postgres:17-alpine taygant-postgres:astra

echo "==> Сохранение образов в архив …"
docker save \
  taygant-postgres:astra \
  taygant-backend:astra \
  taygant-frontend:astra \
  | gzip > "$OUT/images.tar.gz"

size="$(du -h "$OUT/images.tar.gz" | cut -f1)"
echo
echo "Готово: $OUT/images.tar.gz ($size)"
echo "Перенесите архив на целевую Astra и выполните:"
echo "  astra/setup/load-images.sh images.tar.gz"
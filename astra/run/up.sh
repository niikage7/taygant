#!/usr/bin/env bash
#
# Запуск стека taygant на Astra Linux (rootless Docker, host-сеть).
#
# Сеть — host (--network=host): все три контейнера живут в сетевом
# пространстве хоста и общаются через 127.0.0.1. Это обходит ограничения
# rootless-моста и избавляет от DNS-имён сервисов. Фронтенд, собранный
# для Astra, проксирует /backend/* на http://127.0.0.1:4000.
#
# Образы запускаются как обычно. Для работы под ЗПС их содержимое должно
# быть подписано отсоединёнными подписями (keys/sign-images.sh) — при
# включённой ЗПС неподписанные процессы ядро просто не запустит.
#
# Использование: astra/run/up.sh   (запускать из-под пользователя rootless-Docker)

set -euo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

command -v docker >/dev/null 2>&1 || die "docker не найден — сначала setup/install-rootless-docker.sh"
docker info >/dev/null 2>&1 || die "rootless Docker не отвечает — см. README → «Диагностика»"

[ -n "${JWT_SECRET:-}" ] || die "JWT_SECRET не задан — создайте $APP_ROOT/.env.astra по образцу astra/.env.astra.example"

# Предупреждение о подписи (не блокирующее: без ЗПС стек работает и без неё).
if [ ! -f "$SIG_HOST_DIR/.taygant-backend" ]; then
  echo "ВНИМАНИЕ: подписи ЗПС не найдены ($SIG_HOST_DIR)."
  echo "  Для работы при включённой ЗПС выполните: astra/keys/sign-images.sh"
  echo "  (без ЗПС стек запустится и так.)"
fi

mkdir -p "$DATA_DIR/postgres"

# ── PostgreSQL ────────────────────────────────────────────────────────────
if container_running taygant-postgres; then
  echo "postgres уже запущен"
else
  container_exists taygant-postgres && docker rm -f taygant-postgres >/dev/null 2>&1
  echo "запуск postgres …"
  docker run -d --name taygant-postgres --network=host \
    -v "$DATA_DIR/postgres:/var/lib/postgresql/data" \
    -e POSTGRES_USER=taygant \
    -e POSTGRES_PASSWORD=taygant \
    -e POSTGRES_DB=taygant \
    "$TAYGANT_IMAGE_POSTGRES"
fi
echo "ожидание postgres (127.0.0.1:5432) …"
wait_tcp 127.0.0.1 5432 60 || die "postgres не поднялся — смотрите: docker logs taygant-postgres"

# ── Backend ───────────────────────────────────────────────────────────────
if container_running taygant-backend; then
  echo "backend уже запущен"
else
  container_exists taygant-backend && docker rm -f taygant-backend >/dev/null 2>&1
  echo "запуск backend …"
  docker run -d --name taygant-backend --network=host \
    -e APP_PORT=4000 \
    -e "DATABASE_URL=postgres://taygant:taygant@127.0.0.1:5432/taygant?sslmode=disable" \
    -e "JWT_SECRET=$JWT_SECRET" \
    -e ACCESS_TOKEN_TTL=24h \
    -e REFRESH_TOKEN_TTL=720h \
    -e SEED_DEMO_DATA=true \
    -e 'ALLOWED_ORIGINS=http://localhost:3000' \
    -e RATE_LIMIT_MAX=300 \
    -e RATE_LIMIT_WINDOW=1m \
    "$TAYGANT_IMAGE_BACKEND"
fi
echo "ожидание backend (127.0.0.1:4000/ping) …"
wait_http 127.0.0.1 4000 /ping 90 || die "backend не поднялся — смотрите: docker logs taygant-backend"

# ── Frontend ──────────────────────────────────────────────────────────────
if container_running taygant-frontend; then
  echo "frontend уже запущен"
else
  container_exists taygant-frontend && docker rm -f taygant-frontend >/dev/null 2>&1
  echo "запуск frontend …"
  docker run -d --name taygant-frontend --network=host \
    -e NODE_ENV=production \
    -e PORT=3000 \
    -e HOSTNAME=0.0.0.0 \
    -e NEXT_TELEMETRY_DISABLED=1 \
    "$TAYGANT_IMAGE_FRONTEND"
fi
echo "ожидание frontend (127.0.0.1:3000/login) …"
wait_http 127.0.0.1 3000 /login 120 || die "frontend не поднялся — смотрите: docker logs taygant-frontend"

echo
echo "Стек запущен:"
echo "  Frontend:  http://127.0.0.1:3000   (с рабочей станции — http://<IP_ASTRA>:3000)"
echo "  Backend:   http://127.0.0.1:4000/ping"
echo "  PostgreSQL: 127.0.0.1:5432 (taygant/taygant)"
echo
echo "Логи:  docker logs -f taygant-frontend | taygant-backend | taygant-postgres"
echo "Статус: astra/run/status.sh   Стоп: astra/run/down.sh"
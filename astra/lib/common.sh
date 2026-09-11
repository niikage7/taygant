#!/usr/bin/env bash
#
# Общие пути, настройки и функции для скриптов taygant под Astra Linux SE 1.8
# (режим «Воронеж», ЗПС + PARSEC). Подключается из остальных скриптов:
#     . "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"
#
# Переменные можно переопределить через файл .env.astra:
#     APP_ROOT/.env.astra   (по умолчанию APP_ROOT=~/taygant)
#     <репозиторий>/astra/.env.astra

set -uo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SELF_DIR/.." && pwd)"

# --- базовый каталог развёртывания ---
: "${APP_ROOT:=$HOME/taygant}"

# .env.astra грузится до вычисления производных путей: он может переопределить APP_ROOT.
for _env_file in "$APP_ROOT/.env.astra" "$PROJECT_ROOT/.env.astra"; do
  if [ -f "$_env_file" ]; then
    set -a
    # shellcheck disable=source
    . "$_env_file"
    set +a
  fi
done

# --- пути ---
: "${APP_ROOT:=$HOME/taygant}"
: "${DATA_DIR:=$APP_ROOT/data}"          # данные БД и т.п.
: "${STAGING_DIR:=$APP_ROOT/staging}"    # рабочий каталог подписи образов
: "${KEYS_DIR:=$APP_ROOT/keys}"          # ключи подписи (секретный — только здесь)
: "${SIG_HOST_DIR:=/etc/digsig/external_sig}"   # отсоединённые подписи (читает ядро digsig_verif)

# --- имена образов (совпадают с именами из offline/export-images.sh) ---
: "${TAYGANT_IMAGE_POSTGRES:=taygant-postgres:astra}"
: "${TAYGANT_IMAGE_BACKEND:=taygant-backend:astra}"
: "${TAYGANT_IMAGE_FRONTEND:=taygant-frontend:astra}"

# --- подпись файлов (keys/sign-images.sh) ---
# Для файлов внутри контейнеров используется ТОЛЬКО отсоединённая подпись
# (в /etc/digsig/external_sig): docker export/import не переносит xattr,
# а bind-mount в корень контейнера Docker запрещает. Подпись в расширенных
# атрибутах sign-files.sh может делать для обычных файлов на хосте.
: "${SIGN_XATTR:=0}"      # подпись в расширенных атрибутах (только для файлов хоста)
: "${SIGN_DETACHED:=1}"   # отсоединённая подпись (обязательна для контейнеров)

# Идентификатор секретного ключа подписи (создаётся keys/gen-keys.sh).
KEYID="${TAYGANT_KEYID:-}"
if [ -z "$KEYID" ] && [ -f "$KEYS_DIR/KEYID" ]; then
  KEYID="$(cat "$KEYS_DIR/KEYID")"
fi

# ---------------------------------------------------------------------------
# Служебные функции
# ---------------------------------------------------------------------------

log() { printf '%s\n' "$*"; }

die() {
  printf 'ОШИБКА: %s\n' "$*" >&2
  exit 1
}

# is_signable FILE — подлежит ли файл подписи для ЗПС (ELF или скрипт).
# ЗПС проверяет исполняемые файлы (ELF) и скрипты с shebang.
is_signable() {
  local f="$1" t
  t="$(file -b -- "$f" 2>/dev/null || true)"
  case "$t" in
    *ELF*|*script*|*"text executable"*) return 0 ;;
  esac
  return 1
}

# sign_file FILE OUTDIR — подписать один файл (xattr и/или отсоединённо).
# OUTDIR — куда bsign кладёт отсоединённые подписи (только при SIGN_DETACHED=1).
sign_file() {
  local file="$1" outdir="$2"
  [ -n "$KEYID" ] || die "ключ подписи не найден — запустите keys/gen-keys.sh"
  command -v bsign >/dev/null 2>&1 \
    || die "bsign не установлен — см. offline/packages.list (пакет bsign)"

  local flags=(-s -N)
  if [ "${SIGN_XATTR:-0}" = 1 ]; then
    flags+=(-X)
  fi
  if [ "${SIGN_DETACHED:-1}" = 1 ]; then
    flags+=(-D -T "$outdir")
    mkdir -p "$outdir"
  fi

  local pgopt="--batch --no-tty --default-key=${KEYID}"
  if [ -n "${SIGN_PASSPHRASE:-}" ]; then
    pgopt="$pgopt --pinentry-mode=loopback --passphrase=${SIGN_PASSPHRASE}"
  fi

  bsign "${flags[@]}" --pgoptions="$pgopt" "$file"
}

# wait_tcp HOST PORT [TRIES] — ждать открытия TCP-порта (по умолчанию 60 попыток, шаг 1с).
wait_tcp() {
  local host="$1" port="$2" tries="${3:-60}" i=0
  while [ "$i" -lt "$tries" ]; do
    if (exec 3<>"/dev/tcp/$host/$port") 2>/dev/null; then
      exec 3>&- 3<&- || true
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done
  return 1
}

# http_ok URL — успешен ли HTTP-запрос (curl, запасной вариант — wget).
http_ok() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsS --max-time 3 -o /dev/null "$1" 2>/dev/null
  else
    wget -q -O /dev/null --timeout=3 "$1" 2>/dev/null
  fi
}

# wait_http HOST PORT PATH [TRIES] — ждать HTTP 200 по адресу.
wait_http() {
  local host="$1" port="$2" path="$3" tries="${4:-60}" i=0
  while [ "$i" -lt "$tries" ]; do
    if http_ok "http://$host:$port$path"; then
      return 0
    fi
    sleep 1
    i=$((i + 1))
  done
  return 1
}

# container_running NAME
container_running() {
  local name
  while IFS= read -r name; do
    [ "$name" = "$1" ] && return 0
  done < <(docker ps --format '{{.Names}}' 2>/dev/null)
  return 1
}

# container_exists NAME
container_exists() {
  local name
  while IFS= read -r name; do
    [ "$name" = "$1" ] && return 0
  done < <(docker ps -a --format '{{.Names}}' 2>/dev/null)
  return 1
}
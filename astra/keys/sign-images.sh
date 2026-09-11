#!/usr/bin/env bash
#
# Подписывает содержимое образов taygant для работы в ЗПС.
#
# Механизм — ОТСОЕДИНЁННЫЕ подписи (самогенерированный ключ ГОСТ Р 34.10-2012):
#   1) содержимое образа экспортируется в каталог staging;
#   2) каждый ELF-файл и скрипт подписывается (bsign -D) — подпись покрывает
#      только содержимое файла;
#   3) файлы подписей переименовываются по путям ВНУТРИ контейнера
#      (/usr/local/bin/app → @usr@local@bin@app.sign) и кладутся в
#      /etc/digsig/external_sig на ХОСТЕ;
#   4) ядро digsig_verif при запуске процесса в контейнере находит подпись
#      по пути файла (как его видит процесс) и проверяет содержимое.
#
# Сам образ при этом не меняется и запускается обычным `docker run` — это
# не зависит от того, переносит ли docker xattr/подписи внутри слоёв.
#
# Почему не xattr-подписи внутри образа: docker export/import не переносит
# расширенные атрибуты, а bind-mount в корень контейнера («-v dir:/») Docker
# запрещает. Отсоединённые подписи работают в любом случае.
#
# Использование:
#   astra/keys/sign-images.sh                       — все три образа
#   astra/keys/sign-images.sh backend               — один сервис
#
# Требует: docker (rootless), bsign, ключ из keys/gen-keys.sh, sudo для $SIG_HOST_DIR.

set -euo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

[ -n "$KEYID" ] || die "ключ подписи не найден — запустите keys/gen-keys.sh"
command -v bsign >/dev/null 2>&1 || die "bsign не установлен — см. offline/packages.list (пакет bsign)"
docker info >/dev/null 2>&1 || die "rootless Docker не работает — см. README → «Диагностика»"

SERVICE_FILTER="${1:-}"

image_of() {
  case "$1" in
    postgres) echo "$TAYGANT_IMAGE_POSTGRES" ;;
    backend)  echo "$TAYGANT_IMAGE_BACKEND" ;;
    frontend) echo "$TAYGANT_IMAGE_FRONTEND" ;;
    *) die "неизвестный сервис: $1 (ожидалось postgres|backend|frontend)" ;;
  esac
}

# install_detached SIGDIR ROOTFS NAME — переименовать отсоединённые подписи
# из имён по путям staging в имена по путям внутри контейнера и установить
# в $SIG_HOST_DIR.
install_detached() {
  local sigdir="$1" rootfs="$2" name="$3"
  local prefix n=0 sig base dest

  # Имя файла подписи = абсолютный путь файла с /→@ (так делает bsign -D).
  # В ${var#"pattern"} кавычки делают pattern литеральной строкой — так
  # безопасно снимаем префикс пути staging.
  prefix="$(printf '%s' "$rootfs" | sed 's#/#@#g')"       # @home@user@…@staging@postgres

  sudo mkdir -p "$SIG_HOST_DIR" || die "[$name] нет прав sudo для $SIG_HOST_DIR"
  for sig in "$sigdir"/*.sign; do
    [ -e "$sig" ] || continue
    base="$(basename "$sig")"                   # @home@…@postgres@usr@bin@foo.sign
    dest="${base#"$prefix"}"                    # @usr@bin@foo.sign
    dest="${dest#@}"                            # usr@bin@foo.sign
    sudo cp "$sig" "$SIG_HOST_DIR/@${dest}" || die "[$name] не удалось установить подпись $sig"
    n=$((n + 1))
  done
  # Маркер «сервис подписан» — для проверки в run/up.sh.
  sudo touch "$SIG_HOST_DIR/.taygant-${name}" 2>/dev/null || true
  echo "[$name] отсоединённых подписей установлено в $SIG_HOST_DIR: $n"
}

# sign_service NAME IMAGE
sign_service() {
  local name="$1" image="$2"
  local work="$STAGING_DIR/$name" sigdir="$STAGING_DIR/$name-sigs"
  rm -rf "$work" "$sigdir"
  mkdir -p "$work"

  echo
  echo "===== $name ($image) ====="

  echo "[$name] создание контейнера из образа …"
  local cid
  cid="$(docker create "$image")" || die "[$name] не удалось создать контейнер из $image"
  echo "[$name] экспорт ФС образа …"
  if ! docker export "$cid" | tar -x -C "$work" --no-same-owner; then
    docker rm -f "$cid" >/dev/null 2>&1 || true
    die "[$name] ошибка распаковки ФС образа"
  fi
  docker rm -f "$cid" >/dev/null 2>&1 || true

  # Корень ФС — сам каталог распаковки, либо единственный верхний подкаталог.
  local rootfs="$work"
  if { [ ! -d "$work/usr" ] && [ ! -d "$work/bin" ] && [ ! -d "$work/etc" ]; }; then
    local top
    top="$(find "$work" -mindepth 1 -maxdepth 1 -type d | head -n 1 || true)"
    if [ -n "$top" ]; then
      rootfs="$top"
      echo "[$name] корень ФС: ${rootfs#$work/}"
    fi
  fi

  echo "[$name] подпись ELF-файлов и скриптов (detached) …"
  mkdir -p "$sigdir" "$APP_ROOT/signatures"
  local n=0 rel
  rm -f "$APP_ROOT/signatures/$name.txt"
  while IFS= read -r -d '' f; do
    if is_signable "$f"; then
      sign_file "$f" "$sigdir" || die "[$name] не удалось подписать: $f"
      rel="${f#"$rootfs/"}"
      printf '%s\n' "$rel" >> "$APP_ROOT/signatures/$name.txt"
      n=$((n + 1))
    fi
  done < <(find "$rootfs" -xdev -type f -print0)
  echo "[$name] подписано файлов: $n"
  [ "$n" -gt 0 ] || die "[$name] подписано 0 файлов — проверьте экспорт образа"

  install_detached "$sigdir" "$rootfs" "$name"

  echo "[$name] готово"
}

if [ -n "$SERVICE_FILTER" ]; then
  echo "Подписываю только сервис: $SERVICE_FILTER"
  sign_service "$SERVICE_FILTER" "$(image_of "$SERVICE_FILTER")"
  echo
  echo "Готово. Проверка: astra/keys/verify-signatures.sh"
else
  for name in postgres backend frontend; do
    sign_service "$name" "$(image_of "$name")"
  done
  echo
  echo "Все сервисы подписаны."
  echo "  Отсоединённые подписи: $SIG_HOST_DIR"
  echo
  echo "Дальше: astra/setup/register-keys.sh → astra/setup/enable-zps.sh → перезагрузка → astra/run/up.sh"
fi
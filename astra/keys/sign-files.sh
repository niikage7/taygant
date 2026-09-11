#!/usr/bin/env bash
#
# Подписывает произвольные файлы/каталоги для ЗПС (ELF и скрипты).
# Тип подписи задаётся в .env.astra:
#   SIGN_DETACHED=1 — отсоединённая подпись (в $APP_ROOT/signatures);
#   SIGN_XATTR=1    — дополнительно подпись в расширенных атрибутах.
#
# Использование:
#   astra/keys/sign-files.sh <файл|каталог> [<файл|каталог> …]
#
# Для подписи образов используйте sign-images.sh, а не этот скрипт.

set -euo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

[ $# -gt 0 ] || die "использование: sign-files.sh <файл|каталог> [<файл|каталог> …]"
[ -n "$KEYID" ] || die "ключ подписи не найден — запустите keys/gen-keys.sh"

OUTDIR="$APP_ROOT/signatures"
mkdir -p "$OUTDIR"

count=0
for path in "$@"; do
  [ -e "$path" ] || { echo "пропуск (не существует): $path" >&2; continue; }
  while IFS= read -r -d '' f; do
    if is_signable "$f"; then
      echo "подписываю: $f"
      sign_file "$f" "$OUTDIR" || die "не удалось подписать: $f"
      count=$((count + 1))
    fi
  done < <(find "$path" -type f -print0 2>/dev/null)
done

echo
echo "Подписано файлов: $count"
echo "Отсоединённые подписи: $OUTDIR"
[ "$count" -gt 0 ] || die "ничего не подписано (файлы не найдены или не подходят под критерий ELF/script)"
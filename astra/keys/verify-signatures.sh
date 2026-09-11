#!/usr/bin/env bash
#
# Проверяет, что для всех ELF/скриптов из образов (по манифестам
# keys/sign-images.sh) в $SIG_HOST_DIR лежат отсоединённые подписи с
# именами по путям ВНУТРИ контейнера — ровно так их ищет ядро digsig_verif.
# Выводит список проблем и завершается с кодом 1, если они есть.
#
# Использование: astra/keys/verify-signatures.sh

set -uo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

MANIFESTS="$APP_ROOT/signatures"
[ -d "$MANIFESTS" ] || die "манифесты не найдены ($MANIFESTS) — сначала keys/sign-images.sh"
sudo test -d "$SIG_HOST_DIR" 2>/dev/null || die "каталог подписей не найден: $SIG_HOST_DIR (запустите keys/sign-images.sh)"

rc=0
total=0
missing=0
for manifest in "$MANIFESTS"/*.txt; do
  [ -e "$manifest" ] || continue
  svc="$(basename "$manifest" .txt)"
  svc_missing=0
  svc_total=0
  while IFS= read -r rel; do
    [ -n "$rel" ] || continue
    svc_total=$((svc_total + 1))
    sig_name="@${rel//\//@}.sign"
    if sudo test -f "$SIG_HOST_DIR/$sig_name"; then
      :
    else
      echo "  НЕТ ПОДПИСИ: $svc/$rel  (ожидался файл $SIG_HOST_DIR/$sig_name)"
      svc_missing=$((svc_missing + 1))
    fi
  done < "$manifest"
  total=$((total + svc_total))
  missing=$((missing + svc_missing))
  echo "[$svc] файлов: $svc_total, без подписи: $svc_missing"
done

echo
if [ "$missing" -gt 0 ]; then
  echo "Нет подписей для $missing из $total файлов. Переподпишите: astra/keys/sign-images.sh"
  exit 1
fi
echo "OK: подписи на месте для $total файлов (пути — внутри контейнера)."
echo "Содержимое подписей гарантировано на этапе подписания (подписываются те же байты файлов образа)."
#!/usr/bin/env bash
#
# Скачивает .deb-пакеты (и все зависимости) для оффлайн-развёртывания на
# Astra Linux SE 1.8. Запускать НА МАШИНЕ С INTERNET (лучше — на самой Astra
# до отключения сети: тогда версии пакетов будут точно из её репозитория).
#
# Использование:
#   astra/offline/export-apt-deps.sh [каталог_вывода]     (по умолчанию astra/offline/out)
#
# Результат: <вывод>/apt-deps.tar.gz  → перенести на целевую Astra и распаковать.

set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT="${1:-$SELF_DIR/out}"
DEBS="$OUT/debs"

command -v apt-get >/dev/null 2>&1 || { echo "Нужен Debian-совместимый apt (Astra/Ubuntu/Debian)." >&2; exit 1; }

mkdir -p "$DEBS"

# Пакеты из packages.list, без комментариев.
mapfile -t PACKAGES < <(sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$SELF_DIR/packages.list")

echo "==> Обновление индексов apt …"
sudo apt-get update

echo "==> Скачивание пакетов и зависимостей (${#PACKAGES[@]} шт.): ${PACKAGES[*]}"
sudo apt-get install -y --download-only --no-install-recommends "${PACKAGES[@]}"

echo "==> Копирование .deb в $DEBS …"
sudo cp /var/cache/apt/archives/*.deb "$DEBS"/ 2>/dev/null || true

count="$(find "$DEBS" -maxdepth 1 -name '*.deb' | wc -l)"
[ "$count" -gt 0 ] || { echo "Ни одного .deb не скачано." >&2; exit 1; }

echo "==> Упаковка …"
tar -C "$OUT" -czf "$OUT/apt-deps.tar.gz" debs

echo
echo "Готово: $OUT/apt-deps.tar.gz ($count .deb)"
echo "Перенесите его на целевую Astra и выполните:"
echo "  tar -xzf apt-deps.tar.gz"
echo "  astra/setup/install-rootless-docker.sh debs"
#!/usr/bin/env bash
#
# Регистрирует публичный ключ подписи ЗПС в системе, чтобы ядро (модуль
# digsig_verif) могло проверять подписи файлов.
#
# Что делает:
#   - кладёт публичный ключ в /etc/digsig/xattr_keys (для xattr-подписей,
#     сделанных самогенерированным ключом);
#   - кладёт его же как pubkey.gpg в /etc/digsig/keys/legacy/keys;
#   - пересобирает initramfs (ключ «загружается в ядро» при загрузке).
#
# После этого — включить ЗПС (setup/enable-zps.sh) и ПЕРЕЗАГРУЗИТЬ ОС.
#
# Использование: astra/setup/register-keys.sh

set -euo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

PUB="$KEYS_DIR/taygant-pub.gpg"
[ -f "$PUB" ] || die "нет публичного ключа: $PUB — сначала запустите keys/gen-keys.sh"

echo "Регистрация ключа $PUB …"
sudo mkdir -p /etc/digsig/xattr_keys /etc/digsig/keys/legacy/keys
sudo install -m 644 -o root -g root "$PUB" /etc/digsig/xattr_keys/taygant-pub.gpg
sudo install -m 644 -o root -g root "$PUB" /etc/digsig/keys/legacy/keys/pubkey.gpg

echo "Пересборка initramfs (ключ попадёт в ядро при загрузке) …"
sudo update-initramfs -uk all

echo
echo "Готово. Ключ зарегистрирован:"
echo "  /etc/digsig/xattr_keys/taygant-pub.gpg"
echo "  /etc/digsig/keys/legacy/keys/pubkey.gpg"
echo
echo "Дальше: astra/setup/enable-zps.sh → перезагрузка ОС."
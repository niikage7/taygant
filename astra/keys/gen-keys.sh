#!/usr/bin/env bash
#
# Генерирует пару ключей подписи ГОСТ Р 34.10-2012 для ЗПС и экспортирует
# публичный ключ и сертификат отзыва.
#
# ВАЖНО про самогенерированные ключи (по документации Астра):
#   - встраиваемые ELF-подписи такими ключами НЕ проверяются (нужны ключи,
#     подписанные изготовителем) — поэтому скрипты подписывают в расширенных
#     атрибутах (xattr) и отсоединённой подписью;
#   - секретный ключ не должен покидать эту машину.
#
# По умолчанию ключ создаётся БЕЗ пароля (удобно для демо). Чтобы задать
# пароль, положите его в SIGN_PASSPHRASE в .env.astra (или экспортируйте).
#
# Использование: astra/keys/gen-keys.sh
# Результат (в $KEYS_DIR, по умолчанию ~/taygant/keys):
#   KEYID              — идентификатор секретного ключа
#   taygant-pub.gpg    — публичный ключ (регистрируется в ядре через setup/register-keys.sh)
#   taygant-revoke.rev — сертификат отзыва (хранить в надёжном месте)

set -euo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

ORG_NAME="${TAYGANT_ORG_NAME:-TAYGANT}"

command -v gpg >/dev/null 2>&1 || die "gpg не установлен — см. offline/packages.list (пакет gnupg)"

mkdir -p "$KEYS_DIR"

if [ -f "$KEYS_DIR/KEYID" ]; then
  existing="$(cat "$KEYS_DIR/KEYID")"
  echo "Ключ уже существует: $existing"
  echo "  секретный:  в связке gpg пользователя $(id -un)"
  echo "  публичный:  $KEYS_DIR/taygant-pub.gpg"
  echo "Для генерации нового ключа удалите $KEYS_DIR/KEYID (и при желании — секретный ключ: gpg --delete-secret-key)."
  exit 0
fi

PARAMS="$KEYS_DIR/gen-params.gpg"
cat > "$PARAMS" <<EOF
Key-Type:GOST_R34.10-2012
Key-Usage: sign
Name-Real: ${ORG_NAME}
Name-Comment: taygant digsig signing key (ZPS)
Name-Email: taygant@localhost
Expire-Date: 0
EOF
if [ -n "${SIGN_PASSPHRASE:-}" ]; then
  printf 'Passphrase: %s\n' "$SIGN_PASSPHRASE" >> "$PARAMS"
else
  printf '%%no-protection\n' >> "$PARAMS"
fi
printf '%%commit\n' >> "$PARAMS"

echo "==> Генерация ключа ГОСТ Р 34.10-2012 (${ORG_NAME}) …"
gpg --batch --gen-key "$PARAMS"

FPR="$(gpg --list-secret-keys --with-colons | awk -F: '$1=="fpr" {print $10; exit}')"
[ -n "$FPR" ] || die "не удалось получить отпечаток ключа"

printf '%s' "$FPR" > "$KEYS_DIR/KEYID"

echo "==> Экспорт публичного ключа …"
# Бинарный формат (keyring) — именно его ожидает digsig_verif в /etc/digsig/keys.
gpg --export "$FPR" > "$KEYS_DIR/taygant-pub.gpg"

echo "==> Сертификат отзыва …"
# Причина 0 = «без указания причины», дальше пустой ввод — подтверждение.
if ! gpg --output "$KEYS_DIR/taygant-revoke.rev" --gen-revoke "$FPR" \
    --batch --command-fd 0 >/dev/null 2>&1 <<EOF
y
0

EOF
then
  echo "  (сертификат отзыва не создан — при желании позже: gpg --gen-revoke $FPR)"
fi

rm -f "$PARAMS"
chmod 700 "$KEYS_DIR"
[ -f "$KEYS_DIR/taygant-pub.gpg" ] && chmod 644 "$KEYS_DIR/taygant-pub.gpg"

echo
echo "Готово. Ключ подписи: $FPR"
echo "  публичный:  $KEYS_DIR/taygant-pub.gpg"
echo "  отзыв:      $KEYS_DIR/taygant-revoke.rev (храните в надёжном месте)"
echo "  секретный:  в связке gpg пользователя $(id -un) — никуда не копируйте"
echo
echo "Дальнейшие шаги:"
echo "  1) astra/keys/sign-images.sh   — подписать файлы в образах"
echo "  2) astra/setup/register-keys.sh — зарегистрировать публичный ключ в ядре"
echo "  3) astra/setup/enable-zps.sh   — включить ЗПС и перезагрузиться"
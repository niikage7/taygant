#!/usr/bin/env bash
#
# Оффлайн-установка Docker в rootless-режиме на Astra Linux SE 1.8 (Воронеж).
#
# Входные данные: каталог с .deb-пакетами (результат offline/export-apt-deps.sh).
#   tar -xzf apt-deps.tar.gz   →  каталог debs/
#   astra/setup/install-rootless-docker.sh debs
#
# Что делает:
#   1) устанавливает пакеты (docker-ce, docker-ce-rootless-extras, bsign и др.)
#      из локальных .deb без доступа в интернет;
#   2) назначает текущему пользователю диапазоны subordinate uid/gid
#      (нужны rootless-демону);
#   3) включает unprivileged user namespaces, если это разрешено конфигом ядра;
#   4) устанавливает и запускает systemd-юнит rootless-демона Docker
#      (dockerd-rootless-setuptool.sh install);
#   5) включает lingering — демон стартует при загрузке без входа пользователя.
#
# ВАЖНО: пользователь, из-под которого выполняется скрипт, и будет «владельцем»
# контейнеров. Запускайте от обычного пользователя (не root), sudo используется
# только для привилегированных шагов.

set -euo pipefail

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../lib/common.sh
. "$SELF_DIR/../lib/common.sh"

DEBS_DIR="${1:-}"
[ -n "$DEBS_DIR" ] || die "укажите каталог с .deb: ./setup/install-rootless-docker.sh debs"
[ -d "$DEBS_DIR" ] || die "каталог не найден: $DEBS_DIR"
debs=("$DEBS_DIR"/*.deb)
[ -e "${debs[0]}" ] || die "в $DEBS_DIR нет .deb-пакетов"

[ "$(id -u)" -ne 0 ] || die "запускайте от обычного пользователя (rootless Docker не работает из-под root)"
USER_RUN="$(id -un)"

echo "===== 1/5. Установка пакетов из $DEBS_DIR ====="
# Кладём .deb в apt-кэш, чтобы apt-get -f нашёл зависимости оффлайн.
sudo cp "$DEBS_DIR"/*.deb /var/cache/apt/archives/ 2>/dev/null || true
# Устанавливаем всё, что можно, и доустанавливаем зависимости из кэша.
sudo dpkg -i "$DEBS_DIR"/*.deb >/dev/null 2>&1 || true
sudo apt-get install -f -y --no-install-recommends -o Acquire::Retries=0
sudo dpkg --configure -a >/dev/null 2>&1 || true

echo "===== 2/5. Диапазоны subordinate uid/gid для $USER_RUN ====="
if ! grep -q "^${USER_RUN}:" /etc/subuid 2>/dev/null; then
  sudo usermod --add-subuids 100000-165535 "$USER_RUN" || die "не удалось добавить subuids"
fi
if ! grep -q "^${USER_RUN}:" /etc/subgid 2>/dev/null; then
  sudo usermod --add-subgids 100000-165535 "$USER_RUN" || die "не удалось добавить subgids"
fi
grep "^${USER_RUN}:" /etc/subuid /etc/subgid

echo "===== 3/5. User namespaces (sysctl) ====="
if [ "$(sysctl -n kernel.unprivileged_userns_clone 2>/dev/null)" = "0" ]; then
  echo 'kernel.unprivileged_userns_clone=1' | sudo tee /etc/sysctl.d/99-rootless-docker.conf >/dev/null
  sudo sysctl --system >/dev/null
  echo "  включено kernel.unprivileged_userns_clone=1"
elif [ -z "$(sysctl -n kernel.unprivileged_userns_clone 2>/dev/null)" ]; then
  echo "  параметр kernel.unprivileged_userns_clone отсутствует (управляется конфигом ядра) — продолжаем"
else
  echo "  kernel.unprivileged_userns_clone уже включён"
fi

echo "===== 4/5. Установка rootless-демона Docker ====="
SETUPTOOL="$(command -v dockerd-rootless-setuptool.sh || true)"
if [ -z "$SETUPTOOL" ]; then
  # Пакет docker-ce-rootless-extras ставит его в /usr/local/bin или /usr/bin.
  for p in /usr/local/bin/dockerd-rootless-setuptool.sh /usr/bin/dockerd-rootless-setuptool.sh; do
    [ -x "$p" ] && { SETUPTOOL="$p"; break; }
  done
fi
[ -n "$SETUPTOOL" ] || die "dockerd-rootless-setuptool.sh не найден — проверьте установку docker-ce-rootless-extras"
echo "  Утилита: $SETUPTOOL"
"$SETUPTOOL" install

echo "===== 5/5. Lingering (автостарт при загрузке) ====="
loginctl enable-linger "$USER_RUN" 2>/dev/null || true

echo
echo "===== Проверка ====="
export PATH="$HOME/.local/bin:$PATH"
if docker info >/dev/null 2>&1; then
  echo "OK: rootless Docker работает:"
  docker info --format '{{.ServerVersion}} / {{.OperatingSystem}}' 2>/dev/null || docker version
else
  echo "Клиент docker найден, но демон не отвечает. Проверьте:"
  echo "  systemctl --user status docker"
  echo "  journalctl --user -u docker -n 50"
  echo "Если юнит не запустился — выйдите и зайдите заново (или: loginctl enable-linger, systemctl --user start docker)."
fi
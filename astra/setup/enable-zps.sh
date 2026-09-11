#!/usr/bin/env bash
#
# Включает режим замкнутой программной среды (ЗПС) на Astra Linux SE 1.8.
# Для x.8 используется команда astra-digsig-control; для более старых
# обновлений — astra-zps-control (запасной вариант).
#
# ПЕРЕД запуском:
#   1) astra/keys/gen-keys.sh        — ключ подписи
#   2) astra/keys/sign-images.sh     — подписать файлы сервисов
#   3) astra/setup/register-keys.sh  — публичный ключ в ядро (initramfs)
#
# ПОСЛЕ включения обязательна перезагрузка ОС.
#
# Отключение ЗПС:  sudo astra-digsig-control disable  (или astra-zps-control disable)
#
# Использование: astra/setup/enable-zps.sh

set -euo pipefail

if command -v astra-digsig-control >/dev/null 2>&1; then
  echo "Текущий статус ЗПС:"
  sudo astra-digsig-control status 2>/dev/null || true
  echo
  echo "Включение ЗПС …"
  sudo astra-digsig-control enable
elif command -v astra-zps-control >/dev/null 2>&1; then
  echo "Текущий статус ЗПС:"
  sudo astra-zps-control status 2>/dev/null || true
  echo
  echo "Включение ЗПС …"
  sudo astra-zps-control enable
else
  echo "Утилита управления ЗПС не найдена (astra-digsig-control / astra-zps-control)." >&2
  echo "Проверьте: dpkg -l | grep -i digsig" >&2
  exit 1
fi

echo
echo "ЗПС включена. ПЕРЕЗАГРУЗИТЕ ОС:  sudo reboot"
echo "После перезагрузки: astra/run/up.sh и проверка, что контейнеры поднялись под ЗПС."
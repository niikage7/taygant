#!/usr/bin/env bash
#
# Смоук-тест endpoints стека. Удобно вызывать из status.sh и вручную.
#
# Использование: astra/run/healthcheck.sh

set -uo pipefail

# shellcheck source=../lib/common.sh
. "$(cd "$(dirname "${BASH_SOURCE[0]}")/../lib" && pwd)/common.sh"

ok=0; bad=0
probe() { # name url
  if http_ok "$2"; then
    echo "  OK   $1  →  $2"
    ok=$((ok + 1))
  else
    echo "  FAIL $1  →  $2"
    bad=$((bad + 1))
  fi
}

if (exec 3<>/dev/tcp/127.0.0.1/5432) 2>/dev/null; then
  echo "  OK   postgres  →  tcp://127.0.0.1:5432"
  ok=$((ok + 1))
else
  echo "  FAIL postgres  →  tcp://127.0.0.1:5432"
  bad=$((bad + 1))
fi
probe backend  "http://127.0.0.1:4000/ping"
probe frontend "http://127.0.0.1:3000/login"

echo
[ "$bad" -eq 0 ] || exit 1
echo "Все проверки пройдены."
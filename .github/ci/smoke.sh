#!/usr/bin/env bash
# Смоук-тест запущенного стека: проходит путь браузера
#   фронтенд → прокси /backend → Go API → PostgreSQL.
#
# Проверки выполняются изнутри контейнера frontend, поэтому не зависят от
# опубликованных портов (в CI их нет вовсе). Стек выбирается переменными
# docker compose из окружения: COMPOSE_PROJECT_NAME, COMPOSE_FILE, JWT_SECRET.
#
# SMOKE_DEMO_LOGIN=true — дополнительно войти демо-пользователем. Только для CI:
# на сервере пароль демо-учёток могли сменить, и смоук-тест не должен из-за
# этого откатывать рабочий деплой.
#
# Анализатор не видит вызовов через check "$@" и считает функции неиспользуемыми.
# shellcheck disable=SC2329
set -uo pipefail

failed=0
rows=()

in_frontend() { docker compose exec -T frontend "$@"; }

# check "описание" команда... — выполняет проверку и запоминает результат.
check() {
  local title=$1
  shift
  local output
  if output=$("$@" 2>&1); then
    echo "✅ $title"
    rows+=("| ✅ | $title |")
  else
    echo "❌ $title"
    while IFS= read -r line; do echo "     $line"; done <<<"$output"
    echo "::error title=Смоук-тест не прошёл::$title"
    rows+=("| ❌ | $title |")
    failed=1
  fi
}

page() {
  in_frontend wget -q -O /dev/null "http://127.0.0.1:3000$1"
}

# body_contains ожидаемое url [аргументы wget...]
body_contains() {
  local want=$1
  shift
  local body
  body=$(in_frontend wget -q -O - "$@") || return 1
  grep -q -- "$want" <<<"$body" || { echo "ответ: ${body:0:300}"; return 1; }
}

# status_is код url — busybox wget сообщает код ответа в тексте ошибки.
status_is() {
  local out
  out=$(in_frontend wget -q -O /dev/null "$2" 2>&1)
  grep -q "$1" <<<"$out" || { echo "ожидался HTTP $1, получено: ${out:-200 OK}"; return 1; }
}

all_healthy() {
  local states
  states=$(docker compose ps -a --format '{{.Service}}: {{.State}} {{.Health}}')
  echo "$states"
  ! grep -qv 'running healthy' <<<"$states"
}

echo "🩺 Смоук-тест стека «${COMPOSE_PROJECT_NAME:-по умолчанию}»"
check "Все контейнеры запущены и здоровы" all_healthy
check "Страница входа /login открывается" page /login
check "Пользовательское соглашение /terms открывается" page /terms
check "Фронтенд проксирует запросы в бэкенд, бэкенд видит базу (/backend/ping → pong)" \
  body_contains pong http://127.0.0.1:3000/backend/ping
check "API закрыт без токена (/api/v1/projects → 401)" \
  status_is 401 http://127.0.0.1:3000/backend/api/v1/projects
check "MCP-сервер для ассистентов доступен через прокси и закрыт без личного ключа (/mcp → 401)" \
  status_is 401 http://127.0.0.1:3000/backend/mcp
if [ "${SMOKE_DEMO_LOGIN:-false}" = "true" ]; then
  check "Вход демо-пользователем через API (сидинг, bcrypt, JWT)" \
    body_contains accessToken \
    --header 'Content-Type: application/json' \
    --post-data '{"email":"volkov@taygant.dev","password":"demo1234"}' \
    http://127.0.0.1:3000/backend/api/v1/auth/login
fi

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
  {
    echo "### 🩺 Смоук-тест"
    echo
    echo "| | Проверка |"
    echo "|:-:|---|"
    printf '%s\n' "${rows[@]}"
    echo
  } >>"$GITHUB_STEP_SUMMARY"
fi

exit "$failed"

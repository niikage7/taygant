# syntax=docker/dockerfile:1
#
# Образ фронтенда: Next.js в режиме standalone (см. output в next.config.ts).
#
# BACKEND_URL передаётся при СБОРКЕ, а не при запуске: rewrites из next.config.ts
# вычисляются во время `next build` и вшиваются в сборку. В docker compose это
# адрес сервиса бэкенда внутри сети — http://backend:4000.

# --- База: Node + pnpm нужной версии (берётся из packageManager в package.json) ---
FROM node:22-alpine AS base
ENV NEXT_TELEMETRY_DISABLED=1 \
    COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN corepack enable
WORKDIR /app

# --- Зависимости отдельным слоем: пересобираются только при изменении lock-файла ---
FROM base AS deps
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm install --frozen-lockfile --store-dir /pnpm/store

# --- Сборка ---
FROM base AS build
ARG BACKEND_URL=http://backend:4000
ENV BACKEND_URL=${BACKEND_URL}
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build

# --- Запуск: только результат сборки, без исходников и dev-зависимостей ---
FROM node:22-alpine AS run
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
WORKDIR /app

# Пользователь node (uid 1000) уже есть в официальном образе — root не нужен.
COPY --from=build --chown=node:node /app/public ./public
COPY --from=build --chown=node:node /app/.next/standalone ./
COPY --from=build --chown=node:node /app/.next/static ./.next/static

USER node
EXPOSE 3000

CMD ["node", "server.js"]

# --- Сборка ---
FROM golang:1.26-alpine AS build

WORKDIR /src

# Зависимости отдельным слоем: пересобираются только при изменении go.mod/go.sum,
# а не при каждой правке кода.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO не нужен — получаем статический бинарник, который запускается в голом alpine.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/taygant-api .

# --- Запуск ---
FROM alpine:3.21

# Сертификаты и таймзоны: без tzdata time.LoadLocation внутри контейнера не работает.
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app

COPY --from=build /out/taygant-api /usr/local/bin/taygant-api

USER app
EXPOSE 4000

ENTRYPOINT ["/usr/local/bin/taygant-api"]

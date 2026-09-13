// Package config собирает настройки приложения из переменных окружения.
// Значения по умолчанию рассчитаны на локальную разработку (фронт на localhost:3000),
// в docker compose они переопределяются через environment сервиса backend.
package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config — настройки, влияющие на поведение HTTP-слоя.
type Config struct {
	// Port — порт, который слушает приложение (APP_PORT).
	Port string
	// AllowedOrigins — origin'ы, которым браузер разрешит читать ответы API (ALLOWED_ORIGINS).
	AllowedOrigins []string
	// TrustedProxies — адреса/подсети прокси, чьему X-Forwarded-For можно верить (TRUSTED_PROXIES).
	// Пустой список означает, что заголовок игнорируется и IP берётся из соединения.
	TrustedProxies []string
	// RateLimitMax — максимум запросов к /api/v1 с одного IP за окно (RATE_LIMIT_MAX).
	RateLimitMax int
	// RateLimitWindow — длительность окна лимитера (RATE_LIMIT_WINDOW).
	RateLimitWindow time.Duration
	// DatabaseURL — строка подключения к PostgreSQL (DATABASE_URL).
	DatabaseURL string
	// JWTSecret — ключ подписи access- и refresh-токенов (JWT_SECRET).
	// Смена ключа инвалидирует все выданные токены.
	JWTSecret string
	// AccessTokenTTL — время жизни access-токена (ACCESS_TOKEN_TTL).
	AccessTokenTTL time.Duration
	// RefreshTokenTTL — время жизни refresh-токена (REFRESH_TOKEN_TTL).
	RefreshTokenTTL time.Duration
	// SeedDemoData — наполнять ли пустую БД демонстрационными данными (SEED_DEMO_DATA).
	SeedDemoData bool
}

// Load читает конфигурацию из окружения, подставляя значения по умолчанию.
// JWT_SECRET обязателен: без него приложение не стартует, чтобы токены
// никогда не подписывались общеизвестным ключом из репозитория.
func Load() Config {
	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET не задан: сгенерируйте свой ключ (openssl rand -hex 32) и укажите его в .env")
	}

	return Config{
		Port:            envString("APP_PORT", "4000"),
		AllowedOrigins:  envOrigins("ALLOWED_ORIGINS", "http://localhost:3000"),
		TrustedProxies:  envList("TRUSTED_PROXIES", "127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"),
		RateLimitMax:    envInt("RATE_LIMIT_MAX", 300),
		RateLimitWindow: envDuration("RATE_LIMIT_WINDOW", time.Minute),
		DatabaseURL:     envString("DATABASE_URL", "postgres://taygant:taygant@localhost:5432/taygant?sslmode=disable"),
		JWTSecret:       jwtSecret,
		AccessTokenTTL:  envDuration("ACCESS_TOKEN_TTL", 24*time.Hour),
		RefreshTokenTTL: envDuration("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		SeedDemoData:    envBool("SEED_DEMO_DATA", true),
	}
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// envList разбирает список значений через запятую, отбрасывая пробелы и пустые элементы.
// В отличие от envString, явно заданная пустая переменная не подменяется значением
// по умолчанию: TRUSTED_PROXIES="" означает "не доверять никаким прокси".
func envList(key, fallback string) []string {
	raw, ok := os.LookupEnv(key)
	if !ok {
		raw = fallback
	}
	return parseList(raw)
}

func parseList(raw string) []string {
	var items []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// envOrigins дополнительно снимает завершающий слэш: браузер присылает Origin без него,
// и "http://localhost:3000/" в списке никогда бы не совпал.
//
// Список гарантированно непустой: Fiber трактует пустой AllowOrigins как "*", то есть
// опечатка в ALLOWED_ORIGINS иначе молча открыла бы API всем origin'ам.
func envOrigins(key, fallback string) []string {
	origins := envList(key, fallback)
	if len(origins) == 0 {
		origins = parseList(fallback)
	}
	for i, origin := range origins {
		origins[i] = strings.TrimRight(origin, "/")
	}
	return origins
}

func envInt(key string, fallback int) int {
	if v, err := strconv.Atoi(envString(key, "")); err == nil && v > 0 {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	if v, err := strconv.ParseBool(envString(key, "")); err == nil {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, err := time.ParseDuration(envString(key, "")); err == nil && v > 0 {
		return v
	}
	return fallback
}

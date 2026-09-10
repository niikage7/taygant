// Package server собирает Fiber-приложение: глобальные middleware, обработку ошибок
// и монтирование маршрутов API.
package server

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"taygant_backend/internal/api"
	"taygant_backend/internal/config"
)

// New создаёт настроенное приложение, готовое к Listen.
func New(cfg config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "taygant-api",
		ErrorHandler: errorHandler,

		// За обратным прокси (nginx / Next.js в docker compose) реальный IP клиента
		// приходит в X-Forwarded-For. Fiber доверяет этому заголовку только если запрос
		// пришёл от адреса из TrustedProxies — иначе IP берётся из соединения, и заголовок
		// нельзя подделать снаружи. От этого зависит корректность rate limit.
		ProxyHeader:             fiber.HeaderXForwardedFor,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          cfg.TrustedProxies,
		EnableIPValidation:      true,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${ip} ${status} ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(corsConfig(cfg)))

	// Health-check для docker compose — намеренно вне /api/v1 и без rate limit.
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	api.Register(app, rateLimiter(cfg))

	return app
}

// corsConfig разрешает браузеру обращаться к API только со страниц из белого списка origin'ов.
// CORS — ограничение на стороне браузера: он не мешает curl или боту достучаться до API,
// поэтому реальный доступ должна резать авторизация, а не эти настройки.
func corsConfig(cfg config.Config) cors.Config {
	return cors.Config{
		AllowOrigins: strings.Join(cfg.AllowedOrigins, ","),
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		}, ","),
		AllowHeaders: strings.Join([]string{
			fiber.HeaderOrigin,
			fiber.HeaderContentType,
			fiber.HeaderAccept,
			fiber.HeaderAuthorization,
		}, ","),
		// Нужен фронту, чтобы вытащить имя файла при скачивании /projects/{id}/export.
		ExposeHeaders: fiber.HeaderContentDisposition,
		// Авторизация построена на Bearer-токене в заголовке, cookies не используются.
		AllowCredentials: false,
		MaxAge:           int((12 * time.Hour).Seconds()),
	}
}

// rateLimiter ограничивает частоту запросов к API с одного IP.
func rateLimiter(cfg config.Config) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.RateLimitMax,
		Expiration: cfg.RateLimitWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		// Preflight-запросы браузера не являются обращениями к API и квоту не тратят.
		Next: func(c *fiber.Ctx) bool {
			return c.Method() == fiber.MethodOptions
		},
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many requests")
		},
	})
}

// errorHandler приводит любые ошибки к единому JSON-формату, чтобы фронт мог
// разбирать их одинаково независимо от источника ошибки.
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "internal server error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"path":  c.Path(),
	})
}

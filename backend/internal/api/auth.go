package api

import "github.com/gofiber/fiber/v2"

// registerAuthRoutes — тег Auth: аутентификация и выдача токенов.
func registerAuthRoutes(r fiber.Router) {
	auth := r.Group("/auth")

	auth.Post("/login", authLogin)
	auth.Post("/refresh", authRefresh)
}

// POST /auth/login — вход пользователя.
// Проверяет логин/пароль и выдаёт пару access/refresh токенов.
// Body: { email, password }. 200 -> { accessToken, refreshToken, user }, 401 -> неверные учётные данные.
func authLogin(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /auth/refresh — обновление access-токена.
// Body: { refreshToken }. 200 -> { accessToken }.
func authRefresh(c *fiber.Ctx) error {
	return notImplemented(c)
}

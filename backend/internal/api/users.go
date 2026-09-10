package api

import "github.com/gofiber/fiber/v2"

// registerUsersRoutes — тег Users: пользователи платформы.
func registerUsersRoutes(r fiber.Router) {
	users := r.Group("/users")

	users.Get("/me", usersMe)
	users.Get("/", usersList)
}

// GET /users/me — текущий пользователь.
// Возвращает профиль пользователя, соответствующего переданному access-токену. 200 -> User.
func usersMe(c *fiber.Ctx) error {
	return notImplemented(c)
}

// GET /users — список пользователей (для назначения ответственных).
// Query: search — подстрока для поиска по имени или email. 200 -> []User.
func usersList(c *fiber.Ctx) error {
	return notImplemented(c)
}

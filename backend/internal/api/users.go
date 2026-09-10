package api

import (
	"github.com/gofiber/fiber/v2"
)

// registerUsersRoutes — тег Users: пользователи платформы.
func (a *API) registerUsersRoutes(r fiber.Router) {
	users := r.Group("/users")

	users.Get("/me", a.usersMe)
	users.Get("/", a.usersList)
}

// GET /users/me — текущий пользователь.
// Возвращает профиль пользователя, соответствующего переданному access-токену. 200 -> User.
//
// Запроса к БД здесь нет: requireAuth уже загрузил пользователя, чтобы
// убедиться, что учётная запись жива, и положил его в контекст.
func (a *API) usersMe(c *fiber.Ctx) error {
	return c.JSON(newUserDTO(currentUser(c)))
}

// GET /users — список пользователей (для назначения ответственных).
// Query: search — подстрока для поиска по имени или email. 200 -> []User.
func (a *API) usersList(c *fiber.Ctx) error {
	users, err := a.users.List(c.Context(), c.Query("search"))
	if err != nil {
		return fail(err)
	}

	out := make([]userDTO, 0, len(users))
	for _, u := range users {
		out = append(out, newUserDTO(u))
	}
	return c.JSON(out)
}

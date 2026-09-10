package api

import (
	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/service"
)

// registerAuthRoutes — тег Auth: аутентификация и выдача токенов.
// Единственная группа без requireAuth: получить токен, предъявив токен, невозможно.
func (a *API) registerAuthRoutes(r fiber.Router) {
	auth := r.Group("/auth")

	auth.Post("/register", a.authRegister)
	auth.Post("/login", a.authLogin)
	auth.Post("/refresh", a.authRefresh)
}

// registerRequest — тело POST /auth/register.
type registerRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"fullName"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

// POST /auth/register — регистрация пользователя.
//
// Эндпоинта нет в api-spec.yml: спецификация описывает только вход, предполагая
// заранее заведённые учётные записи. Расширение согласовано с командой —
// без него пользователей можно было бы создавать только сидингом.
// Body: { email, password, fullName, department, position }.
// 201 -> { accessToken, refreshToken, user }, 400 -> некорректные данные, 409 -> email занят.
func (a *API) authRegister(c *fiber.Ctx) error {
	body, err := parseBody[registerRequest](c)
	if err != nil {
		return err
	}

	user, tokens, err := a.auth.Register(c.Context(), toRegisterInput(body))
	if err != nil {
		return fail(err)
	}

	return c.Status(fiber.StatusCreated).JSON(authTokensDTO{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
		User:         newUserDTO(user),
	})
}

// toRegisterInput переносит тело запроса во входные данные сервиса.
func toRegisterInput(body registerRequest) service.RegisterInput {
	return service.RegisterInput{
		Email:      body.Email,
		Password:   body.Password,
		FullName:   body.FullName,
		Department: body.Department,
		Position:   body.Position,
	}
}

// loginRequest — тело POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /auth/login — вход пользователя.
// Проверяет логин/пароль и выдаёт пару access/refresh токенов.
// Body: { email, password }. 200 -> { accessToken, refreshToken, user }, 401 -> неверные учётные данные.
func (a *API) authLogin(c *fiber.Ctx) error {
	body, err := parseBody[loginRequest](c)
	if err != nil {
		return err
	}

	user, tokens, err := a.auth.Login(c.Context(), body.Email, body.Password)
	if err != nil {
		return fail(err)
	}

	return c.JSON(authTokensDTO{
		AccessToken:  tokens.Access,
		RefreshToken: tokens.Refresh,
		User:         newUserDTO(user),
	})
}

// refreshRequest — тело POST /auth/refresh.
type refreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// POST /auth/refresh — обновление access-токена.
// Body: { refreshToken }. 200 -> { accessToken }, 401 -> токен недействителен.
func (a *API) authRefresh(c *fiber.Ctx) error {
	body, err := parseBody[refreshRequest](c)
	if err != nil {
		return err
	}

	access, err := a.auth.Refresh(c.Context(), body.RefreshToken)
	if err != nil {
		return fail(err)
	}

	return c.JSON(accessTokenDTO{AccessToken: access})
}

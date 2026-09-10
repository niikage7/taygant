package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/auth"
	"taygant_backend/internal/models"
)

// Ключи, под которыми middleware кладёт данные запроса в контекст Fiber.
const (
	localsUserID = "userID"
	localsUser   = "user"
)

// requireAuth пропускает дальше только запросы с действительным access-токеном.
//
// Пользователь подгружается из БД, а не берётся из токена: токен живёт сутки,
// и за это время учётная запись может быть удалена или переименована.
// Это лишний запрос на каждый вызов, но он же гарантирует, что хендлеры
// работают с актуальными данными.
func (a *API) requireAuth(c *fiber.Ctx) error {
	token, err := bearerToken(c)
	if err != nil {
		return err
	}

	userID, err := a.tokens.Parse(token, auth.KindAccess)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "недействительный или истёкший токен")
	}

	user, err := a.auth.UserByID(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "пользователь не найден")
	}

	c.Locals(localsUserID, user.ID)
	c.Locals(localsUser, user)
	return c.Next()
}

// bearerToken достаёт токен из заголовка Authorization.
func bearerToken(c *fiber.Ctx) (string, error) {
	const prefix = "Bearer "

	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "требуется авторизация")
	}
	// Регистр схемы по RFC 7235 значения не имеет, поэтому сравниваем без учёта регистра.
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", fiber.NewError(fiber.StatusUnauthorized, "ожидается заголовок Authorization: Bearer <token>")
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", fiber.NewError(fiber.StatusUnauthorized, "пустой токен")
	}
	return token, nil
}

// currentUserID возвращает идентификатор пользователя, установленный requireAuth.
// Вызов из маршрута без requireAuth — ошибка программиста, поэтому вернётся uuid.Nil.
func currentUserID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(localsUserID).(uuid.UUID)
	return id
}

// currentUser возвращает пользователя, установленного requireAuth.
func currentUser(c *fiber.Ctx) models.User {
	user, _ := c.Locals(localsUser).(models.User)
	return user
}

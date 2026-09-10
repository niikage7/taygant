package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/access"
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

// requireMember проверяет, что пользователь состоит в команде проекта, и
// возвращает загруженные проект и запись участника — их почти всегда нужно
// использовать дальше (например, для расчёта сводных показателей).
func (a *API) requireMember(c *fiber.Ctx, projectID uuid.UUID) (access.Membership, error) {
	m, err := access.Load(c.Context(), a.db, projectID, currentUserID(c))
	if err != nil {
		return access.Membership{}, fail(err)
	}
	return m, nil
}

// requireEdit — то же, что requireMember, но дополнительно проверяет право
// редактировать план проекта (задачи, спринты, вехи, чек-лист, связи).
func (a *API) requireEdit(c *fiber.Ctx, projectID uuid.UUID) error {
	m, err := a.requireMember(c, projectID)
	if err != nil {
		return err
	}
	return fail(m.RequireEdit())
}

// requireManage — то же, что requireMember, но дополнительно проверяет право
// управлять проектом и его командой (уровень доступа full).
func (a *API) requireManage(c *fiber.Ctx, projectID uuid.UUID) (access.Membership, error) {
	m, err := a.requireMember(c, projectID)
	if err != nil {
		return access.Membership{}, err
	}
	if err := fail(m.RequireManage()); err != nil {
		return access.Membership{}, err
	}
	return m, nil
}

// requireEditByTask резолвит projectId по задаче и проверяет право её редактировать.
// Используется хендлерами вложенных ресурсов (dependencies, checklist, comments,
// history), которые адресуются taskId без projectId в пути.
func (a *API) requireEditByTask(c *fiber.Ctx, taskID uuid.UUID) (uuid.UUID, error) {
	projectID, err := a.tasks.ProjectID(c.Context(), taskID)
	if err != nil {
		return uuid.Nil, fail(err)
	}
	return projectID, a.requireEdit(c, projectID)
}

// requireMemberByTask — то же, что requireEditByTask, но только для чтения:
// достаточно быть участником проекта, редактировать не обязательно.
func (a *API) requireMemberByTask(c *fiber.Ctx, taskID uuid.UUID) (uuid.UUID, error) {
	projectID, err := a.tasks.ProjectID(c.Context(), taskID)
	if err != nil {
		return uuid.Nil, fail(err)
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return uuid.Nil, err
	}
	return projectID, nil
}

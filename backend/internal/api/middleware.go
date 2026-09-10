package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/access"
	"taygant_backend/internal/auth"
	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
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

// requireManage — то же, что requireMember, но дополнительно проверяет право
// управлять проектом: правку плана (задачи, спринты, вехи, чек-лист, связи),
// параметры проекта и состав команды. Это уровень доступа full.
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

// requireTaskWrite проверяет право изменить конкретную задачу и возвращает
// загруженную задачу вместе с записью участника.
//
// Уровень edit разрешает правку только своих задач (где участник — исполнитель),
// причём лишь статуса и сроков; какие поля пришли в запросе, знает хендлер,
// поэтому полная проверка набора полей остаётся там.
func (a *API) requireTaskWrite(c *fiber.Ctx, taskID uuid.UUID) (access.Membership, models.Task, error) {
	task, err := a.tasks.GetBrief(c.Context(), taskID)
	if err != nil {
		return access.Membership{}, models.Task{}, fail(err)
	}
	m, err := a.requireMember(c, task.ProjectID)
	if err != nil {
		return access.Membership{}, models.Task{}, err
	}
	if !m.CanEditTask(task) {
		return access.Membership{}, models.Task{}, fail(service.ErrForbidden)
	}
	return m, task, nil
}

// requireManageByTask резолвит projectId по задаче и проверяет право управлять
// планом проекта. Используется хендлерами вложенных ресурсов (dependencies,
// checklist, simulation), которые адресуются taskId без projectId в пути.
func (a *API) requireManageByTask(c *fiber.Ctx, taskID uuid.UUID) (uuid.UUID, error) {
	projectID, err := a.tasks.ProjectID(c.Context(), taskID)
	if err != nil {
		return uuid.Nil, fail(err)
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return uuid.Nil, err
	}
	return projectID, nil
}

// requireMemberByTask — то же, что requireManageByTask, но только для чтения:
// достаточно быть участником проекта, управлять не обязательно.
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

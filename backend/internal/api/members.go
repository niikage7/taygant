package api

import "github.com/gofiber/fiber/v2"

// registerMembersRoutes — тег Members: состав команды проекта и роли участников.
func (a *API) registerMembersRoutes(r fiber.Router) {
	members := r.Group("/projects/:projectId/members")

	members.Get("/", a.membersList)
	members.Post("/", a.membersCreate)
	members.Patch("/:memberId", a.membersUpdate)
	members.Delete("/:memberId", a.membersDelete)
}

// GET /projects/{projectId}/members — команда проекта с ролями и уровнями доступа.
// 200 -> []ProjectMember.
func (a *API) membersList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects/{projectId}/members — добавить существующего пользователя в команду проекта.
// Body: ProjectMemberCreateRequest. 201 -> ProjectMember.
func (a *API) membersCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /projects/{projectId}/members/{memberId} — изменить роль/уровень доступа участника.
// Body: { projectRole, sprintRole, accessLevel }. 200 -> ProjectMember.
func (a *API) membersUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /projects/{projectId}/members/{memberId} — исключить участника из команды.
// Ранее назначенные ему задачи не удаляются. 204.
func (a *API) membersDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}

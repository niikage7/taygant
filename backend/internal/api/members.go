package api

import "github.com/gofiber/fiber/v2"

// registerMembersRoutes — тег Members: состав команды проекта и роли участников.
func registerMembersRoutes(r fiber.Router) {
	members := r.Group("/projects/:projectId/members")

	members.Get("/", membersList)
	members.Post("/", membersCreate)
	members.Patch("/:memberId", membersUpdate)
	members.Delete("/:memberId", membersDelete)
}

// GET /projects/{projectId}/members — команда проекта с ролями и уровнями доступа.
// 200 -> []ProjectMember.
func membersList(c *fiber.Ctx) error {
	return notImplemented(c)
}

// POST /projects/{projectId}/members — добавить существующего пользователя в команду проекта.
// Body: ProjectMemberCreateRequest. 201 -> ProjectMember.
func membersCreate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// PATCH /projects/{projectId}/members/{memberId} — изменить роль/уровень доступа участника.
// Body: { projectRole, sprintRole, accessLevel }. 200 -> ProjectMember.
func membersUpdate(c *fiber.Ctx) error {
	return notImplemented(c)
}

// DELETE /projects/{projectId}/members/{memberId} — исключить участника из команды.
// Ранее назначенные ему задачи не удаляются. 204.
func membersDelete(c *fiber.Ctx) error {
	return notImplemented(c)
}

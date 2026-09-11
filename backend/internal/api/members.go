package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// registerMembersRoutes — тег Members: состав команды проекта и роли участников.
func (a *API) registerMembersRoutes(r fiber.Router) {
	members := r.Group("/projects/:projectId/members")

	members.Get("/", a.membersList)
	members.Post("/", a.membersCreate)
	members.Patch("/:memberId", a.membersUpdate)
	members.Delete("/:memberId", a.membersDelete)
}

// projectMemberCreateRequest — тело добавления участника; используется и в
// POST /projects/{projectId}/members, и внутри ProjectCreateRequest.members.
type projectMemberCreateRequest struct {
	UserID      uuid.UUID          `json:"userId"`
	ProjectRole models.ProjectRole `json:"projectRole"`
	SprintRole  string             `json:"sprintRole"`
	AccessLevel models.AccessLevel `json:"accessLevel"`
}

// GET /projects/{projectId}/members — команда проекта с ролями и уровнями доступа.
// 200 -> []ProjectMember.
func (a *API) membersList(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireMember(c, projectID); err != nil {
		return err
	}

	members, err := a.members.List(c.Context(), projectID)
	if err != nil {
		return fail(err)
	}
	out := make([]projectMemberDTO, 0, len(members))
	for _, m := range members {
		out = append(out, newProjectMemberDTO(m))
	}
	return c.JSON(out)
}

// POST /projects/{projectId}/members — добавить существующего пользователя в команду проекта.
// Body: ProjectMemberCreateRequest. 201 -> ProjectMember.
func (a *API) membersCreate(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[projectMemberCreateRequest](c)
	if err != nil {
		return err
	}

	member, err := a.members.Add(c.Context(), projectID, service.MemberInput{
		UserID:      body.UserID,
		ProjectRole: body.ProjectRole,
		SprintRole:  body.SprintRole,
		AccessLevel: body.AccessLevel,
	})
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newProjectMemberDTO(member))
}

// memberUpdateRequest — тело PATCH /projects/{projectId}/members/{memberId}.
type memberUpdateRequest struct {
	ProjectRole *models.ProjectRole `json:"projectRole"`
	SprintRole  *string             `json:"sprintRole"`
	AccessLevel *models.AccessLevel `json:"accessLevel"`
}

// PATCH /projects/{projectId}/members/{memberId} — изменить роль/уровень доступа участника.
// Body: { projectRole, sprintRole, accessLevel }. 200 -> ProjectMember.
func (a *API) membersUpdate(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	memberID, err := pathUUID(c, "memberId")
	if err != nil {
		return err
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[memberUpdateRequest](c)
	if err != nil {
		return err
	}

	member, err := a.members.Update(c.Context(), projectID, memberID, service.MemberUpdateInput{
		ProjectRole: body.ProjectRole,
		SprintRole:  body.SprintRole,
		AccessLevel: body.AccessLevel,
	})
	if err != nil {
		return fail(err)
	}
	return c.JSON(newProjectMemberDTO(member))
}

// DELETE /projects/{projectId}/members/{memberId} — исключить участника из команды.
// Ранее назначенные ему задачи не удаляются. 204.
func (a *API) membersDelete(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	memberID, err := pathUUID(c, "memberId")
	if err != nil {
		return err
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	if err := a.members.Remove(c.Context(), projectID, memberID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

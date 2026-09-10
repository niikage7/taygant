package api

import (
	"github.com/gofiber/fiber/v2"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// registerProjectsRoutes — тег Projects: проекты и их параметры.
func (a *API) registerProjectsRoutes(r fiber.Router) {
	projects := r.Group("/projects")

	projects.Get("/", a.projectsList)
	projects.Post("/", a.projectsCreate)
	projects.Get("/:projectId", a.projectsGet)
	projects.Patch("/:projectId", a.projectsUpdate)
	projects.Delete("/:projectId", a.projectsDelete)
}

// GET /projects — список проектов пользователя с краткими сводными метриками.
// Query: status (ProjectStatus), search. 200 -> []ProjectSummary.
func (a *API) projectsList(c *fiber.Ctx) error {
	status := models.ProjectStatus(c.Query("status"))
	if status != "" && !status.Valid() {
		return fiber.NewError(fiber.StatusBadRequest, "недопустимый статус проекта")
	}

	projects, err := a.projects.List(c.Context(), currentUserID(c), status, c.Query("search"))
	if err != nil {
		return fail(err)
	}

	out := make([]projectSummaryDTO, 0, len(projects))
	for _, p := range projects {
		metrics, err := a.metricsFor(c, p)
		if err != nil {
			return fail(err)
		}
		out = append(out, newProjectSummaryDTO(p, metrics))
	}
	return c.JSON(out)
}

// projectCreateRequest — тело POST /projects.
type projectCreateRequest struct {
	Name                      string                       `json:"name"`
	Description               string                       `json:"description"`
	CustomerOrg               string                       `json:"customerOrg"`
	StartDate                 models.Date                  `json:"startDate"`
	Deadline                  models.Date                  `json:"deadline"`
	WorkingCalendarType       models.WorkingCalendarType   `json:"workingCalendarType"`
	IncludePublicHolidays     *bool                        `json:"includePublicHolidays"`
	AutoRecalculateDependents *bool                        `json:"autoRecalculateDependents"`
	HighlightCriticalPath     *bool                        `json:"highlightCriticalPath"`
	Members                   []projectMemberCreateRequest `json:"members"`
	SaveAsDraft               bool                         `json:"saveAsDraft"`
}

// POST /projects — создать проект (шаг 1 мастера инициации).
// Создаёт проект вместе с параметрами Ганта и опциональным начальным составом команды.
// Body: ProjectCreateRequest. 201 -> Project.
func (a *API) projectsCreate(c *fiber.Ctx) error {
	body, err := parseBody[projectCreateRequest](c)
	if err != nil {
		return err
	}

	members := make([]service.MemberInput, 0, len(body.Members))
	for _, m := range body.Members {
		members = append(members, service.MemberInput{
			UserID:      m.UserID,
			ProjectRole: m.ProjectRole,
			SprintRole:  m.SprintRole,
			AccessLevel: m.AccessLevel,
		})
	}

	project, err := a.projects.Create(c.Context(), currentUserID(c), service.CreateInput{
		Name:                      body.Name,
		Description:               body.Description,
		CustomerOrg:               body.CustomerOrg,
		StartDate:                 body.StartDate,
		Deadline:                  body.Deadline,
		WorkingCalendarType:       body.WorkingCalendarType,
		IncludePublicHolidays:     body.IncludePublicHolidays,
		AutoRecalculateDependents: body.AutoRecalculateDependents,
		HighlightCriticalPath:     body.HighlightCriticalPath,
		Members:                   members,
		SaveAsDraft:               body.SaveAsDraft,
	})
	if err != nil {
		return fail(err)
	}

	metrics, err := a.metricsFor(c, project)
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(newProjectDTO(project, metrics))
}

// GET /projects/{projectId} — полная карточка проекта, включая календарь и настройки Ганта.
// 200 -> Project, 404 -> не найден.
func (a *API) projectsGet(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	m, err := a.requireMember(c, projectID)
	if err != nil {
		return err
	}

	metrics, err := a.metricsFor(c, m.Project)
	if err != nil {
		return fail(err)
	}
	return c.JSON(newProjectDTO(m.Project, metrics))
}

// projectUpdateRequest — тело PATCH /projects/{projectId}.
type projectUpdateRequest struct {
	Name                *string                     `json:"name"`
	Description         *string                     `json:"description"`
	Status              *models.ProjectStatus       `json:"status"`
	Phase               *string                     `json:"phase"`
	StartDate           *models.Date                `json:"startDate"`
	Deadline            *models.Date                `json:"deadline"`
	WorkingCalendarType *models.WorkingCalendarType `json:"workingCalendarType"`
}

// PATCH /projects/{projectId} — изменить параметры проекта (сроки, статус, описание, календарь).
// Только участник с полным доступом: это настройки самого проекта, а не его задач.
// Body: ProjectUpdateRequest. 200 -> Project.
func (a *API) projectsUpdate(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	body, err := parseBody[projectUpdateRequest](c)
	if err != nil {
		return err
	}

	project, err := a.projects.Update(c.Context(), projectID, service.UpdateInput{
		Name:                body.Name,
		Description:         body.Description,
		Status:              body.Status,
		Phase:               body.Phase,
		StartDate:           body.StartDate,
		Deadline:            body.Deadline,
		WorkingCalendarType: body.WorkingCalendarType,
	})
	if err != nil {
		return fail(err)
	}

	metrics, err := a.metricsFor(c, project)
	if err != nil {
		return fail(err)
	}
	return c.JSON(newProjectDTO(project, metrics))
}

// DELETE /projects/{projectId} — удалить (архивировать) проект.
// Переводит проект в статус archived; данные не удаляются физически. 204.
func (a *API) projectsDelete(c *fiber.Ctx) error {
	projectID, err := pathUUID(c, "projectId")
	if err != nil {
		return err
	}
	if _, err := a.requireManage(c, projectID); err != nil {
		return err
	}

	if err := a.projects.Archive(c.Context(), projectID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// metricsFor — общая точка расчёта сводных показателей проекта; используется
// везде, где отдаётся ProjectSummary/Project.
func (a *API) metricsFor(c *fiber.Ctx, project models.Project) (service.Metrics, error) {
	return service.ComputeMetrics(c.Context(), a.db, project)
}

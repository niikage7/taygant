// Package api содержит HTTP-слой приложения: объявление и реализацию всех
// эндпоинтов, описанных в api-spec.yml. Файлы разбиты по тегам спецификации.
package api

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"taygant_backend/internal/auth"
	"taygant_backend/internal/config"
	"taygant_backend/internal/service"
)

// API хранит зависимости HTTP-слоя. Хендлеры объявлены как методы, чтобы
// добираться до сервисов без пакетных глобальных переменных.
type API struct {
	db     *gorm.DB
	cfg    config.Config
	tokens *auth.TokenIssuer

	auth         *service.Auth
	users        *service.Users
	projects     *service.Projects
	members      *service.Members
	sprints      *service.Sprints
	milestones   *service.Milestones
	tasks        *service.Tasks
	dependencies *service.Dependencies
	checklist    *service.Checklist
	comments     *service.Comments
	history      *service.History
	simulation   *service.Simulation
	dashboard    *service.Dashboard
	workload     *service.Workload
	export       *service.Export
	mcpTokens    *service.MCPTokens
}

// New собирает HTTP-слой поверх готового подключения к БД.
func New(db *gorm.DB, cfg config.Config) *API {
	tokens := auth.NewTokenIssuer(cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)

	// tasks и milestones собираются отдельно: Dashboard и Export переиспользуют
	// их логику (эффективные статусы, CPM, статус вехи), а не дублируют её.
	tasksSvc := service.NewTasks(db)
	milestonesSvc := service.NewMilestones(db)
	workloadSvc := service.NewWorkload(db)

	return &API{
		db:           db,
		cfg:          cfg,
		tokens:       tokens,
		auth:         service.NewAuth(db, tokens),
		users:        service.NewUsers(db),
		projects:     service.NewProjects(db),
		members:      service.NewMembers(db),
		sprints:      service.NewSprints(db),
		milestones:   milestonesSvc,
		tasks:        tasksSvc,
		dependencies: service.NewDependencies(db),
		checklist:    service.NewChecklist(db),
		comments:     service.NewComments(db),
		history:      service.NewHistory(db),
		simulation:   service.NewSimulation(db),
		dashboard:    service.NewDashboard(db, tasksSvc, milestonesSvc, workloadSvc),
		workload:     workloadSvc,
		export:       service.NewExport(db, tasksSvc, milestonesSvc),
		mcpTokens:    service.NewMCPTokens(db),
	}
}

// Register монтирует все маршруты API на базовый путь /api/v1.
// Переданные middleware применяются ко всей группе /api/v1, но не к health-check.
//
// Маршруты разделены на две группы. Всё, кроме входа и регистрации, лежит
// в protected: доступ по умолчанию закрыт, и новый эндпоинт нельзя случайно
// оставить открытым — для этого его пришлось бы намеренно объявить в v1.
func (a *API) Register(app *fiber.App, middlewares ...fiber.Handler) {
	v1 := app.Group("/api/v1", middlewares...)
	a.registerAuthRoutes(v1)

	protected := v1.Group("", a.requireAuth)

	a.registerUsersRoutes(protected)
	a.registerMCPTokensRoutes(protected)
	a.registerProjectsRoutes(protected)
	a.registerMembersRoutes(protected)
	a.registerSprintsRoutes(protected)
	a.registerMilestonesRoutes(protected)
	a.registerTasksRoutes(protected)
	a.registerDependenciesRoutes(protected)
	a.registerChecklistRoutes(protected)
	a.registerCommentsRoutes(protected)
	a.registerHistoryRoutes(protected)
	a.registerSimulationRoutes(protected)
	a.registerWorkloadRoutes(protected)
	a.registerDashboardRoutes(protected)
	a.registerExportRoutes(protected)
}

// notImplemented — временная заглушка для эндпоинтов, которые ещё не реализованы.
// Формат ответа совпадает с общим обработчиком ошибок в internal/server.
func notImplemented(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error":   "not implemented",
		"message": "not implemented",
		"path":    c.Path(),
	})
}

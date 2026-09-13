// Проверка витрины на настоящем PostgreSQL: сид прогоняется в отдельной
// временной схеме, а результат читается теми же сервисами, что отдают данные
// фронтенду. Даты витрины считаются от сегодняшнего дня, поэтому тест
// сторожит, чтобы она не «протухла»: связи не нарушены, на канбане заняты все
// колонки, на Ганте и обзоре есть что показать.
//
// Без TEST_DATABASE_URL тест пропускается — как интеграционные тесты API.
package database_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"taygant_backend/internal/database"
	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

func TestSeedShowcase(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL не задан — проверка сида пропущена")
	}
	db := openTestSchema(t, dsn)

	if err := database.Seed(db); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	// Повторный запуск — как рестарт контейнера: ничего не должен дублировать.
	if err := database.Seed(db); err != nil {
		t.Fatalf("повторный Seed: %v", err)
	}
	var projects int64
	if err := db.Model(&models.Project{}).Count(&projects).Error; err != nil {
		t.Fatal(err)
	}
	if projects != 4 {
		t.Fatalf("проектов %d, ожидалось 4: демо-проект и три проекта витрины", projects)
	}

	ctx := context.Background()
	tasksSvc := service.NewTasks(db)
	dashboardSvc := service.NewDashboard(db, tasksSvc, service.NewMilestones(db), service.NewWorkload(db))

	load := func(t *testing.T, code string) (models.Project, []models.Task, []models.TaskDependency) {
		t.Helper()
		var project models.Project
		if err := db.First(&project, "code = ?", code).Error; err != nil {
			t.Fatalf("найти проект %s: %v", code, err)
		}
		tasks, err := tasksSvc.List(ctx, project.ID, project, service.Filter{})
		if err != nil {
			t.Fatalf("задачи проекта %s: %v", code, err)
		}
		ids := make([]uuid.UUID, len(tasks))
		for i, task := range tasks {
			ids[i] = task.ID
		}
		var deps []models.TaskDependency
		if len(ids) > 0 {
			if err := db.Where("successor_task_id IN ?", ids).Find(&deps).Error; err != nil {
				t.Fatalf("связи проекта %s: %v", code, err)
			}
		}
		assertNoBrokenLinks(t, ctx, tasksSvc, project, tasks)
		return project, tasks, deps
	}

	t.Run("кабинет абитуриента", func(t *testing.T) {
		project, tasks, deps := load(t, "TPU-DEMO-CABINET")

		// Канбан: все хранимые и вычисленные статусы, а просрочка — и в «Плане»
		// (без прогресса), и «В работе» (с прогрессом).
		statuses := map[models.TaskStatus]int{}
		var overdueStarted, overdueNotStarted, unassigned bool
		milestoneTasks, grouped, ungrouped := 0, 0, 0
		criticalAtRisk := false
		for _, task := range tasks {
			statuses[task.Status]++
			if task.Status == models.TaskStatusOverdue {
				overdueStarted = overdueStarted || task.ProgressPercent > 0
				overdueNotStarted = overdueNotStarted || task.ProgressPercent == 0
			}
			unassigned = unassigned || task.AssigneeID == nil
			if task.IsMilestone {
				milestoneTasks++
			}
			if task.MilestoneID != nil {
				grouped++
			} else {
				ungrouped++
			}
			if task.IsCriticalPath && task.Status != models.TaskStatusDone && task.PlanVsActualDeviation() < 0 {
				criticalAtRisk = true
			}
		}
		for _, status := range []models.TaskStatus{models.TaskStatusPlanned, models.TaskStatusInProgress,
			models.TaskStatusDone, models.TaskStatusOverdue, models.TaskStatusBlocked} {
			if statuses[status] == 0 {
				t.Errorf("нет задач со статусом %s — на канбане не будет видно этого состояния", status)
			}
		}
		if !overdueStarted || !overdueNotStarted {
			t.Errorf("просрочка должна быть и начатой, и не начатой: начатая %v, не начатая %v", overdueStarted, overdueNotStarted)
		}
		if !unassigned {
			t.Error("нет задачи без ответственного")
		}

		// Гант.
		if !criticalAtRisk {
			t.Error("нет отстающей задачи на критическом пути — предупреждение над Гантом не появится")
		}
		if milestoneTasks < 2 || grouped == 0 || ungrouped == 0 {
			t.Errorf("задач-вех %d, в КТ-группах %d, вне вех %d — нужны все три вида строк", milestoneTasks, grouped, ungrouped)
		}
		types := map[models.DependencyType]bool{}
		predecessors := map[uuid.UUID]int{}
		lag := false
		for _, dep := range deps {
			types[dep.Type] = true
			predecessors[dep.SuccessorTaskID]++
			lag = lag || dep.LagDays > 0
		}
		if !types[models.DependencyFS] || !types[models.DependencySS] || !types[models.DependencyFF] || !lag {
			t.Errorf("типы связей %v, лаг %v — нужны FS, SS, FF и связь с лагом", types, lag)
		}
		maxPredecessors := 0
		for _, n := range predecessors {
			maxPredecessors = max(maxPredecessors, n)
		}
		if maxPredecessors < 3 {
			t.Errorf("больше всего предшественников у одной задачи: %d, нужно 3 — колонка «Пред.» сворачивает их в «#N +2»", maxPredecessors)
		}

		// Обзор.
		dashboard, err := dashboardSvc.Get(ctx, project.ID)
		if err != nil {
			t.Fatalf("дашборд: %v", err)
		}
		if dashboard.ScheduleAdherence.Status != "at_risk" {
			t.Errorf("соблюдение сроков %q, ожидалось at_risk (резерв %d дн., отклонение %d дн.)",
				dashboard.ScheduleAdherence.Status, dashboard.ScheduleAdherence.ProjectBufferDays,
				dashboard.ScheduleAdherence.ForecastDeviationDays)
		}
		if len(dashboard.Risks) == 0 || len(dashboard.AttentionTasks) == 0 {
			t.Errorf("рисков %d, задач с отставанием %d — карточки обзора будут пустыми", len(dashboard.Risks), len(dashboard.AttentionTasks))
		}
		overloaded := false
		for _, w := range dashboard.TeamWorkload {
			overloaded = overloaded || w.UtilizationPercent > 100
		}
		if !overloaded {
			t.Error("в текущем спринте никто не перегружен — карточка загрузки не покажет перегруз")
		}
	})

	t.Run("портал публикаций", func(t *testing.T) {
		project, tasks, _ := load(t, "TPU-DEMO-SCIENCE")
		if len(tasks) == 0 {
			t.Fatal("в проекте нет задач")
		}
		dashboard, err := dashboardSvc.Get(ctx, project.ID)
		if err != nil {
			t.Fatalf("дашборд: %v", err)
		}
		if dashboard.ScheduleAdherence.Status != "on_track" {
			t.Errorf("соблюдение сроков %q, ожидалось on_track", dashboard.ScheduleAdherence.Status)
		}
	})

	t.Run("черновик чат-бота", func(t *testing.T) {
		project, tasks, _ := load(t, "TPU-DEMO-CHATBOT")
		if project.Status != models.ProjectStatusDraft || len(tasks) != 0 {
			t.Errorf("статус %q, задач %d — ожидался черновик без задач", project.Status, len(tasks))
		}
	})
}

// assertNoBrokenLinks сверяет плановые даты с CPM: ранний старт, посчитанный
// по связям, не должен быть позже планового. Иначе задача начинается раньше,
// чем позволяют её предшественники, и Гант рисует стрелку назад.
func assertNoBrokenLinks(t *testing.T, ctx context.Context, tasks *service.Tasks, project models.Project, list []models.Task) {
	t.Helper()
	summary, err := tasks.ComputeSchedule(ctx, project.ID)
	if err != nil {
		t.Fatalf("CPM проекта %s: %v", project.Code, err)
	}
	for _, task := range list {
		early := summary.Tasks[task.ID].EarlyStart
		if !early.Equal(task.StartDate) {
			t.Errorf("%s «%s»: начало %s, а связи позволяют не раньше %s", project.Code, task.Title, task.StartDate, early)
		}
	}
}

// openTestSchema создаёт временную схему, применяет миграции и удаляет схему
// по окончании теста.
func openTestSchema(t *testing.T, dsn string) *gorm.DB {
	t.Helper()
	schema := fmt.Sprintf("seed_test_%d", time.Now().UnixNano())

	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("подключение к тестовой БД: %v", err)
	}
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatalf("создание тестовой схемы: %v", err)
	}
	t.Cleanup(func() {
		admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		if sqlDB, err := admin.DB(); err == nil {
			sqlDB.Close()
		}
	})

	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL: %v", err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()

	db, err := database.Connect(u.String(), 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

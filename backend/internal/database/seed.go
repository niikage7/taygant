package database

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/auth"
	"taygant_backend/internal/models"
)

// DemoPassword — пароль всех демонстрационных учётных записей.
// Он же печатается в лог при первом запуске, чтобы не искать его в коде перед показом.
const DemoPassword = "demo1234"

// demoProjectCode — фиксированный код демо-проекта. Именно по нему, а не по
// количеству пользователей, проверяется идемпотентность посева проекта:
// пользователи и проект могли быть созданы в разные запуски (например, если
// проект добавили в сидер уже после первого docker compose up).
const demoProjectCode = "TPU-DEMO-GANTT"

// demoUsers — команда для демонстрационного проекта.
// Роли в проекте здесь не задаются: они относятся к проекту, а не к человеку,
// и назначаются при создании проекта в project_members.
var demoUsers = []models.User{
	{Email: "volkov@taygant.dev", FullName: "Александр Волков", Department: "Институт кибернетики", Position: "Lead Fullstack"},
	{Email: "orlova@taygant.dev", FullName: "Мария Орлова", Department: "Институт кибернетики", Position: "Бизнес-аналитик"},
	{Email: "petrov@taygant.dev", FullName: "Дмитрий Петров", Department: "Отдел разработки", Position: "Backend-разработчик"},
	{Email: "smirnova@taygant.dev", FullName: "Елена Смирнова", Department: "Отдел разработки", Position: "Frontend-разработчик"},
	{Email: "kuznetsov@taygant.dev", FullName: "Игорь Кузнецов", Department: "Дирекция", Position: "Куратор проекта"},
	{Email: "ivanova@taygant.dev", FullName: "Ольга Иванова", Department: "Отдел качества", Position: "QA-инженер"},
}

// Seed наполняет пустую БД демонстрационными данными: командой, проектом,
// спринтами, вехами и задачами с зависимостями, а также проектами витрины
// возможностей (см. seed_showcase.go).
//
// Пользователи и демо-проект проверяются на идемпотентность раздельно (см.
// demoProjectCode): один общий счётчик "есть хоть один пользователь" не подошёл
// бы, если проект добавили в сидер позже первого запуска.
func Seed(db *gorm.DB) error {
	users, err := seedUsers(db)
	if err != nil {
		return err
	}
	if err := seedDemoProject(db, users); err != nil {
		return err
	}
	return seedShowcase(db)
}

// seedUsers создаёт демо-команду, если таблица пользователей пуста, и в любом
// случае возвращает актуальный список — он нужен seedDemoProject для ссылок
// на авторов и ответственных по email.
func seedUsers(db *gorm.DB) ([]models.User, error) {
	var existing []models.User
	if err := db.Order("full_name").Find(&existing).Error; err != nil {
		return nil, fmt.Errorf("выбрать существующих пользователей: %w", err)
	}
	if len(existing) > 0 {
		return existing, nil
	}

	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return nil, fmt.Errorf("подготовить пароль демо-пользователей: %w", err)
	}

	// Хеш считается один раз и переиспользуется: bcrypt намеренно медленный,
	// и шесть отдельных вычислений заметно задержали бы старт приложения.
	users := make([]models.User, len(demoUsers))
	for i, u := range demoUsers {
		u.PasswordHash = hash
		users[i] = u
	}

	if err := db.Create(&users).Error; err != nil {
		return nil, fmt.Errorf("создать демо-пользователей: %w", err)
	}

	log.Printf("создано демо-пользователей: %d, пароль у всех — %q", len(users), DemoPassword)
	return users, nil
}

// seedDemoProject создаёт один демонстрационный проект со спринтами, вехами,
// задачами и зависимостями между ними — цепочка "Анализ -> Дизайн -> Разработка
// -> Тестирование" из кейса, плюс параллельная ветка и заведомо просроченная и
// заблокированная задачи, чтобы на демо сразу было видно автоматический расчёт
// статусов (см. service.Tasks.resolveStatuses), а не только пустой реестр.
//
// Идемпотентна по demoProjectCode: повторный запуск при уже существующем
// демо-проекте ничего не делает и не портит ручные правки, внесённые на демо.
func seedDemoProject(db *gorm.DB, users []models.User) error {
	var count int64
	if err := db.Model(&models.Project{}).Where("code = ?", demoProjectCode).Count(&count).Error; err != nil {
		return fmt.Errorf("проверить наличие демо-проекта: %w", err)
	}
	if count > 0 {
		return nil
	}

	byEmail := make(map[string]models.User, len(users))
	for _, u := range users {
		byEmail[u.Email] = u
	}
	get := func(email string) models.User { return byEmail[email] }

	volkov := get("volkov@taygant.dev")
	orlova := get("orlova@taygant.dev")
	petrov := get("petrov@taygant.dev")
	smirnova := get("smirnova@taygant.dev")
	kuznetsov := get("kuznetsov@taygant.dev")
	ivanova := get("ivanova@taygant.dev")

	if volkov.ID == uuid.Nil || orlova.ID == uuid.Nil || petrov.ID == uuid.Nil ||
		smirnova.ID == uuid.Nil || kuznetsov.ID == uuid.Nil || ivanova.ID == uuid.Nil {
		// Демо-пользователей переименовали или удалили не через сидер —
		// продолжать посев проекта с неполной командой означало бы молча
		// потерять часть назначений, поэтому лучше явно ничего не делать.
		log.Printf("посев демо-проекта пропущен: не найдены все демо-пользователи по email")
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		project := models.Project{
			Code:                      demoProjectCode,
			Name:                      "Внедрение системы мониторинга оборудования",
			Description:               "Автоматизация сбора показаний оборудования и уведомлений об отклонениях для Института кибернетики ТПУ.",
			CustomerOrg:               "Институт кибернетики ТПУ",
			Status:                    models.ProjectStatusActive,
			Phase:                     "Фаза 2: Разработка и тестирование",
			StartDate:                 models.NewDate(2026, 8, 18),
			Deadline:                  models.NewDate(2026, 9, 30),
			WorkingCalendarType:       models.Calendar52,
			AutoRecalculateDependents: true,
			HighlightCriticalPath:     true,
			CreatedByID:               volkov.ID,
		}
		if err := tx.Create(&project).Error; err != nil {
			return fmt.Errorf("создать демо-проект: %w", err)
		}

		members := []models.ProjectMember{
			{ProjectID: project.ID, UserID: volkov.ID, ProjectRole: models.ProjectRoleProjectManager, SprintRole: "Руководитель проекта", AccessLevel: models.AccessLevelFull, WeeklyHoursLimit: 40},
			{ProjectID: project.ID, UserID: orlova.ID, ProjectRole: models.ProjectRoleAnalyst, SprintRole: "Аналитик", AccessLevel: models.AccessLevelEdit, WeeklyHoursLimit: 40},
			{ProjectID: project.ID, UserID: petrov.ID, ProjectRole: models.ProjectRoleDeveloper, SprintRole: "Backend-разработчик", AccessLevel: models.AccessLevelEdit, WeeklyHoursLimit: 40},
			{ProjectID: project.ID, UserID: smirnova.ID, ProjectRole: models.ProjectRoleDeveloper, SprintRole: "Frontend-разработчик", AccessLevel: models.AccessLevelEdit, WeeklyHoursLimit: 40},
			{ProjectID: project.ID, UserID: kuznetsov.ID, ProjectRole: models.ProjectRoleCurator, SprintRole: "Куратор", AccessLevel: models.AccessLevelView, WeeklyHoursLimit: 10},
			{ProjectID: project.ID, UserID: ivanova.ID, ProjectRole: models.ProjectRoleOther, SprintRole: "QA-инженер", AccessLevel: models.AccessLevelEdit, WeeklyHoursLimit: 40},
		}
		if err := tx.Create(&members).Error; err != nil {
			return fmt.Errorf("добавить команду демо-проекта: %w", err)
		}

		sprints := []models.Sprint{
			{ProjectID: project.ID, Name: "Спринт 1 · Анализ и дизайн", ReleaseVersion: "0.1", StartDate: models.NewDate(2026, 8, 18), EndDate: models.NewDate(2026, 8, 31)},
			{ProjectID: project.ID, Name: "Спринт 2 · Разработка", ReleaseVersion: "0.2", StartDate: models.NewDate(2026, 9, 1), EndDate: models.NewDate(2026, 9, 16)},
			{ProjectID: project.ID, Name: "Спринт 3 · Тестирование и релиз", ReleaseVersion: "1.0", StartDate: models.NewDate(2026, 9, 17), EndDate: models.NewDate(2026, 9, 30)},
		}
		if err := tx.Create(&sprints).Error; err != nil {
			return fmt.Errorf("создать спринты демо-проекта: %w", err)
		}
		sprint1, sprint2, sprint3 := sprints[0].ID, sprints[1].ID, sprints[2].ID

		analysisDone := models.NewDate(2026, 8, 24)
		milestones := []models.Milestone{
			{ProjectID: project.ID, Code: "КТ-1", Name: "Анализ и дизайн завершены", PlannedDate: analysisDone, ActualDate: &analysisDone, Status: models.MilestoneStatusDone},
			{ProjectID: project.ID, Code: "КТ-2", Name: "Готовность MVP", PlannedDate: models.NewDate(2026, 9, 16), Status: models.MilestoneStatusPlanned},
			{ProjectID: project.ID, Code: "КТ-3", Name: "Финальный релиз", PlannedDate: models.NewDate(2026, 9, 30), Status: models.MilestoneStatusPlanned},
		}
		if err := tx.Create(&milestones).Error; err != nil {
			return fmt.Errorf("создать вехи демо-проекта: %w", err)
		}

		// Задачи создаются по одной (не батчем), чтобы аккуратно проставить код,
		// WBS-номер и порядок так же, как это делает сервис задач — сид не должен
		// разъезжаться с реальным поведением API.
		mk := func(sprintID uuid.UUID, order int, code, wbs, title string, assignee models.User, status models.TaskStatus, progress int, start, end models.Date, actualStart, actualEnd *models.Date, weight float64) models.Task {
			return models.Task{
				ProjectID:       project.ID,
				SprintID:        &sprintID,
				Code:            code,
				WBSNumber:       wbs,
				OrderIndex:      order,
				Title:           title,
				AssigneeID:      &assignee.ID,
				Status:          status,
				StartDate:       start,
				EndDate:         end,
				ActualStartDate: actualStart,
				ActualEndDate:   actualEnd,
				ProgressPercent: progress,
				WeightPercent:   weight,
			}
		}
		date := models.NewDate
		dp := func(d models.Date) *models.Date { return &d }

		t1 := mk(sprint1, 0, "TASK-001", "1", "Анализ требований", orlova, models.TaskStatusDone, 100,
			date(2026, 8, 18), date(2026, 8, 20), dp(date(2026, 8, 18)), dp(date(2026, 8, 20)), 10)
		t1.MilestoneID = &milestones[0].ID
		t2 := mk(sprint1, 1, "TASK-002", "2", "Дизайн архитектуры", orlova, models.TaskStatusDone, 100,
			date(2026, 8, 21), date(2026, 8, 24), dp(date(2026, 8, 21)), dp(date(2026, 8, 24)), 10)
		t2.MilestoneID = &milestones[0].ID
		t3 := mk(sprint1, 2, "TASK-003", "3", "Прототип интерфейса", smirnova, models.TaskStatusDone, 100,
			date(2026, 8, 21), date(2026, 8, 25), dp(date(2026, 8, 22)), dp(date(2026, 8, 25)), 10)
		t3.MilestoneID = &milestones[0].ID
		// Плановый конец раньше сегодняшней даты и задача не завершена — сервис
		// на чтении сам покажет её как overdue (см. resolveStatuses), это не баг сида.
		t4 := mk(sprint2, 3, "TASK-004", "4", "Разработка API сбора показаний", petrov, models.TaskStatusInProgress, 60,
			date(2026, 9, 1), date(2026, 9, 10), dp(date(2026, 9, 1)), nil, 15)
		t4.MilestoneID = &milestones[1].ID
		t5 := mk(sprint2, 4, "TASK-005", "5", "Диаграмма Ганта на фронтенде", smirnova, models.TaskStatusInProgress, 40,
			date(2026, 9, 3), date(2026, 9, 16), dp(date(2026, 9, 3)), nil, 15)
		t5.MilestoneID = &milestones[1].ID
		// Предшественник (TASK-004) ещё не done — сервис на чтении покажет
		// эту задачу как blocked, хотя в таблице лежит planned.
		t6 := mk(sprint2, 5, "TASK-006", "6", "Интеграция с БД показаний", petrov, models.TaskStatusPlanned, 0,
			date(2026, 9, 11), date(2026, 9, 15), nil, nil, 10)
		t6.MilestoneID = &milestones[1].ID
		t7 := mk(sprint3, 6, "TASK-007", "7", "Тестирование", ivanova, models.TaskStatusPlanned, 0,
			date(2026, 9, 17), date(2026, 9, 22), nil, nil, 10)
		t7.MilestoneID = &milestones[2].ID
		t8 := mk(sprint3, 7, "TASK-008", "8", "Исправление дефектов", petrov, models.TaskStatusPlanned, 0,
			date(2026, 9, 23), date(2026, 9, 25), nil, nil, 5)
		t8.MilestoneID = &milestones[2].ID
		t9 := mk(sprint3, 8, "TASK-009", "9", "Внедрение и приёмка", volkov, models.TaskStatusPlanned, 0,
			date(2026, 9, 28), date(2026, 9, 30), nil, nil, 10)
		t9.MilestoneID = &milestones[2].ID
		t10 := mk(sprint2, 9, "TASK-010", "10", "Документация пользователя", orlova, models.TaskStatusPlanned, 0,
			date(2026, 9, 15), date(2026, 9, 20), nil, nil, 5)
		// Задача-веха: событие нулевой длительности на графике, привязанное к
		// КТ-2 «Готовность MVP» — не путать саму КТ (models.Milestone) с этой
		// задачей (models.Task.IsMilestone), см. комментарий над seedDemoProject.
		mvpReady := mk(sprint2, 10, "TASK-011", "11", "Готовность MVP", volkov, models.TaskStatusPlanned, 0,
			date(2026, 9, 16), date(2026, 9, 16), nil, nil, 0)
		mvpReady.IsMilestone = true
		mvpReady.MilestoneID = &milestones[1].ID

		tasks := []models.Task{t1, t2, t3, t4, t5, t6, t7, t8, t9, t10, mvpReady}
		if err := tx.Create(&tasks).Error; err != nil {
			return fmt.Errorf("создать задачи демо-проекта: %w", err)
		}

		// FS-цепочка "Анализ -> Дизайн -> Разработка -> Тестирование" из кейса,
		// плюс параллельная ветка прототипа UI и независимая документация.
		type link struct{ from, to int }
		links := []link{
			{0, 1},  // Анализ -> Дизайн
			{0, 2},  // Анализ -> Прототип UI
			{1, 3},  // Дизайн -> API
			{2, 4},  // Прототип UI -> Гант на фронте
			{3, 5},  // API -> Интеграция с БД
			{4, 6},  // Гант на фронте -> Тестирование
			{5, 6},  // Интеграция с БД -> Тестирование
			{6, 7},  // Тестирование -> Исправление дефектов
			{7, 8},  // Исправление дефектов -> Внедрение
			{4, 10}, // Гант на фронте -> Готовность MVP (веха)
			{5, 10}, // Интеграция с БД -> Готовность MVP (веха)
		}
		deps := make([]models.TaskDependency, 0, len(links))
		for _, l := range links {
			deps = append(deps, models.TaskDependency{
				PredecessorTaskID: tasks[l.from].ID,
				SuccessorTaskID:   tasks[l.to].ID,
				Type:              models.DependencyFS,
			})
		}
		if err := tx.Create(&deps).Error; err != nil {
			return fmt.Errorf("создать связи демо-задач: %w", err)
		}

		// tasks[i], а не t1/t4 напрямую: tx.Create(&tasks) записал сгенерированные ID
		// в элементы среза tasks, а не в t1..t10 — те остались копиями со старым
		// (нулевым) ID, потому что models.Task передаётся по значению.
		checklist := []models.ChecklistItem{
			{TaskID: tasks[0].ID, Text: "Собраны требования от заказчика", IsDone: true, OrderIndex: 0},
			{TaskID: tasks[0].ID, Text: "Требования согласованы с куратором", IsDone: true, OrderIndex: 1},
			{TaskID: tasks[3].ID, Text: "Эндпоинт приёма показаний готов", IsDone: true, OrderIndex: 0},
			{TaskID: tasks[3].ID, Text: "Покрыт тестами", IsDone: false, OrderIndex: 1},
		}
		if err := tx.Create(&checklist).Error; err != nil {
			return fmt.Errorf("создать чек-лист демо-задач: %w", err)
		}

		comments := []models.TaskComment{
			{TaskID: tasks[3].ID, AuthorID: volkov.ID, Text: "Как продвигается? Не успеваем к спринту."},
			{TaskID: tasks[3].ID, AuthorID: petrov.ID, Text: "Основная часть готова, доделываю обработку ошибок."},
		}
		if err := tx.Create(&comments).Error; err != nil {
			return fmt.Errorf("создать комментарии демо-задач: %w", err)
		}

		history := make([]models.TaskHistoryEntry, 0, len(tasks))
		for _, t := range tasks {
			history = append(history, models.TaskHistoryEntry{TaskID: t.ID, ActorID: volkov.ID, Action: models.HistoryActionCreated})
		}
		if err := tx.Create(&history).Error; err != nil {
			return fmt.Errorf("создать историю демо-задач: %w", err)
		}

		log.Printf("создан демо-проект %q: %d задач, %d спринтов, %d вех", project.Code, len(tasks), len(sprints), len(milestones))
		return nil
	})
}

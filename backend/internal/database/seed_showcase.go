package database

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"taygant_backend/internal/auth"
	"taygant_backend/internal/models"
)

// Витрина возможностей: проекты, на которых видно всё, что умеет фронтенд.
// Главное — канбан с задачами во всех колонках (включая просроченные и
// заблокированные) и диаграмма Ганта с группами контрольных точек, вехами,
// критическим путём под угрозой, связями FS/SS/FF, лагом и задачами с
// несколькими предшественниками.
//
// В отличие от seedDemoProject, даты здесь считаются от дня посева в рабочих
// днях календаря проекта. Фиксированные даты через месяц превращают витрину в
// сплошь просроченный план, а относительные держат на экране и текущий спринт,
// и просрочку, и работы впереди. Связи разложены без нарушений: каждая задача
// начинается не раньше, чем позволяют её предшественники (это проверяет
// seed_showcase_test.go тем же CPM, что считает Гант).
//
// Каждый проект идемпотентен по своему коду, как и демо-проект.

const (
	showcaseCabinetCode = "TPU-DEMO-CABINET"
	showcaseScienceCode = "TPU-DEMO-SCIENCE"
	showcaseChatbotCode = "TPU-DEMO-CHATBOT"
)

const (
	emailVolkov    = "volkov@taygant.dev"
	emailOrlova    = "orlova@taygant.dev"
	emailPetrov    = "petrov@taygant.dev"
	emailSmirnova  = "smirnova@taygant.dev"
	emailKuznetsov = "kuznetsov@taygant.dev"
	emailIvanova   = "ivanova@taygant.dev"
	emailSokolov   = "sokolov@taygant.dev"
)

// showcaseUsers — участники витрины сверх demoUsers.
var showcaseUsers = []models.User{
	{Email: emailSokolov, FullName: "Никита Соколов", Department: "Отдел разработки", Position: "UI/UX-дизайнер"},
}

// seedShowcase наполняет БД проектами витрины. Вызывается из Seed после
// демо-проекта: демо-команда к этому моменту уже создана.
func seedShowcase(db *gorm.DB) error {
	team, ok, err := showcaseTeam(db)
	if err != nil || !ok {
		return err
	}
	for _, spec := range []showcaseProject{showcaseCabinet(), showcaseScience(), showcaseChatbot()} {
		if err := seedShowcaseProject(db, team, spec); err != nil {
			return err
		}
	}
	return nil
}

// showcaseTeam находит демо-команду и досоздаёт участников витрины. Как и
// seedDemoProject, без полной демо-команды ничего не делает: учётки с общим
// демо-паролем появляются только в базе, которую сидер наполнял с нуля.
func showcaseTeam(db *gorm.DB) (map[string]models.User, bool, error) {
	emails := make([]string, 0, len(demoUsers)+len(showcaseUsers))
	for _, u := range demoUsers {
		emails = append(emails, u.Email)
	}
	for _, u := range showcaseUsers {
		emails = append(emails, u.Email)
	}

	var found []models.User
	if err := db.Where("email IN ?", emails).Find(&found).Error; err != nil {
		return nil, false, fmt.Errorf("выбрать команду витрины: %w", err)
	}
	team := make(map[string]models.User, len(emails))
	for _, u := range found {
		team[u.Email] = u
	}
	for _, u := range demoUsers {
		if _, ok := team[u.Email]; !ok {
			log.Printf("посев витрины пропущен: не найден демо-пользователь %s", u.Email)
			return nil, false, nil
		}
	}

	var missing []models.User
	for _, u := range showcaseUsers {
		if _, ok := team[u.Email]; !ok {
			missing = append(missing, u)
		}
	}
	if len(missing) == 0 {
		return team, true, nil
	}
	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return nil, false, fmt.Errorf("подготовить пароль участников витрины: %w", err)
	}
	for i := range missing {
		missing[i].PasswordHash = hash
	}
	if err := db.Create(&missing).Error; err != nil {
		return nil, false, fmt.Errorf("создать участников витрины: %w", err)
	}
	for _, u := range missing {
		team[u.Email] = u
	}
	return team, true, nil
}

// ─── Описание проекта витрины ────────────────────────────────────────────────

// showcaseProject — проект витрины. Все дни — рабочие дни календаря проекта
// относительно сегодняшнего (см. planClock): 0 — сегодня, -3 — три рабочих дня
// назад, 5 — через пять рабочих дней.
type showcaseProject struct {
	clock                             planClock
	code, name, description, customer string
	phase                             string
	status                            models.ProjectStatus
	highlightCriticalPath             bool
	start, deadline                   int
	owner                             string
	members                           []showcaseMember
	sprints                           []showcaseSprint
	milestones                        []showcaseMilestone
	tasks                             []showcaseTask
	links                             []showcaseLink
	checklist                         []showcaseCheck
	comments                          []showcaseComment
	events                            []showcaseEvent
}

type showcaseMember struct {
	email      string
	role       models.ProjectRole
	sprintRole string
	access     models.AccessLevel
	hours      int
}

type showcaseSprint struct {
	name, release string
	start, end    int
}

// showcaseMilestone — контрольная точка: заголовок группы задач на Ганте.
type showcaseMilestone struct {
	code, name string
	planned    int
	reached    *int // nil — ещё не достигнута
}

type showcaseTask struct {
	key         string // ссылка на задачу из связей, чек-листов и журнала
	title       string
	description string
	assignee    string // email; пусто — без ответственного
	milestone   string // код КТ; пусто — вне вех
	sprint      int    // индекс в sprints; -1 — вне спринтов
	// milestoneTask — задача-веха: точка нулевой длительности (end не нужен).
	milestoneTask bool
	// status — хранимый статус. overdue и blocked сервис вычислит сам при чтении.
	status             models.TaskStatus
	progress           int
	start, end         int
	factStart, factEnd *int
	weight             float64
	external           *showcaseExternal
}

type showcaseExternal struct {
	provider, url, referenceID string
}

type showcaseLink struct {
	from, to string
	typ      models.DependencyType
	lag      int
}

type showcaseCheck struct {
	task, text string
	done       bool
}

type showcaseComment struct {
	task, author, text string
	at                 moment
}

// showcaseEvent — запись журнала задачи, как её пишет сервис задач.
type showcaseEvent struct {
	task, actor, action       string
	field, oldValue, newValue string
	at                        moment
}

// moment — время внутри рабочего дня. Часы по UTC: в Томске (UTC+7) и Москве
// (UTC+3) значения 2–12 приходятся на рабочее время.
type moment struct{ day, hour, minute int }

func when(day, hour, minute int) moment { return moment{day, hour, minute} }

func fact(day int) *int { return &day }

func statusEvent(task, actor string, from, to models.TaskStatus, at moment) showcaseEvent {
	return showcaseEvent{task: task, actor: actor, action: models.HistoryActionStatusChanged,
		field: "status", oldValue: string(from), newValue: string(to), at: at}
}

func progressEvent(task, actor string, from, to int, at moment) showcaseEvent {
	return showcaseEvent{task: task, actor: actor, action: models.HistoryActionProgressSet,
		field: "progressPercent", oldValue: strconv.Itoa(from), newValue: strconv.Itoa(to), at: at}
}

// ─── Календарь витрины ───────────────────────────────────────────────────────

// planClock раскладывает план в рабочих днях от сегодняшнего дня.
type planClock struct {
	calendar models.WorkingCalendarType
	// base — сегодня, а в выходной — последний рабочий день перед ним.
	base models.Date
	now  time.Time
}

func newPlanClock(calendar models.WorkingCalendarType) planClock {
	now := time.Now().UTC()
	base := models.DateOf(now)
	for !isShowcaseWorkday(calendar, base) {
		base = base.AddDays(-1)
	}
	return planClock{calendar: calendar, base: base, now: now}
}

// isShowcaseWorkday повторяет правило internal/schedule: воскресенье выходной
// в обоих календарях, суббота — в 5/2.
func isShowcaseWorkday(calendar models.WorkingCalendarType, d models.Date) bool {
	switch d.Weekday() {
	case time.Sunday:
		return false
	case time.Saturday:
		return calendar != models.Calendar52
	default:
		return true
	}
}

// day — дата через n рабочих дней от base; n < 0 — в прошлом.
func (c planClock) day(n int) models.Date {
	d := c.base
	step := 1
	if n < 0 {
		step, n = -1, -n
	}
	for n > 0 {
		d = d.AddDays(step)
		if isShowcaseWorkday(c.calendar, d) {
			n--
		}
	}
	return d
}

func (c planClock) dayPtr(n int) *models.Date {
	d := c.day(n)
	return &d
}

// at — момент записи журнала. Не позже текущего: запись из будущего выглядела
// бы в журнале ошибкой.
func (c planClock) at(m moment) time.Time {
	d := c.day(m.day)
	t := time.Date(d.Year(), d.Month(), d.Day(), m.hour, m.minute, 0, 0, time.UTC)
	if t.After(c.now) {
		return c.now
	}
	return t
}

// ─── Посев ───────────────────────────────────────────────────────────────────

// seedShowcaseProject создаёт проект витрины целиком в одной транзакции.
func seedShowcaseProject(db *gorm.DB, team map[string]models.User, spec showcaseProject) error {
	var count int64
	if err := db.Model(&models.Project{}).Where("code = ?", spec.code).Count(&count).Error; err != nil {
		return fmt.Errorf("проверить наличие проекта %s: %w", spec.code, err)
	}
	if count > 0 {
		return nil
	}

	c := spec.clock
	wrap := func(what string, err error) error {
		return fmt.Errorf("проект витрины %s: %s: %w", spec.code, what, err)
	}
	userID := func(email string) (uuid.UUID, error) {
		u, ok := team[email]
		if !ok {
			return uuid.Nil, fmt.Errorf("проект витрины %s: %s не входит в команду витрины", spec.code, email)
		}
		return u.ID, nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		ownerID, err := userID(spec.owner)
		if err != nil {
			return err
		}
		project := models.Project{
			Code:                      spec.code,
			Name:                      spec.name,
			Description:               spec.description,
			CustomerOrg:               spec.customer,
			Status:                    spec.status,
			Phase:                     spec.phase,
			StartDate:                 c.day(spec.start),
			Deadline:                  c.day(spec.deadline),
			WorkingCalendarType:       c.calendar,
			AutoRecalculateDependents: true,
			HighlightCriticalPath:     spec.highlightCriticalPath,
			CreatedByID:               ownerID,
		}
		if err := tx.Create(&project).Error; err != nil {
			return wrap("создать проект", err)
		}

		members := make([]models.ProjectMember, 0, len(spec.members))
		for _, m := range spec.members {
			id, err := userID(m.email)
			if err != nil {
				return err
			}
			members = append(members, models.ProjectMember{
				ProjectID:        project.ID,
				UserID:           id,
				ProjectRole:      m.role,
				SprintRole:       m.sprintRole,
				AccessLevel:      m.access,
				WeeklyHoursLimit: m.hours,
			})
		}
		if err := createAll(tx, members); err != nil {
			return wrap("добавить команду", err)
		}

		sprints := make([]models.Sprint, len(spec.sprints))
		for i, s := range spec.sprints {
			sprints[i] = models.Sprint{ProjectID: project.ID, Name: s.name, ReleaseVersion: s.release,
				StartDate: c.day(s.start), EndDate: c.day(s.end)}
		}
		if err := createAll(tx, sprints); err != nil {
			return wrap("создать спринты", err)
		}

		milestones := make([]models.Milestone, len(spec.milestones))
		for i, m := range spec.milestones {
			milestones[i] = models.Milestone{ProjectID: project.ID, Code: m.code, Name: m.name,
				PlannedDate: c.day(m.planned), Status: models.MilestoneStatusPlanned}
			// Статус вехи сервис выводит из фактической даты при чтении; в
			// таблицу кладём согласованное с ней значение.
			if m.reached != nil {
				milestones[i].ActualDate = c.dayPtr(*m.reached)
				milestones[i].Status = models.MilestoneStatusDone
			}
		}
		if err := createAll(tx, milestones); err != nil {
			return wrap("создать контрольные точки", err)
		}
		milestoneIDs := make(map[string]uuid.UUID, len(milestones))
		for _, m := range milestones {
			milestoneIDs[m.Code] = m.ID
		}

		tasks := make([]models.Task, len(spec.tasks))
		for i, t := range spec.tasks {
			task := models.Task{
				ProjectID: project.ID,
				// Код, WBS-номер и порядок — как их выдаёт сервис задач задачам
				// верхнего уровня, созданным по очереди.
				Code:            fmt.Sprintf("TASK-%03d", i+1),
				WBSNumber:       strconv.Itoa(i + 1),
				OrderIndex:      i,
				Title:           t.title,
				Description:     t.description,
				Status:          t.status,
				IsMilestone:     t.milestoneTask,
				StartDate:       c.day(t.start),
				EndDate:         c.day(t.end),
				ProgressPercent: t.progress,
				WeightPercent:   t.weight,
			}
			if t.milestoneTask {
				task.EndDate = task.StartDate
			}
			if t.assignee != "" {
				id, err := userID(t.assignee)
				if err != nil {
					return err
				}
				task.AssigneeID = &id
			}
			if t.sprint >= 0 {
				task.SprintID = &sprints[t.sprint].ID
			}
			if t.milestone != "" {
				id, ok := milestoneIDs[t.milestone]
				if !ok {
					return fmt.Errorf("проект витрины %s: задача %q ссылается на неизвестную КТ %q", spec.code, t.key, t.milestone)
				}
				task.MilestoneID = &id
			}
			if t.factStart != nil {
				task.ActualStartDate = c.dayPtr(*t.factStart)
			}
			if t.factEnd != nil {
				task.ActualEndDate = c.dayPtr(*t.factEnd)
			}
			if t.external != nil {
				task.ExternalProvider = t.external.provider
				task.ExternalURL = t.external.url
				task.ExternalReferenceID = t.external.referenceID
				task.ExternalSynced = true
			}
			tasks[i] = task
		}
		if err := createAll(tx, tasks); err != nil {
			return wrap("создать задачи", err)
		}
		taskIDs := make(map[string]uuid.UUID, len(tasks))
		for i, t := range spec.tasks {
			taskIDs[t.key] = tasks[i].ID
		}
		taskID := func(key string) (uuid.UUID, error) {
			id, ok := taskIDs[key]
			if !ok {
				return uuid.Nil, fmt.Errorf("проект витрины %s: неизвестная задача %q", spec.code, key)
			}
			return id, nil
		}

		deps := make([]models.TaskDependency, 0, len(spec.links))
		for _, l := range spec.links {
			from, err := taskID(l.from)
			if err != nil {
				return err
			}
			to, err := taskID(l.to)
			if err != nil {
				return err
			}
			deps = append(deps, models.TaskDependency{PredecessorTaskID: from, SuccessorTaskID: to, Type: l.typ, LagDays: l.lag})
		}
		if err := createAll(tx, deps); err != nil {
			return wrap("создать связи", err)
		}

		checklist := make([]models.ChecklistItem, 0, len(spec.checklist))
		nextOrder := make(map[uuid.UUID]int)
		for _, item := range spec.checklist {
			id, err := taskID(item.task)
			if err != nil {
				return err
			}
			checklist = append(checklist, models.ChecklistItem{TaskID: id, Text: item.text, IsDone: item.done, OrderIndex: nextOrder[id]})
			nextOrder[id]++
		}
		if err := createAll(tx, checklist); err != nil {
			return wrap("создать чек-листы", err)
		}

		// План заведён в первый день проекта: запись о создании — первая в
		// журнале каждой задачи.
		history := make([]models.TaskHistoryEntry, 0, len(tasks)+len(spec.comments)+len(spec.events))
		for i, task := range tasks {
			history = append(history, historyEntry(c.at(when(spec.start, 2, i)), task.ID, ownerID,
				models.HistoryActionCreated, "", "", ""))
		}

		comments := make([]models.TaskComment, 0, len(spec.comments))
		for _, cm := range spec.comments {
			id, err := taskID(cm.task)
			if err != nil {
				return err
			}
			author, err := userID(cm.author)
			if err != nil {
				return err
			}
			at := c.at(cm.at)
			comments = append(comments, models.TaskComment{Model: models.Model{CreatedAt: at, UpdatedAt: at},
				TaskID: id, AuthorID: author, Text: cm.text})
			history = append(history, historyEntry(at, id, author, models.HistoryActionCommented, "", "", ""))
		}
		if err := createAll(tx, comments); err != nil {
			return wrap("создать комментарии", err)
		}

		for _, e := range spec.events {
			id, err := taskID(e.task)
			if err != nil {
				return err
			}
			actor, err := userID(e.actor)
			if err != nil {
				return err
			}
			history = append(history, historyEntry(c.at(e.at), id, actor, e.action, e.field, e.oldValue, e.newValue))
		}
		if err := createAll(tx, history); err != nil {
			return wrap("создать журнал задач", err)
		}

		log.Printf("создан проект витрины %q: задач %d, связей %d, контрольных точек %d",
			project.Code, len(tasks), len(deps), len(milestones))
		return nil
	})
}

// createAll вставляет срез одним запросом; пустой срез gorm отклонил бы ошибкой.
// Сгенерированные ID попадают в элементы переданного среза.
func createAll[T any](tx *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return tx.Create(&rows).Error
}

func historyEntry(at time.Time, taskID, actorID uuid.UUID, action, field, oldValue, newValue string) models.TaskHistoryEntry {
	optional := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	return models.TaskHistoryEntry{
		Model:    models.Model{CreatedAt: at, UpdatedAt: at},
		TaskID:   taskID,
		ActorID:  actorID,
		Action:   action,
		Field:    optional(field),
		OldValue: optional(oldValue),
		NewValue: optional(newValue),
		Source:   models.HistorySourceApp,
	}
}

// ─── Проекты витрины ─────────────────────────────────────────────────────────

// showcaseCabinet — основной проект витрины, календарь 5/2.
//
// Что на нём видно:
//   - канбан: «План» (впереди, заблокированные, просроченная веха), «В работе»
//     (в работе и просроченная с прогрессом), «Завершены»;
//   - Гант: пять КТ-групп и задачи вне вех, четыре задачи-вехи, критический
//     путь, отстающий прототип на нём (предупреждение над диаграммой), связи
//     FS/SS/FF, лаг, веха с тремя предшественниками, задача без ответственного;
//   - обзор: проект «в зоне риска» (резерв — один рабочий день), просрочки,
//     задачи с отставанием, перегруженные участники в текущем спринте, КТ с
//     сорванной датой;
//   - роли: Кузнецов только смотрит, Соколов работает на полставки (20 ч/нед).
func showcaseCabinet() showcaseProject {
	c := newPlanClock(models.Calendar52)
	const (
		done       = models.TaskStatusDone
		inProgress = models.TaskStatusInProgress
		planned    = models.TaskStatusPlanned
	)

	return showcaseProject{
		clock:       c,
		code:        showcaseCabinetCode,
		name:        "Личный кабинет абитуриента",
		description: "Онлайн-подача документов в приёмную комиссию ТПУ: анкета, загрузка сканов, статус заявления и обмен с 1С:Университет.",
		customer:    "Приёмная комиссия ТПУ",
		phase:       "Фаза 2: Разработка MVP",
		status:      models.ProjectStatusActive,
		// Прогноз по CPM — день сдачи (+25), дедлайн на рабочий день позже:
		// резерв в пределах порога, и обзор показывает «в зоне риска».
		highlightCriticalPath: true,
		start:                 -25,
		deadline:              26,
		owner:                 emailVolkov,
		members: []showcaseMember{
			{emailVolkov, models.ProjectRoleProjectManager, "Руководитель проекта", models.AccessLevelFull, 40},
			{emailOrlova, models.ProjectRoleAnalyst, "Аналитик", models.AccessLevelEdit, 40},
			{emailPetrov, models.ProjectRoleDeveloper, "Backend-разработчик", models.AccessLevelEdit, 40},
			{emailSmirnova, models.ProjectRoleDeveloper, "Frontend-разработчик", models.AccessLevelEdit, 40},
			{emailSokolov, models.ProjectRoleOther, "Дизайнер (полставки)", models.AccessLevelEdit, 20},
			{emailIvanova, models.ProjectRoleOther, "QA-инженер", models.AccessLevelEdit, 40},
			{emailKuznetsov, models.ProjectRoleCurator, "Куратор от заказчика", models.AccessLevelView, 8},
		},
		sprints: []showcaseSprint{
			{"Спринт 1 · Исследование", "0.1", -25, -16},
			{"Спринт 2 · Дизайн и прототип", "0.2", -15, -6},
			{"Спринт 3 · MVP", "0.5", -5, 7},
			{"Спринт 4 · Опытная эксплуатация", "1.0", 8, 26},
		},
		milestones: []showcaseMilestone{
			{"КТ-1", "Требования утверждены", -16, fact(-16)},
			{"КТ-2", "Прототип утверждён", -2, nil},
			{"КТ-3", "Готовность MVP", 7, nil},
			{"КТ-4", "Готовность к опытной эксплуатации", 17, nil},
			{"КТ-5", "Сдача проекта", 25, nil},
		},
		tasks: []showcaseTask{
			// КТ-1 — закрыта вовремя.
			{key: "interviews", title: "Интервью с приёмной комиссией", assignee: emailOrlova, milestone: "КТ-1", sprint: 0,
				status: done, progress: 100, start: -25, end: -22, factStart: fact(-25), factEnd: fact(-22), weight: 4},
			{key: "analysis", title: "Анализ требований к кабинету", assignee: emailOrlova, milestone: "КТ-1", sprint: 0,
				status: done, progress: 100, start: -21, end: -17, factStart: fact(-21), factEnd: fact(-18), weight: 6},
			{key: "requirements", title: "Требования утверждены заказчиком", assignee: emailVolkov, milestone: "КТ-1", sprint: 0,
				milestoneTask: true, status: done, progress: 100, start: -16, factStart: fact(-16), factEnd: fact(-16)},

			// КТ-2 — прототип отстаёт: он на критическом пути и просрочен.
			{key: "design", title: "Дизайн-макеты личного кабинета", assignee: emailSokolov, milestone: "КТ-2", sprint: 1,
				status: done, progress: 100, start: -15, end: -9, factStart: fact(-15), factEnd: fact(-9), weight: 8},
			{key: "prototype", title: "Прототип интерфейса", assignee: emailSmirnova, milestone: "КТ-2", sprint: 1,
				status: inProgress, progress: 70, start: -12, end: -3, factStart: fact(-12), weight: 8,
				description: "Кликабельный прототип кабинета на реальных макетах.\n\n" +
					"**Экраны:** вход, анкета абитуриента, загрузка документов.\n\n" +
					"Демонстрация заказчику — сразу после готовности загрузки документов."},
			{key: "demo", title: "Демонстрация прототипа заказчику", assignee: emailVolkov, milestone: "КТ-2", sprint: 2,
				milestoneTask: true, status: planned, start: -2},

			// КТ-3 — текущий спринт.
			{key: "api", title: "API личного кабинета", assignee: emailPetrov, milestone: "КТ-3", sprint: 1,
				status: inProgress, progress: 55, start: -14, end: 3, factStart: fact(-14), weight: 12,
				description: "REST API кабинета: профиль абитуриента, документы, статус заявления.\n\n" +
					"- контракт описан в OpenAPI;\n- статус заявления нужен интеграции с 1С.",
				external: &showcaseExternal{provider: "GitLab", url: "https://gitlab.example.com/tpu/abit-cabinet/-/issues/142", referenceID: "142"}},
			{key: "esia", title: "Вход через Госуслуги (ЕСИА)", assignee: emailPetrov, milestone: "КТ-3", sprint: 2,
				status: done, progress: 100, start: -5, end: -2, factStart: fact(-5), factEnd: fact(-1), weight: 5},
			{key: "onec", title: "Интеграция с 1С:Университет", assignee: emailPetrov, milestone: "КТ-3", sprint: 2,
				status: planned, start: 1, end: 6, weight: 8},
			{key: "applyForm", title: "Экран подачи заявления", assignee: emailSmirnova, milestone: "КТ-3", sprint: 2,
				status: planned, start: -2, end: 6, weight: 8},
			{key: "uikit", title: "UI-кит мобильной версии", assignee: emailSokolov, milestone: "КТ-3", sprint: 2,
				status: inProgress, progress: 25, start: -1, end: 5, factStart: fact(-1), weight: 5},
			{key: "notifications", title: "Тексты уведомлений для абитуриентов", assignee: emailOrlova, milestone: "КТ-3", sprint: 2,
				status: planned, start: 2, end: 4, weight: 2},
			{key: "mvp", title: "MVP готов к приёмке", assignee: emailVolkov, milestone: "КТ-3", sprint: 2,
				milestoneTask: true, status: planned, start: 7},

			// КТ-4 и КТ-5 — впереди.
			{key: "loadTest", title: "Нагрузочное тестирование", assignee: emailIvanova, milestone: "КТ-4", sprint: 3,
				status: planned, start: 8, end: 12, weight: 6,
				description: "Пиковая нагрузка приёмной кампании: **1000 одновременных подач** за 10 минут до закрытия приёма."},
			{key: "bugfix", title: "Исправление дефектов", assignee: emailPetrov, milestone: "КТ-4", sprint: 3,
				status: planned, start: 13, end: 17, weight: 6},
			{key: "manual", title: "Руководство пользователя", assignee: emailOrlova, milestone: "КТ-4", sprint: 3,
				status: planned, start: 12, end: 17, weight: 4},
			{key: "pilot", title: "Опытная эксплуатация", assignee: emailVolkov, milestone: "КТ-5", sprint: 3,
				status: planned, start: 18, end: 24, weight: 10,
				description: "Две недели работы с настоящими абитуриентами одного института, ежедневный разбор обращений."},
			{key: "handover", title: "Проект сдан заказчику", assignee: emailVolkov, milestone: "КТ-5", sprint: 3,
				milestoneTask: true, status: planned, start: 25},

			// Вне контрольных точек.
			{key: "cicd", title: "Настройка CI/CD", assignee: emailPetrov, sprint: 0,
				status: done, progress: 100, start: -20, end: -17, factStart: fact(-20), factEnd: fact(-16), weight: 3},
			{key: "security", title: "Аудит безопасности", assignee: emailIvanova, sprint: 3,
				status: planned, start: 9, end: 11, weight: 3},
			{key: "support", title: "Регламент техподдержки", sprint: 3,
				status: planned, start: 19, end: 22, weight: 2},
		},
		links: []showcaseLink{
			{"interviews", "analysis", models.DependencyFS, 0},
			{"analysis", "requirements", models.DependencyFS, 0},
			{"requirements", "design", models.DependencyFS, 0},
			{"requirements", "api", models.DependencyFS, 0},
			{"requirements", "esia", models.DependencyFS, 0},
			// Прототип стартует через три дня после начала макетов — по первым экранам.
			{"design", "prototype", models.DependencySS, 3},
			{"design", "demo", models.DependencyFS, 0},
			{"prototype", "demo", models.DependencyFS, 0},
			{"prototype", "applyForm", models.DependencyFS, 0},
			{"design", "uikit", models.DependencyFS, 0},
			// Интеграцию начинают, когда у API готов контракт — через неделю после старта.
			{"api", "onec", models.DependencySS, 5},
			{"onec", "mvp", models.DependencyFS, 0},
			{"applyForm", "mvp", models.DependencyFS, 0},
			{"uikit", "mvp", models.DependencyFS, 0},
			{"mvp", "loadTest", models.DependencyFS, 0},
			{"loadTest", "bugfix", models.DependencyFS, 0},
			{"mvp", "manual", models.DependencyFS, 0},
			// Руководство дописывают к последним исправлениям — финиш вместе.
			{"bugfix", "manual", models.DependencyFF, 0},
			{"bugfix", "pilot", models.DependencyFS, 0},
			{"manual", "pilot", models.DependencyFS, 0},
			{"pilot", "handover", models.DependencyFS, 0},
			// Регламент — через день после руководства: нужен его вычитанный текст.
			{"manual", "support", models.DependencyFS, 1},
		},
		checklist: []showcaseCheck{
			{"analysis", "Пользовательские сценарии абитуриента", true},
			{"analysis", "Нефункциональные требования", true},
			{"prototype", "Экран входа", true},
			{"prototype", "Анкета абитуриента", true},
			{"prototype", "Загрузка документов", false},
			{"api", "Эндпоинты профиля абитуриента", true},
			{"api", "Загрузка документов (PDF, JPG)", true},
			{"api", "Статусы заявления", false},
			{"api", "Описание в OpenAPI", false},
			{"loadTest", "Сценарий 1000 одновременных подач", false},
			{"loadTest", "Отчёт по времени отклика", false},
		},
		comments: []showcaseComment{
			{"api", emailVolkov, "Для 1С нужен статус заявления в API — учти это в контракте.", when(-6, 9, 0)},
			{"api", emailPetrov, "Добавил поле status в OpenAPI, заглушку уже отдаю.", when(-5, 3, 30)},
			{"demo", emailVolkov, "Заказчик просит показать прототип на планёрке приёмной комиссии. Переносим демонстрацию, как только закроем загрузку документов.", when(-4, 5, 0)},
			{"prototype", emailVolkov, "Прототип отстаёт от плана — что мешает?", when(-3, 3, 15)},
			{"prototype", emailSmirnova, "Ждали финальные иконки. Вход и анкета готовы, осталась загрузка документов.", when(-3, 4, 2)},
			{"prototype", emailSokolov, "Иконки выложил в Figma, раздел «Кабинет → Документы».", when(-2, 2, 40)},
			{"onec", emailOrlova, "Запросила доступ к тестовому контуру 1С — обещают выдать до начала задачи.", when(-1, 7, 20)},
		},
		events: []showcaseEvent{
			statusEvent("interviews", emailOrlova, planned, inProgress, when(-25, 3, 0)),
			statusEvent("interviews", emailOrlova, inProgress, done, when(-22, 10, 30)),
			statusEvent("cicd", emailPetrov, planned, inProgress, when(-20, 4, 0)),
			statusEvent("analysis", emailOrlova, planned, inProgress, when(-21, 2, 30)),
			progressEvent("analysis", emailOrlova, 0, 60, when(-19, 11, 0)),
			statusEvent("analysis", emailOrlova, inProgress, done, when(-18, 11, 0)),
			statusEvent("requirements", emailVolkov, planned, done, when(-16, 8, 0)),
			statusEvent("cicd", emailPetrov, inProgress, done, when(-16, 12, 0)),
			statusEvent("design", emailSokolov, planned, inProgress, when(-15, 3, 0)),
			statusEvent("api", emailPetrov, planned, inProgress, when(-14, 3, 0)),
			statusEvent("prototype", emailSmirnova, planned, inProgress, when(-12, 2, 45)),
			statusEvent("design", emailSokolov, inProgress, done, when(-9, 10, 0)),
			{task: "api", actor: emailPetrov, action: models.HistoryActionUpdated, field: "checklist",
				newValue: "Эндпоинты профиля абитуриента: выполнен", at: when(-9, 9, 0)},
			progressEvent("api", emailPetrov, 0, 30, when(-8, 11, 0)),
			progressEvent("prototype", emailSmirnova, 0, 40, when(-7, 11, 0)),
			{task: "prototype", actor: emailVolkov, action: models.HistoryActionDateShifted, field: "endDate",
				oldValue: c.day(-5).String(), newValue: c.day(-3).String(), at: when(-6, 5, 0)},
			statusEvent("esia", emailPetrov, planned, inProgress, when(-5, 2, 0)),
			{task: "api", actor: emailPetrov, action: models.HistoryActionUpdated, field: "checklist",
				newValue: "Загрузка документов (PDF, JPG): выполнен", at: when(-4, 9, 0)},
			progressEvent("prototype", emailSmirnova, 40, 70, when(-3, 10, 10)),
			progressEvent("api", emailPetrov, 30, 55, when(-2, 11, 30)),
			statusEvent("esia", emailPetrov, inProgress, done, when(-1, 9, 0)),
			statusEvent("uikit", emailSokolov, planned, inProgress, when(-1, 4, 0)),
			progressEvent("uikit", emailSokolov, 0, 25, when(-1, 10, 0)),
		},
	}
}

// showcaseScience — второй проект: календарь 6/1, подсветка критического пути
// выключена, план идёт с запасом (обзор — «в графике»). Здесь Смирнова —
// руководитель, а Волков и Кузнецов — участники: роль и доступ у человека свои
// в каждом проекте.
func showcaseScience() showcaseProject {
	c := newPlanClock(models.Calendar61)
	const (
		done       = models.TaskStatusDone
		inProgress = models.TaskStatusInProgress
		planned    = models.TaskStatusPlanned
	)

	return showcaseProject{
		clock:                 c,
		code:                  showcaseScienceCode,
		name:                  "Портал научных публикаций",
		description:           "Единый каталог публикаций сотрудников ТПУ с импортом из eLibrary и поиском по индексации.",
		customer:              "Научно-техническая библиотека ТПУ",
		phase:                 "Фаза 1: Проектирование и бета-версия",
		status:                models.ProjectStatusActive,
		highlightCriticalPath: false,
		start:                 -10,
		deadline:              30,
		owner:                 emailSmirnova,
		members: []showcaseMember{
			{emailSmirnova, models.ProjectRoleProjectManager, "Руководитель проекта", models.AccessLevelFull, 40},
			{emailVolkov, models.ProjectRoleDeveloper, "Архитектор", models.AccessLevelEdit, 16},
			{emailKuznetsov, models.ProjectRoleAnalyst, "Методист библиотеки", models.AccessLevelEdit, 20},
			{emailIvanova, models.ProjectRoleOther, "Наблюдатель от QA", models.AccessLevelView, 8},
		},
		sprints: []showcaseSprint{
			{"Спринт 1 · Проектирование", "0.1", -10, 3},
			{"Спринт 2 · Бета-версия", "0.9", 4, 13},
		},
		milestones: []showcaseMilestone{
			{"КТ-1", "Бета-версия портала", 12, nil},
		},
		tasks: []showcaseTask{
			{key: "requirements", title: "Сбор требований к порталу", assignee: emailKuznetsov, sprint: 0,
				status: done, progress: 100, start: -10, end: -6, factStart: fact(-10), factEnd: fact(-6), weight: 15},
			{key: "dataModel", title: "Модель данных публикаций", assignee: emailVolkov, sprint: 0,
				status: inProgress, progress: 50, start: -5, end: 2, factStart: fact(-5), weight: 20},
			{key: "catalog", title: "Макет каталога", assignee: emailSmirnova, sprint: 0,
				status: inProgress, progress: 60, start: -4, end: 3, factStart: fact(-4), weight: 20},
			{key: "agreement", title: "Согласование с библиотекой ТПУ", assignee: emailKuznetsov, sprint: 0,
				status: planned, start: 0, end: 2, weight: 10},
			{key: "import", title: "Импорт из eLibrary", assignee: emailVolkov, milestone: "КТ-1", sprint: 1,
				status: planned, start: 3, end: 9, weight: 15},
			{key: "search", title: "Поиск по каталогу", assignee: emailSmirnova, milestone: "КТ-1", sprint: 1,
				status: planned, start: 4, end: 10, weight: 20},
			{key: "beta", title: "Бета-версия открыта библиотеке", assignee: emailSmirnova, milestone: "КТ-1", sprint: 1,
				milestoneTask: true, status: planned, start: 12},
		},
		links: []showcaseLink{
			{"requirements", "dataModel", models.DependencyFS, 0},
			{"requirements", "catalog", models.DependencyFS, 0},
			{"dataModel", "import", models.DependencyFS, 0},
			{"catalog", "search", models.DependencyFS, 0},
			{"dataModel", "search", models.DependencySS, 2},
			{"import", "beta", models.DependencyFS, 0},
			{"search", "beta", models.DependencyFS, 0},
		},
		checklist: []showcaseCheck{
			{"dataModel", "Сущности: публикация, автор, журнал", true},
			{"dataModel", "Связь автора с ORCID", false},
		},
		comments: []showcaseComment{
			{"catalog", emailKuznetsov, "Библиотека просит фильтр по индексации в Scopus и WoS.", when(-2, 6, 0)},
			{"catalog", emailSmirnova, "Добавила фильтр в макет — посмотрите вкладку «Каталог».", when(-2, 7, 30)},
		},
		events: []showcaseEvent{
			statusEvent("requirements", emailKuznetsov, planned, inProgress, when(-10, 3, 0)),
			statusEvent("requirements", emailKuznetsov, inProgress, done, when(-6, 10, 0)),
			statusEvent("dataModel", emailVolkov, planned, inProgress, when(-5, 4, 0)),
			statusEvent("catalog", emailSmirnova, planned, inProgress, when(-4, 3, 0)),
			progressEvent("dataModel", emailVolkov, 0, 50, when(-1, 9, 0)),
			progressEvent("catalog", emailSmirnova, 0, 60, when(-1, 11, 0)),
		},
	}
}

// showcaseChatbot — черновик без задач: пустое состояние Ганта и канбана и
// третий пункт в переключателе проектов.
func showcaseChatbot() showcaseProject {
	return showcaseProject{
		clock:                 newPlanClock(models.Calendar52),
		code:                  showcaseChatbotCode,
		name:                  "Чат-бот приёмной кампании",
		description:           "Ответы абитуриентам на частые вопросы в мессенджерах. План ещё не составлен.",
		customer:              "Приёмная комиссия ТПУ",
		status:                models.ProjectStatusDraft,
		highlightCriticalPath: true,
		start:                 10,
		deadline:              50,
		owner:                 emailVolkov,
		members: []showcaseMember{
			{emailVolkov, models.ProjectRoleProjectManager, "Руководитель проекта", models.AccessLevelFull, 40},
			{emailOrlova, models.ProjectRoleAnalyst, "Аналитик", models.AccessLevelEdit, 40},
		},
	}
}

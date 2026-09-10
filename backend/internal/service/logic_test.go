package service

// Тесты чистой логики сервисов — без базы данных, выполняются за миллисекунды.
// Сценарии целиком (через HTTP и PostgreSQL) проверяет internal/api.

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"taygant_backend/internal/models"
)

func edge(from, to uuid.UUID) models.TaskDependency {
	return models.TaskDependency{PredecessorTaskID: from, SuccessorTaskID: to}
}

func TestWouldCreateCycle(t *testing.T) {
	a, b, c, d := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	// a → b → c и отдельная ветка a → d.
	edges := []models.TaskDependency{edge(a, b), edge(b, c), edge(a, d)}

	cases := []struct {
		name     string
		from, to uuid.UUID
		want     bool
	}{
		{"прямая обратная связь", b, a, true},
		{"замыкание через цепочку", c, a, true},
		{"связь вперёд по цепочке", a, c, false},
		{"между параллельными ветками", d, c, false},
		{"из конца цепочки в боковую ветку", c, d, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := wouldCreateCycle(edges, tc.from, tc.to); got != tc.want {
				t.Fatalf("wouldCreateCycle = %v, ожидалось %v", got, tc.want)
			}
		})
	}
}

func TestWorkingDaysBetween(t *testing.T) {
	monday := models.NewDate(2026, time.September, 7)
	if monday.Weekday() != time.Monday {
		t.Fatalf("опорная дата %s — не понедельник", monday)
	}
	saturday := monday.AddDays(5)

	cases := []struct {
		name       string
		start, end models.Date
		calendar   models.WorkingCalendarType
		want       int
	}{
		{"неделя при 5/2", monday, monday.AddDays(6), models.Calendar52, 5},
		{"две недели при 5/2", monday, monday.AddDays(13), models.Calendar52, 10},
		{"один будний день", monday, monday, models.Calendar52, 1},
		{"суббота при 5/2 выходная", saturday, saturday, models.Calendar52, 0},
		{"конец раньше начала", monday.AddDays(3), monday, models.Calendar52, 0},
		// Шестидневка: суббота рабочая, выходной только воскресенье. Сейчас код
		// исключает субботу при любом календаре, и 6/1 считается как 5/2.
		{"неделя при 6/1", monday, monday.AddDays(6), models.Calendar61, 6},
		{"суббота при 6/1 рабочая", saturday, saturday, models.Calendar61, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := workingDaysBetween(tc.start, tc.end, tc.calendar); got != tc.want {
				t.Fatalf("workingDaysBetween(%s, %s, %s) = %d, ожидалось %d", tc.start, tc.end, tc.calendar, got, tc.want)
			}
		})
	}
}

func TestResolveStatuses(t *testing.T) {
	today := models.Today()
	past, future := today.AddDays(-1), today.AddDays(5)
	task := func(status models.TaskStatus, end models.Date) models.Task {
		return models.Task{Model: models.Model{ID: uuid.New()}, Status: status, EndDate: end}
	}

	unfinished := task(models.TaskStatusInProgress, future)
	finished := task(models.TaskStatusDone, past)
	late := task(models.TaskStatusPlanned, past)
	waiting := task(models.TaskStatusPlanned, future)
	lateAndWaiting := task(models.TaskStatusInProgress, past)
	free := task(models.TaskStatusPlanned, future)
	// Предшественник есть в карте статусов, но не в списке задач — так
	// вызывает resolveStatuses Get для одной задачи.
	outside := task(models.TaskStatusDone, future)
	freeOfOutside := task(models.TaskStatusPlanned, future)

	tasks := []models.Task{unfinished, finished, late, waiting, lateAndWaiting, free, freeOfOutside}
	statuses := statusMap(append([]models.Task{outside}, tasks...))
	deps := []models.TaskDependency{
		edge(unfinished.ID, waiting.ID),
		edge(unfinished.ID, lateAndWaiting.ID),
		edge(finished.ID, free.ID),
		edge(outside.ID, freeOfOutside.ID),
	}

	resolveStatuses(tasks, statuses, deps)

	want := []models.TaskStatus{
		models.TaskStatusInProgress, // в срок и без предшественников — не трогаем
		models.TaskStatusDone,       // done финальный, даже если срок прошёл
		models.TaskStatusOverdue,
		models.TaskStatusBlocked,
		models.TaskStatusOverdue, // просрочка важнее блокировки
		models.TaskStatusPlanned, // предшественник завершён
		models.TaskStatusPlanned, // предшественник вне списка, но завершён
	}
	for i := range tasks {
		if tasks[i].Status != want[i] {
			t.Errorf("задача %d: статус %q, ожидался %q", i, tasks[i].Status, want[i])
		}
	}
}

func TestDeriveMilestoneStatuses(t *testing.T) {
	today := models.Today()
	reached := today.AddDays(-20)
	milestones := []models.Milestone{
		{Code: "final", PlannedDate: today.AddDays(30)},
		{Code: "done", PlannedDate: today.AddDays(-10), ActualDate: &reached},
		{Code: "current", PlannedDate: today.AddDays(-2)},
		{Code: "planned", PlannedDate: today.AddDays(7)},
	}

	deriveMilestoneStatuses(milestones)

	for _, m := range milestones {
		if string(m.Status) != m.Code {
			t.Errorf("веха %q получила статус %q", m.Code, m.Status)
		}
	}
	// Риск считается только для недостигнутой вехи с прошедшей плановой датой.
	for _, m := range milestones {
		switch {
		case m.Code == "current" && m.RiskDays == nil:
			t.Error("просроченная на 2 дня веха: riskDays = nil, ожидалось 2")
		case m.Code == "current" && *m.RiskDays != 2:
			t.Errorf("просроченная на 2 дня веха: riskDays = %d, ожидалось 2", *m.RiskDays)
		case m.Code != "current" && m.RiskDays != nil:
			t.Errorf("веха %q: riskDays = %d, ожидался nil", m.Code, *m.RiskDays)
		}
	}
}

// Обратный слеш — символ экранирования в LIKE. Замена `\` → `\` ничего не
// меняет: введённый пользователем "\" портит шаблон (поиск "\" ищет "%").
func TestEscapeLikePattern(t *testing.T) {
	cases := map[string]string{
		"обычный текст": "обычный текст",
		"100%":          `100\%`,
		"snake_case":    `snake\_case`,
		`C:\temp`:       `C:\\temp`,
	}
	for in, want := range cases {
		if got := escapeLikePattern(in); got != want {
			t.Errorf("escapeLikePattern(%q) = %q, ожидалось %q", in, got, want)
		}
	}
}

// dummyHash выравнивает время ответа для несуществующего email. Это работает,
// только если строка — настоящий bcrypt-хеш той же стоимости, что и у реальных
// паролей: битый хеш отбрасывается мгновенно и выдаёт себя по времени.
func TestDummyHashCostsAsMuchAsRealOne(t *testing.T) {
	cost, err := bcrypt.Cost([]byte(dummyHash))
	if err != nil {
		t.Fatalf("dummyHash не разбирается как bcrypt: %v", err)
	}
	if cost != bcrypt.DefaultCost {
		t.Fatalf("стоимость dummyHash = %d, у реальных паролей %d", cost, bcrypt.DefaultCost)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte("anything")); !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		t.Fatalf("сравнение с dummyHash завершилось %v, ожидалось полноценное несовпадение", err)
	}
}

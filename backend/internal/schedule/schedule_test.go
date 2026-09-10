package schedule

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"taygant_backend/internal/models"
)

func d(y int, m time.Month, day int) models.Date { return models.NewDate(y, m, day) }

func newID(t *testing.T) uuid.UUID {
	t.Helper()
	return uuid.New()
}

// Двухзадачная цепочка FS без зазора: если B начинается ровно на следующий
// день после конца A, обе задачи критические, резерва нет ни у одной.
func TestComputeTightFSChainBothCritical(t *testing.T) {
	a, b := newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 6), End: d(2026, 1, 10)},
	}
	deps := []Dependency{{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS}}

	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	summary := g.Compute()

	if !summary.Tasks[a].Critical || !summary.Tasks[b].Critical {
		t.Fatalf("ожидались обе критические, получено a=%+v b=%+v", summary.Tasks[a], summary.Tasks[b])
	}
	if summary.Tasks[a].SlackDays != 0 || summary.Tasks[b].SlackDays != 0 {
		t.Fatalf("ожидался нулевой резерв, получено a=%d b=%d", summary.Tasks[a].SlackDays, summary.Tasks[b].SlackDays)
	}
	if !summary.ComputedFinish.Equal(d(2026, 1, 10)) {
		t.Fatalf("ComputedFinish = %s, ожидалось 2026-01-10", summary.ComputedFinish)
	}
}

// Та же цепочка, но B запланирована с четырёхдневным запасом после A:
// запас принадлежит A (она может сдвинуться на эти 4 дня, не потревожив B),
// а B остаётся критической — она сама определяет финиш ветки графа.
func TestComputeBufferBelongsToUpstreamTask(t *testing.T) {
	a, b := newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},   // предшественник заканчивается 5-го
		{ID: b, Start: d(2026, 1, 10), End: d(2026, 1, 14)}, // а последователь начинается только 10-го — 4 дня запаса
	}
	deps := []Dependency{{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS}}

	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	summary := g.Compute()

	if summary.Tasks[a].SlackDays != 4 {
		t.Fatalf("SlackDays(a) = %d, ожидалось 4", summary.Tasks[a].SlackDays)
	}
	if summary.Tasks[a].Critical {
		t.Fatal("a не должна быть критической — у неё есть резерв")
	}
	if !summary.Tasks[b].Critical || summary.Tasks[b].SlackDays != 0 {
		t.Fatalf("b должна быть критической с нулевым резервом, получено %+v", summary.Tasks[b])
	}
}

// Ромб: A -> B -> D и A -> C -> D. Ветка через C длиннее (без запаса),
// ветка через B короче (есть запас) — критическим должен быть только длинный путь.
func TestComputeDiamondOnlyLongestPathCritical(t *testing.T) {
	a, b, c, dd := newID(t), newID(t), newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 6), End: d(2026, 1, 8)},  // короткая ветка
		{ID: c, Start: d(2026, 1, 6), End: d(2026, 1, 12)}, // длинная ветка
		{ID: dd, Start: d(2026, 1, 13), End: d(2026, 1, 15)},
	}
	deps := []Dependency{
		{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS},
		{PredecessorID: a, SuccessorID: c, Type: models.DependencyFS},
		{PredecessorID: b, SuccessorID: dd, Type: models.DependencyFS},
		{PredecessorID: c, SuccessorID: dd, Type: models.DependencyFS},
	}

	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	summary := g.Compute()

	for _, id := range []uuid.UUID{a, c, dd} {
		if !summary.Tasks[id].Critical {
			t.Errorf("ожидалась критическая задача %s, резерв=%d", id, summary.Tasks[id].SlackDays)
		}
	}
	if summary.Tasks[b].Critical {
		t.Errorf("b не должна быть критической, резерв=%d", summary.Tasks[b].SlackDays)
	}
	// A->C(6 дней) против A->B(2 дня): у B должно быть 4 дня резерва.
	if summary.Tasks[b].SlackDays != 4 {
		t.Errorf("SlackDays(b) = %d, ожидалось 4", summary.Tasks[b].SlackDays)
	}
}

// ComputedFinish — это собственный финиш графа (самый длинный путь по сети),
// а не дедлайн проекта: пакету schedule дедлайн вообще не передаётся.
// Единственная задача графа всегда критическая — она сама определяет финиш.
func TestComputeFinishIsGraphsOwn(t *testing.T) {
	a := newID(t)
	tasks := []Task{{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 20)}}

	g, err := NewGraph(tasks, nil)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}
	summary := g.Compute()

	if !summary.ComputedFinish.Equal(d(2026, 1, 20)) {
		t.Fatalf("ComputedFinish = %s, ожидалось 2026-01-20", summary.ComputedFinish)
	}
	if !summary.Tasks[a].Critical {
		t.Fatal("единственная задача графа обязана быть критической")
	}
}

// Сравнение ComputedFinish с дедлайном (сорван ли срок и на сколько) — забота
// вызывающего кода (дашборда), не этого теста: пакет schedule дедлайна не видит.

// Цикл в графе — ошибка, а не штатный сценарий (сервис связей его не
// допускает), но движок обязан её ловить, а не зависать или молча всё пропустить.
func TestNewGraphDetectsCycle(t *testing.T) {
	a, b := newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 6), End: d(2026, 1, 10)},
	}
	deps := []Dependency{
		{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS},
		{PredecessorID: b, SuccessorID: a, Type: models.DependencyFS},
	}
	if _, err := NewGraph(tasks, deps); err == nil {
		t.Fatal("ожидалась ошибка цикла")
	}
}

// SS/FF/SF должны сдвигать требуемый старт последователя иначе, чем FS —
// проверяем формулы напрямую через двух-задачный граф на каждый тип.
func TestComputeDependencyTypes(t *testing.T) {
	cases := []struct {
		name          string
		depType       models.DependencyType
		predStart     models.Date
		predEnd       models.Date
		succDuration  int // задаём длительность последователя через его собственные даты ниже
		wantSuccStart models.Date
	}{
		{
			name:    "SS: последователь не раньше старта предшественника",
			depType: models.DependencySS, predStart: d(2026, 1, 1), predEnd: d(2026, 1, 10),
			succDuration: 3, wantSuccStart: d(2026, 1, 1),
		},
		{
			name:    "FF: последователь не может закончиться раньше предшественника",
			depType: models.DependencyFF, predStart: d(2026, 1, 1), predEnd: d(2026, 1, 10),
			succDuration: 3, wantSuccStart: d(2026, 1, 7), // 10 - 3
		},
		{
			name:    "SF: последователь не может закончиться раньше старта предшественника",
			depType: models.DependencySF, predStart: d(2026, 1, 5), predEnd: d(2026, 1, 10),
			succDuration: 3, wantSuccStart: d(2026, 1, 2), // 5 - 3
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pred, succ := newID(t), newID(t)
			tasks := []Task{
				{ID: pred, Start: tc.predStart, End: tc.predEnd},
				// Даты последователя ниже целевой — форвард обязан их подвинуть.
				{ID: succ, Start: d(2020, 1, 1), End: d(2020, 1, 1).AddDays(tc.succDuration)},
			}
			deps := []Dependency{{PredecessorID: pred, SuccessorID: succ, Type: tc.depType}}
			g, err := NewGraph(tasks, deps)
			if err != nil {
				t.Fatalf("NewGraph: %v", err)
			}
			summary := g.Compute()
			got := summary.Tasks[succ].EarlyStart
			if !got.Equal(tc.wantSuccStart) {
				t.Fatalf("EarlyStart(succ) = %s, ожидалось %s", got, tc.wantSuccStart)
			}
		})
	}
}

// Жёсткая FS-цепочка без запаса: сдвиг A на 3 дня должен полностью
// прокатиться до B — запаса, чтобы его поглотить, нет.
func TestShiftCascadesThroughTightChain(t *testing.T) {
	a, b := newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 6), End: d(2026, 1, 10)},
	}
	deps := []Dependency{{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS}}
	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	result, err := g.Shift(a, d(2026, 1, 4)) // A сдвинута на +3 дня
	if err != nil {
		t.Fatalf("Shift: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ожидались обе задачи в результате, получено %d: %+v", len(result), result)
	}
	if !result[0].NewStart.Equal(d(2026, 1, 4)) {
		t.Fatalf("A.NewStart = %s, ожидалось 2026-01-04", result[0].NewStart)
	}
	if !result[1].NewStart.Equal(d(2026, 1, 9)) {
		t.Fatalf("B.NewStart = %s, ожидалось 2026-01-09 (тоже +3 дня)", result[1].NewStart)
	}
}

// У B есть 4 дня запаса перед A. Сдвиг A на 2 дня укладывается в запас — B не
// должна сдвинуться вообще, в результате должна остаться только сама A.
func TestShiftWithinBufferDoesNotCascade(t *testing.T) {
	a, b := newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 10), End: d(2026, 1, 14)}, // 4 дня запаса после A
	}
	deps := []Dependency{{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS}}
	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	result, err := g.Shift(a, d(2026, 1, 3)) // +2 дня, запас 4 дня — не должно каскадировать
	if err != nil {
		t.Fatalf("Shift: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ожидалась только A в результате, получено %d: %+v", len(result), result)
	}
	if result[0].ID != a {
		t.Fatalf("единственный элемент должен быть A, получено %s", result[0].ID)
	}
}

// Сдвиг A на 6 дней при запасе в 4 дня — B должна сдвинуться только на
// оставшиеся 2 дня (превышение запаса), а не на все 6.
func TestShiftBeyondBufferMovesOnlyOverflow(t *testing.T) {
	a, b := newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 10), End: d(2026, 1, 14)},
	}
	deps := []Dependency{{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS}}
	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	result, err := g.Shift(a, d(2026, 1, 7)) // +6 дней
	if err != nil {
		t.Fatalf("Shift: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ожидались обе задачи, получено %d: %+v", len(result), result)
	}
	// B должна сдвинуться с 10-го на 12-е (требование: старт A(7)+4дня+1 = 12-е).
	if !result[1].NewStart.Equal(d(2026, 1, 12)) {
		t.Fatalf("B.NewStart = %s, ожидалось 2026-01-12", result[1].NewStart)
	}
}

// Независимая задача (без связи с источником сдвига) никогда не должна
// появиться в результате Shift.
func TestShiftDoesNotTouchUnrelatedTasks(t *testing.T) {
	a, b, unrelated := newID(t), newID(t), newID(t)
	tasks := []Task{
		{ID: a, Start: d(2026, 1, 1), End: d(2026, 1, 5)},
		{ID: b, Start: d(2026, 1, 6), End: d(2026, 1, 10)},
		{ID: unrelated, Start: d(2026, 2, 1), End: d(2026, 2, 5)},
	}
	deps := []Dependency{{PredecessorID: a, SuccessorID: b, Type: models.DependencyFS}}
	g, err := NewGraph(tasks, deps)
	if err != nil {
		t.Fatalf("NewGraph: %v", err)
	}

	result, err := g.Shift(a, d(2026, 1, 10))
	if err != nil {
		t.Fatalf("Shift: %v", err)
	}
	for _, r := range result {
		if r.ID == unrelated {
			t.Fatal("независимая задача не должна была сдвинуться")
		}
	}
}

// Package schedule реализует метод критического пути (CPM): по графу задач и
// связей между ними считает ранние/поздние даты и резерв (slack) каждой задачи.
//
// Пакет не знает про БД и HTTP — он работает с уже загруженными срезами
// Task/Dependency и ничего не запрашивает сам. Вызывающая сторона (сервис
// задач) отвечает за то, чтобы передать полный граф одного проекта: движок
// не проверяет принадлежность задач проекту, только связность самого графа.
package schedule

import (
	"fmt"

	"github.com/google/uuid"

	"taygant_backend/internal/models"
)

// Task — минимальные данные о задаче, нужные движку: её текущие плановые даты.
// Длительность выводится из них (EndDate - StartDate), отдельно не передаётся.
type Task struct {
	ID    uuid.UUID
	Start models.Date
	End   models.Date
}

func (t Task) duration() int { return t.Start.DaysUntil(t.End) }

// Dependency — ребро графа: PredecessorID должен завершиться (или начаться —
// в зависимости от Type) раньше, чем сможет начаться/закончиться SuccessorID.
type Dependency struct {
	PredecessorID uuid.UUID
	SuccessorID   uuid.UUID
	Type          models.DependencyType
	LagDays       int
}

// dateRange — пара дат начала/конца, общая форма для ранних и поздних значений.
type dateRange struct {
	Start models.Date
	End   models.Date
}

// TaskSchedule — результат CPM для одной задачи.
type TaskSchedule struct {
	EarlyStart, EarlyFinish models.Date
	LateStart, LateFinish   models.Date
	// SlackDays — резерв: на сколько дней можно сдвинуть задачу, не сдвинув
	// эту ветку графа за целевую дату (Summary.ComputedFinish). Ноль или
	// меньше — задача на критическом пути.
	SlackDays int
	Critical  bool
}

// Summary — результат расчёта по всему графу.
type Summary struct {
	Tasks map[uuid.UUID]TaskSchedule
	// ComputedFinish — самая поздняя из ранних дат завершения по графу:
	// естественный срок готовности проекта при текущих длительностях и
	// связях, без оглядки на плановый дедлайн. Если он позже дедлайна,
	// проект технически не укладывается в срок при текущем плане.
	ComputedFinish models.Date
}

// Graph — граф проекта, подготовленный к расчёту (топологически отсортирован,
// проверен на циклы).
type Graph struct {
	tasks map[uuid.UUID]Task
	out   map[uuid.UUID][]Dependency // исходящие рёбра, ключ — PredecessorID
	in    map[uuid.UUID][]Dependency // входящие рёбра, ключ — SuccessorID
	order []uuid.UUID                // топологический порядок (предшественник раньше последователя)
}

// NewGraph строит граф и топологически сортирует его алгоритмом Кана.
//
// Цикл в графе означает ошибку в данных, а не штатный сценарий: сервис связей
// (internal/service/dependencies.go) не даёт создать ребро, замыкающее цикл, —
// эта проверка здесь просто defensive-страховка на случай рассинхронизации.
func NewGraph(tasks []Task, deps []Dependency) (*Graph, error) {
	g := &Graph{
		tasks: make(map[uuid.UUID]Task, len(tasks)),
		out:   make(map[uuid.UUID][]Dependency),
		in:    make(map[uuid.UUID][]Dependency),
	}
	for _, t := range tasks {
		g.tasks[t.ID] = t
	}

	indegree := make(map[uuid.UUID]int, len(tasks))
	for _, t := range tasks {
		indegree[t.ID] = 0
	}
	for _, d := range deps {
		// Рёбра на задачи вне переданного среза (другой проект, удалённая
		// задача) молча пропускаются — граф должен остаться согласованным.
		if _, ok := g.tasks[d.PredecessorID]; !ok {
			continue
		}
		if _, ok := g.tasks[d.SuccessorID]; !ok {
			continue
		}
		g.out[d.PredecessorID] = append(g.out[d.PredecessorID], d)
		g.in[d.SuccessorID] = append(g.in[d.SuccessorID], d)
		indegree[d.SuccessorID]++
	}

	queue := make([]uuid.UUID, 0, len(tasks))
	for _, t := range tasks {
		if indegree[t.ID] == 0 {
			queue = append(queue, t.ID)
		}
	}
	order := make([]uuid.UUID, 0, len(tasks))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, d := range g.out[id] {
			indegree[d.SuccessorID]--
			if indegree[d.SuccessorID] == 0 {
				queue = append(queue, d.SuccessorID)
			}
		}
	}
	if len(order) != len(tasks) {
		return nil, fmt.Errorf("граф зависимостей содержит цикл")
	}
	g.order = order
	return g, nil
}

// Order возвращает топологический порядок задач (предшественник всегда раньше
// последователя). Используется симуляцией сдвига, которой нужен тот же порядок
// обхода, но не полный CPM-расчёт.
func (g *Graph) Order() []uuid.UUID { return g.order }

// Out возвращает исходящие рёбра задачи (её последователей).
func (g *Graph) Out(id uuid.UUID) []Dependency { return g.out[id] }

// In возвращает входящие рёбра задачи (её предшественников).
func (g *Graph) In(id uuid.UUID) []Dependency { return g.in[id] }

// Task возвращает узел графа по идентификатору.
func (g *Graph) Task(id uuid.UUID) (Task, bool) {
	t, ok := g.tasks[id]
	return t, ok
}

// Edges возвращает все рёбра графа одним срезом. Нужен, когда по графу с
// изменёнными датами задач (например, после Shift) требуется построить новый
// граф с тем же набором связей — без этого метода пришлось бы заново идти в БД.
func (g *Graph) Edges() []Dependency {
	edges := make([]Dependency, 0, len(g.tasks))
	for _, list := range g.out {
		edges = append(edges, list...)
	}
	return edges
}

// requiredSuccessorStart — минимальная дата начала последователя, которую
// накладывает одно ребро, при заданных датах предшественника.
//
// Гэп в одни сутки для FS — не случайность: демо-данные и весь остальной код
// считают, что «предшественник закончился 20-го» и «последователь начался
// 21-го» — это лаг 0, а не 1. Здесь та же логика: FS с lag=0 разрешает старт
// на следующий день после конца предшественника, а не в тот же день.
func requiredSuccessorStart(pred dateRange, d Dependency, succDuration int) models.Date {
	switch d.Type {
	case models.DependencyFS:
		return pred.End.AddDays(1 + d.LagDays)
	case models.DependencySS:
		return pred.Start.AddDays(d.LagDays)
	case models.DependencyFF:
		// succ.End >= pred.End + lag  =>  succ.Start >= pred.End + lag - duration
		return pred.End.AddDays(d.LagDays - succDuration)
	case models.DependencySF:
		// succ.End >= pred.Start + lag  =>  succ.Start >= pred.Start + lag - duration
		return pred.Start.AddDays(d.LagDays - succDuration)
	default:
		return pred.End.AddDays(1 + d.LagDays)
	}
}

// maxPredecessorFinish — обратная к requiredSuccessorStart: самая поздняя дата
// завершения предшественника, при которой последователь ещё успевает в свои
// поздние даты succLate.
func maxPredecessorFinish(succLate dateRange, d Dependency, predDuration int) models.Date {
	switch d.Type {
	case models.DependencyFS:
		return succLate.Start.AddDays(-1 - d.LagDays)
	case models.DependencySS:
		return succLate.Start.AddDays(-d.LagDays).AddDays(predDuration)
	case models.DependencyFF:
		return succLate.End.AddDays(-d.LagDays)
	case models.DependencySF:
		return succLate.End.AddDays(-d.LagDays).AddDays(predDuration)
	default:
		return succLate.Start.AddDays(-1 - d.LagDays)
	}
}

// forward — прямой проход: ранние даты. Задача без предшественников держится
// собственной плановой даты начала (а не общего начала проекта) — у разных
// независимых веток графа могут быть разные легитимные стартовые даты.
func (g *Graph) forward() map[uuid.UUID]dateRange {
	result := make(map[uuid.UUID]dateRange, len(g.tasks))
	for _, id := range g.order {
		t := g.tasks[id]
		duration := t.duration()
		start := t.Start
		for _, d := range g.in[id] {
			pred := result[d.PredecessorID]
			required := requiredSuccessorStart(pred, d, duration)
			if required.After(start) {
				start = required
			}
		}
		result[id] = dateRange{Start: start, End: start.AddDays(duration)}
	}
	return result
}

// backward — обратный проход: поздние даты, посчитанные от target (целевой
// даты завершения этой ветки графа) в обратном топологическом порядке.
func (g *Graph) backward(target models.Date) map[uuid.UUID]dateRange {
	result := make(map[uuid.UUID]dateRange, len(g.tasks))
	for i := len(g.order) - 1; i >= 0; i-- {
		id := g.order[i]
		t := g.tasks[id]
		duration := t.duration()

		finish := target
		first := true
		for _, d := range g.out[id] {
			succLate := result[d.SuccessorID]
			allowed := maxPredecessorFinish(succLate, d, duration)
			if first || allowed.Before(finish) {
				finish = allowed
				first = false
			}
		}
		result[id] = dateRange{Start: finish.AddDays(-duration), End: finish}
	}
	return result
}

// ShiftedTask — задача, чьи даты изменились в результате Shift.
type ShiftedTask struct {
	ID                         uuid.UUID
	OriginalStart, OriginalEnd models.Date
	NewStart, NewEnd           models.Date
	// CausedBy — тип связи, из-за которой сдвинулась именно эта задача
	// (ребро от того из её предшественников, чей новый срок и потребовал
	// сдвига). Для исходной задачи (первый элемент результата) не заполняется:
	// её сдвиг задан явно, а не вызван зависимостью.
	CausedBy models.DependencyType
}

// Shift пересчитывает минимально необходимый каскад дат при переносе задачи
// sourceID на newStart; длительность задачи (End-Start) сохраняется.
//
// Возвращает только задачи, чьи даты реально изменились, в порядке
// распространения (сама sourceID — всегда первым элементом). Последователь
// сдвигается, только если его текущий запланированный старт перестал
// удовлетворять требованию хотя бы одного (тоже сдвинутого) предшественника —
// у задач с запасом (буфером) относительно предшественника даты не меняются.
// Это и есть каскад «подвинули одну — поехали только те, кому не хватило
// запаса», а не полный пересчёт графа с нуля.
func (g *Graph) Shift(sourceID uuid.UUID, newStart models.Date) ([]ShiftedTask, error) {
	source, ok := g.tasks[sourceID]
	if !ok {
		return nil, fmt.Errorf("задача %s не найдена в графе", sourceID)
	}

	current := make(map[uuid.UUID]dateRange, len(g.tasks))
	for id, t := range g.tasks {
		current[id] = dateRange{Start: t.Start, End: t.End}
	}
	changed := make(map[uuid.UUID]bool, len(g.tasks))
	causedBy := make(map[uuid.UUID]models.DependencyType, len(g.tasks))
	var order []uuid.UUID

	apply := func(id uuid.UUID, r dateRange) {
		if !changed[id] {
			order = append(order, id)
		}
		current[id] = r
		changed[id] = true
	}

	apply(sourceID, dateRange{Start: newStart, End: newStart.AddDays(source.duration())})

	started := false
	for _, id := range g.order {
		if id == sourceID {
			started = true
			continue
		}
		if !started {
			// Задачи до источника в топологическом порядке недостижимы от
			// него (предшественник всегда раньше последователя) — можно
			// пропустить, changed[] для них так и останется false.
			continue
		}

		t := g.tasks[id]
		duration := t.duration()
		start := current[id].Start
		moved := false
		var cause models.DependencyType
		for _, d := range g.in[id] {
			if !changed[d.PredecessorID] {
				continue
			}
			required := requiredSuccessorStart(current[d.PredecessorID], d, duration)
			if required.After(start) {
				start = required
				moved = true
				cause = d.Type
			}
		}
		if moved {
			apply(id, dateRange{Start: start, End: start.AddDays(duration)})
			causedBy[id] = cause
		}
	}

	result := make([]ShiftedTask, 0, len(order))
	for _, id := range order {
		orig := g.tasks[id]
		result = append(result, ShiftedTask{
			ID:            id,
			OriginalStart: orig.Start,
			OriginalEnd:   orig.End,
			NewStart:      current[id].Start,
			NewEnd:        current[id].End,
			CausedBy:      causedBy[id],
		})
	}
	return result, nil
}

// Compute считает CPM по всему графу.
//
// Обратный проход всегда ведётся от ComputedFinish — собственного, естественного
// финиша графа (самого длинного пути по сети), а не от дедлайна проекта.
// Так критический путь показывает структуру графа («какая цепочка задаёт длину
// проекта») независимо от того, есть ли у проекта запас до дедлайна. Если
// анкерить обратный проход на дедлайне, при наличии всего пары дней запаса
// резерв размазывается по всем задачам и критический путь перестаёт быть видно
// вообще — проверено на живых данных, это не гипотетический случай.
// Сравнение ComputedFinish с дедлайном проекта — отдельная забота вызывающего
// кода (см. dashboard), не этого пакета.
func (g *Graph) Compute() Summary {
	forward := g.forward()

	var computedFinish models.Date
	first := true
	for _, r := range forward {
		if first || r.End.After(computedFinish) {
			computedFinish = r.End
			first = false
		}
	}

	backward := g.backward(computedFinish)

	tasks := make(map[uuid.UUID]TaskSchedule, len(g.tasks))
	for id := range g.tasks {
		f, b := forward[id], backward[id]
		slack := f.Start.DaysUntil(b.Start)
		tasks[id] = TaskSchedule{
			EarlyStart:  f.Start,
			EarlyFinish: f.End,
			LateStart:   b.Start,
			LateFinish:  b.End,
			SlackDays:   slack,
			Critical:    slack <= 0,
		}
	}
	return Summary{Tasks: tasks, ComputedFinish: computedFinish}
}

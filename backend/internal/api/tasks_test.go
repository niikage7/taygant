package api_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// ---------- Задачи ----------

func TestCreateTaskComputesDerivedFields(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Постановщик")
	project := newProject(t, owner)
	monday := nextMonday()

	task := newTask(t, owner.Token, project.ID, map[string]any{
		"title":         "  Спроектировать схему БД ",
		"startDate":     monday.String(),
		"endDate":       monday.AddDays(6).String(), // по воскресенье
		"assigneeId":    owner.ID,
		"weightPercent": 12.5,
	})

	if task.Code != "TASK-001" || task.WBSNumber != "1" || task.Status != "planned" {
		t.Fatalf("code=%q wbs=%q status=%q, ожидались TASK-001, 1, planned", task.Code, task.WBSNumber, task.Status)
	}
	if task.Title != "Спроектировать схему БД" {
		t.Fatalf("title = %q, ожидались обрезанные пробелы", task.Title)
	}
	if task.Assignee == nil || task.Assignee.ID != owner.ID {
		t.Fatalf("assignee = %+v, ожидался владелец", task.Assignee)
	}
	if task.WeightPercent != 12.5 || task.ProgressPercent != 0 {
		t.Fatalf("weight=%v progress=%d, ожидались 12.5 и 0", task.WeightPercent, task.ProgressPercent)
	}
	// Неделя пн–вс: при пятидневке 5 рабочих дней.
	if task.DurationWorkingDays != 5 {
		t.Fatalf("durationWorkingDays = %d при календаре 5/2, ожидалось 5", task.DurationWorkingDays)
	}

	// Смена календаря проекта сразу меняет расчёт: при шестидневке суббота рабочая.
	patch(t, owner.Token, "/projects/"+project.ID, map[string]any{"workingCalendarType": "6/1"}).want(t, http.StatusOK)
	got := decode[taskJSON](t, get(t, owner.Token, "/tasks/"+task.ID).want(t, http.StatusOK))
	if got.DurationWorkingDays != 6 {
		t.Fatalf("durationWorkingDays = %d при календаре 6/1, ожидалось 6", got.DurationWorkingDays)
	}
}

// Календарная длительность считается как разность дат (конец не включается),
// а рабочие дни — включительно. У однодневной задачи получается 0 календарных
// и 1 рабочий день: рабочих больше, чем календарных, — на Ганте это сдвиг на день.
func TestTaskDurationsUseSameCounting(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Хронометрист")
	project := newProject(t, owner)
	monday := nextMonday().String()

	task := newTask(t, owner.Token, project.ID, map[string]any{"startDate": monday, "endDate": monday})
	if task.DurationWorkingDays > task.DurationCalendarDays {
		t.Fatalf("однодневная задача: рабочих дней %d, календарных %d — длительности считаются по разным правилам",
			task.DurationWorkingDays, task.DurationCalendarDays)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Проверяющий")
	project := newProject(t, owner)
	path := "/projects/" + project.ID + "/tasks"

	cases := []struct {
		name string
		body map[string]any
	}{
		{"без названия", map[string]any{"title": " ", "startDate": day(1), "endDate": day(2)}},
		{"конец раньше начала", map[string]any{"title": "Т", "startDate": day(5), "endDate": day(1)}},
		{"статус overdue ставит только система", map[string]any{"title": "Т", "startDate": day(1), "endDate": day(2), "status": "overdue"}},
		{"статус blocked ставит только система", map[string]any{"title": "Т", "startDate": day(1), "endDate": day(2), "status": "blocked"}},
		{"дата в чужом формате", map[string]any{"title": "Т", "startDate": "10.09.2026", "endDate": day(2)}},
		// Даты обязательны по спецификации, но их отсутствие не проверяется:
		// пустая дата долетает до INSERT в NOT NULL-колонку и возвращается как 500.
		{"без дат", map[string]any{"title": "Т"}},
		{"пустая строка вместо даты", map[string]any{"title": "Т", "startDate": "", "endDate": day(2)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			post(t, owner.Token, path, tc.body).want(t, http.StatusBadRequest)
		})
	}
}

func TestTaskReferencesMustBelongToProject(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Хозяин")
	stranger := newUser(t, "Чужак")
	project := newProject(t, owner)
	other := newProject(t, owner)
	otherSprint := decode[sprintJSON](t, post(t, owner.Token, "/projects/"+other.ID+"/sprints", map[string]any{
		"name": "Чужой спринт", "startDate": day(1), "endDate": day(5),
	}).want(t, http.StatusCreated))
	otherTask := newTask(t, owner.Token, other.ID, nil)
	path := "/projects/" + project.ID + "/tasks"

	cases := []struct {
		name, field, value string
	}{
		{"ответственный не в команде", "assigneeId", stranger.ID},
		{"спринт другого проекта", "sprintId", otherSprint.ID},
		{"родитель из другого проекта", "parentTaskId", otherTask.ID},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			post(t, owner.Token, path, map[string]any{
				"title": "Т", "startDate": day(1), "endDate": day(2), tc.field: tc.value,
			}).want(t, http.StatusBadRequest)
		})
	}

	task := newTask(t, owner.Token, project.ID, nil)
	patch(t, owner.Token, "/tasks/"+task.ID, map[string]any{"assigneeId": stranger.ID}).want(t, http.StatusBadRequest)
	patch(t, owner.Token, "/tasks/"+task.ID, map[string]any{"sprintId": otherSprint.ID}).want(t, http.StatusBadRequest)
}

func TestSubtaskWBSNumbering(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Декомпозитор")
	project := newProject(t, owner)

	parent := newTask(t, owner.Token, project.ID, nil)
	child1 := newTask(t, owner.Token, project.ID, map[string]any{"parentTaskId": parent.ID})
	child2 := newTask(t, owner.Token, project.ID, map[string]any{"parentTaskId": parent.ID})
	sibling := newTask(t, owner.Token, project.ID, nil)
	grandchild := newTask(t, owner.Token, project.ID, map[string]any{"parentTaskId": child2.ID})

	got := []string{parent.WBSNumber, child1.WBSNumber, child2.WBSNumber, sibling.WBSNumber, grandchild.WBSNumber}
	want := []string{"1", "1.1", "1.2", "2", "1.2.1"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("номера WBS = %v, ожидались %v", got, want)
		}
	}
	if child1.ParentTaskID == nil || *child1.ParentTaskID != parent.ID {
		t.Fatalf("parentTaskId подзадачи = %v", child1.ParentTaskID)
	}
}

// Код задачи, номер WBS и порядок считаются как «число задач в проекте + 1».
// После удаления любой задачи, кроме последней, следующая получает уже занятые
// значения: в проекте появляются две TASK-003 и два пункта WBS «3».
// Тот же приём (count как номер) использует порядок пунктов чек-листа.
func TestTaskNumberingStaysUniqueAfterDeletion(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Удаляющий")
	project := newProject(t, owner)

	first := newTask(t, owner.Token, project.ID, nil)
	newTask(t, owner.Token, project.ID, nil)
	newTask(t, owner.Token, project.ID, nil)
	del(t, owner.Token, "/tasks/"+first.ID).want(t, http.StatusNoContent)
	newTask(t, owner.Token, project.ID, nil)

	tasks := decode[[]taskJSON](t, get(t, owner.Token, "/projects/"+project.ID+"/tasks").want(t, http.StatusOK))
	codes, wbs := map[string]int{}, map[string]int{}
	for _, task := range tasks {
		codes[task.Code]++
		wbs[task.WBSNumber]++
	}
	for code, n := range codes {
		if n > 1 {
			t.Errorf("код %s встречается %d раза", code, n)
		}
	}
	for number, n := range wbs {
		if n > 1 {
			t.Errorf("номер WBS %s встречается %d раза", number, n)
		}
	}
}

func TestUpdateTaskIsPartial(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Редактор задач")
	project := newProject(t, owner)
	task := newTask(t, owner.Token, project.ID, map[string]any{"description": "Описание"})
	path := "/tasks/" + task.ID

	updated := decode[taskJSON](t, patch(t, owner.Token, path, map[string]any{
		"title": "Новое название", "progressPercent": 40,
	}).want(t, http.StatusOK))
	if updated.Title != "Новое название" || updated.ProgressPercent != 40 {
		t.Fatalf("после PATCH title=%q progress=%d", updated.Title, updated.ProgressPercent)
	}
	if updated.StartDate != task.StartDate || updated.EndDate != task.EndDate || updated.Code != task.Code {
		t.Fatalf("частичное обновление задело другие поля: было %+v, стало %+v", task, updated)
	}
	if detail := decode[taskDetailJSON](t, get(t, owner.Token, path).want(t, http.StatusOK)); detail.Description != "Описание" {
		t.Fatalf("описание потерялось: %q", detail.Description)
	}

	cases := []struct {
		name string
		body map[string]any
	}{
		{"прогресс больше 100", map[string]any{"progressPercent": 150}},
		{"отрицательный прогресс", map[string]any{"progressPercent": -1}},
		{"пустое название", map[string]any{"title": "  "}},
		{"конец раньше начала", map[string]any{"endDate": day(-30)}},
		{"системный статус", map[string]any{"status": "blocked"}},
		{"неизвестный статус", map[string]any{"status": "paused"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			patch(t, owner.Token, path, tc.body).want(t, http.StatusBadRequest)
		})
	}

	get(t, owner.Token, "/tasks/"+uuid.NewString()).want(t, http.StatusNotFound)
	patch(t, owner.Token, "/tasks/"+uuid.NewString(), map[string]any{"title": "x"}).want(t, http.StatusNotFound)
	get(t, owner.Token, "/tasks/not-a-uuid").want(t, http.StatusBadRequest)
}

func TestTaskStatusIsDerivedFromDatesAndDependencies(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Контролёр")
	project := newProject(t, owner)

	late := newTask(t, owner.Token, project.ID, map[string]any{"startDate": day(-10), "endDate": day(-5)})
	if late.Status != "overdue" || late.PlanVsActualDeviationDays != -5 {
		t.Fatalf("просроченная задача: status=%q deviation=%d, ожидались overdue и -5", late.Status, late.PlanVsActualDeviationDays)
	}

	finished := newTask(t, owner.Token, project.ID, map[string]any{"startDate": day(-10), "endDate": day(-5), "status": "done"})
	if finished.Status != "done" {
		t.Fatalf("завершённая задача с прошедшим сроком получила статус %q", finished.Status)
	}

	first := newTask(t, owner.Token, project.ID, nil)
	second := newTask(t, owner.Token, project.ID, map[string]any{
		"startDate":    day(11),
		"endDate":      day(20),
		"predecessors": []map[string]any{{"taskId": first.ID, "type": "FS"}},
	})
	if second.Status != "blocked" {
		t.Fatalf("задача с незавершённым предшественником: статус %q, ожидался blocked", second.Status)
	}

	// Регрессия из комментария к resolveStatuses: после завершения
	// предшественника карточка последователя не должна остаться blocked.
	patch(t, owner.Token, "/tasks/"+first.ID, map[string]any{"status": "done"}).want(t, http.StatusOK)
	if got := decode[taskJSON](t, get(t, owner.Token, "/tasks/"+second.ID).want(t, http.StatusOK)); got.Status != "planned" {
		t.Fatalf("после завершения предшественника статус %q, ожидался planned", got.Status)
	}
}

func TestTaskListFilters(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Фильтровщик")
	dev := newUser(t, "Исполнитель")
	project := newProject(t, owner)
	addMember(t, owner, project.ID, dev, "edit")
	sprint := decode[sprintJSON](t, post(t, owner.Token, "/projects/"+project.ID+"/sprints", map[string]any{
		"name": "С1", "startDate": day(0), "endDate": day(14),
	}).want(t, http.StatusCreated))

	mine := newTask(t, owner.Token, project.ID, map[string]any{"assigneeId": dev.ID, "sprintId": sprint.ID})
	late := newTask(t, owner.Token, project.ID, map[string]any{"startDate": day(-9), "endDate": day(-2)})
	marked := newTask(t, owner.Token, project.ID, map[string]any{"title": "Интеграция с GitLab"})

	list := func(token, query string) []string {
		t.Helper()
		tasks := decode[[]taskJSON](t, get(t, token, "/projects/"+project.ID+"/tasks"+query).want(t, http.StatusOK))
		ids := make([]string, 0, len(tasks))
		for _, task := range tasks {
			ids = append(ids, task.ID)
		}
		return ids
	}

	if got := list(owner.Token, ""); len(got) != 3 {
		t.Fatalf("без фильтров %d задач, ожидалось 3", len(got))
	}
	cases := []struct {
		name, token, query, want string
	}{
		{"assigneeId", owner.Token, "?assigneeId=" + dev.ID, mine.ID},
		{"myTasksOnly", dev.Token, "?myTasksOnly=true", mine.ID},
		{"sprintId", owner.Token, "?sprintId=" + sprint.ID, mine.ID},
		{"status=overdue", owner.Token, "?status=overdue", late.ID},
		{"risksOnly", owner.Token, "?risksOnly=true", late.ID},
		{"search без учёта регистра", owner.Token, "?search=gitlab", marked.ID},
	}
	for _, tc := range cases {
		if got := list(tc.token, tc.query); len(got) != 1 || got[0] != tc.want {
			t.Errorf("фильтр %s вернул %v, ожидалась одна задача %s", tc.name, got, tc.want)
		}
	}

	get(t, owner.Token, "/projects/"+project.ID+"/tasks?status=paused").want(t, http.StatusBadRequest)
	get(t, owner.Token, "/projects/"+project.ID+"/tasks?sprintId=abc").want(t, http.StatusBadRequest)
}

func TestTaskPermissionsByAccessLevel(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Владелец")
	viewer := newUser(t, "Наблюдатель")
	outsider := newUser(t, "Посторонний")
	project := newProject(t, owner)
	addMember(t, owner, project.ID, viewer, "view")
	task := newTask(t, owner.Token, project.ID, nil)
	path := "/tasks/" + task.ID

	// Наблюдатель видит план и может обсуждать задачи, но не менять их.
	get(t, viewer.Token, "/projects/"+project.ID+"/tasks").want(t, http.StatusOK)
	get(t, viewer.Token, path).want(t, http.StatusOK)
	post(t, viewer.Token, path+"/comments", map[string]any{"text": "Вопрос по срокам"}).want(t, http.StatusCreated)
	post(t, viewer.Token, "/projects/"+project.ID+"/tasks", map[string]any{"title": "Т", "startDate": day(1), "endDate": day(2)}).want(t, http.StatusForbidden)
	patch(t, viewer.Token, path, map[string]any{"title": "Правка"}).want(t, http.StatusForbidden)
	post(t, viewer.Token, path+"/checklist", map[string]any{"text": "Пункт"}).want(t, http.StatusForbidden)
	del(t, viewer.Token, path).want(t, http.StatusForbidden)

	// Постороннему не положено даже знать, что задача существует.
	for _, tc := range []struct {
		name string
		r    response
	}{
		{"GET задачи", get(t, outsider.Token, path)},
		{"PATCH задачи", patch(t, outsider.Token, path, map[string]any{"title": "Взлом"})},
		{"DELETE задачи", del(t, outsider.Token, path)},
		{"комментарии", get(t, outsider.Token, path+"/comments")},
		{"история", get(t, outsider.Token, path+"/history")},
		{"чек-лист", get(t, outsider.Token, path+"/checklist")},
		{"связи", get(t, outsider.Token, path+"/dependencies")},
	} {
		if tc.r.Status != http.StatusNotFound {
			t.Errorf("%s для постороннего: код %d, ожидался 404", tc.name, tc.r.Status)
		}
	}
}

// weightPercent не проверяется: отрицательный вес принимается и ломает
// взвешенный прогресс проекта — он выходит за пределы 0–100%.
func TestTaskWeightIsValidated(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Весовщик")
	project := newProject(t, owner)

	done := newTask(t, owner.Token, project.ID, map[string]any{"weightPercent": 10})
	patch(t, owner.Token, "/tasks/"+done.ID, map[string]any{"progressPercent": 100}).want(t, http.StatusOK)

	r := post(t, owner.Token, "/projects/"+project.ID+"/tasks", map[string]any{
		"title": "Отрицательный вес", "startDate": day(1), "endDate": day(2), "weightPercent": -5,
	})
	if r.Status == http.StatusBadRequest {
		return
	}
	p := decode[projectJSON](t, get(t, owner.Token, "/projects/"+project.ID).want(t, http.StatusOK))
	t.Fatalf("отрицательный weightPercent принят (код %d), прогресс проекта теперь %.0f%%", r.Status, p.ProgressPercent)
}

func TestProjectMetricsFollowTasks(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Аналитик")
	project := newProject(t, owner)

	heavy := newTask(t, owner.Token, project.ID, map[string]any{"weightPercent": 30})
	newTask(t, owner.Token, project.ID, map[string]any{"weightPercent": 70})
	patch(t, owner.Token, "/tasks/"+heavy.ID, map[string]any{"progressPercent": 100}).want(t, http.StatusOK)

	p := decode[projectJSON](t, get(t, owner.Token, "/projects/"+project.ID).want(t, http.StatusOK))
	if p.ProgressPercent != 30 {
		t.Fatalf("взвешенный прогресс = %v, ожидалось 30 (задача весом 30%% выполнена)", p.ProgressPercent)
	}

	// Каждая просроченная задача — минус 10 к индексу здоровья и один критический риск.
	newTask(t, owner.Token, project.ID, map[string]any{"startDate": day(-10), "endDate": day(-1)})
	p = decode[projectJSON](t, get(t, owner.Token, "/projects/"+project.ID).want(t, http.StatusOK))
	if p.HealthIndex != 90 || p.CriticalRisksCount != 1 {
		t.Fatalf("health=%d risks=%d, ожидались 90 и 1", p.HealthIndex, p.CriticalRisksCount)
	}
}

func TestDeleteTaskRemovesItsLinksChecklistAndComments(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Чистильщик")
	project := newProject(t, owner)
	keep := newTask(t, owner.Token, project.ID, nil)
	doomed := newTask(t, owner.Token, project.ID, map[string]any{
		"predecessors": []map[string]any{{"taskId": keep.ID, "type": "FS"}},
	})
	item := decode[checklistJSON](t, post(t, owner.Token, "/tasks/"+doomed.ID+"/checklist", map[string]any{"text": "Пункт"}).want(t, http.StatusCreated))
	post(t, owner.Token, "/tasks/"+doomed.ID+"/comments", map[string]any{"text": "Комментарий"}).want(t, http.StatusCreated)

	del(t, owner.Token, "/tasks/"+doomed.ID).want(t, http.StatusNoContent)
	get(t, owner.Token, "/tasks/"+doomed.ID).want(t, http.StatusNotFound)
	patch(t, owner.Token, "/checklist/"+item.ID, map[string]any{"isDone": true}).want(t, http.StatusNotFound)
	if deps := decode[dependencyListJSON](t, get(t, owner.Token, "/tasks/"+keep.ID+"/dependencies").want(t, http.StatusOK)); len(deps.Successors) != 0 {
		t.Fatalf("связь с удалённой задачей осталась: %+v", deps.Successors)
	}
	del(t, owner.Token, "/tasks/"+doomed.ID).want(t, http.StatusNotFound)
}

// ---------- Связи между задачами ----------

func TestDependencyLifecycle(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Связист")
	project := newProject(t, owner)
	design := newTask(t, owner.Token, project.ID, map[string]any{"title": "Дизайн"})
	build := newTask(t, owner.Token, project.ID, map[string]any{"title": "Вёрстка"})

	dep := decode[dependencyJSON](t, post(t, owner.Token, "/tasks/"+build.ID+"/dependencies", map[string]any{
		"relatedTaskId": design.ID, "direction": "predecessor", "type": "FS", "lagDays": 2,
	}).want(t, http.StatusCreated))
	if dep.PredecessorTaskID != design.ID || dep.SuccessorTaskID != build.ID ||
		dep.PredecessorTitle != "Дизайн" || dep.SuccessorTitle != "Вёрстка" || dep.LagDays != 2 {
		t.Fatalf("созданная связь = %+v", dep)
	}

	fromBuild := decode[dependencyListJSON](t, get(t, owner.Token, "/tasks/"+build.ID+"/dependencies").want(t, http.StatusOK))
	fromDesign := decode[dependencyListJSON](t, get(t, owner.Token, "/tasks/"+design.ID+"/dependencies").want(t, http.StatusOK))
	if len(fromBuild.Predecessors) != 1 || len(fromBuild.Successors) != 0 || len(fromDesign.Successors) != 1 {
		t.Fatalf("связь видна неверно: у вёрстки %+v, у дизайна %+v", fromBuild, fromDesign)
	}
	if detail := decode[taskDetailJSON](t, get(t, owner.Token, "/tasks/"+build.ID).want(t, http.StatusOK)); len(detail.Predecessors) != 1 || detail.Predecessors[0].ID != dep.ID {
		t.Fatalf("в карточке задачи предшественники = %+v", detail.Predecessors)
	}

	// Повтор той же пары — конфликт, даже если задать её с другой стороны и другим типом.
	post(t, owner.Token, "/tasks/"+build.ID+"/dependencies", map[string]any{
		"relatedTaskId": design.ID, "direction": "predecessor", "type": "FS",
	}).want(t, http.StatusConflict)
	post(t, owner.Token, "/tasks/"+design.ID+"/dependencies", map[string]any{
		"relatedTaskId": build.ID, "direction": "successor", "type": "SS",
	}).want(t, http.StatusConflict)

	del(t, owner.Token, "/dependencies/"+dep.ID).want(t, http.StatusNoContent)
	del(t, owner.Token, "/dependencies/"+dep.ID).want(t, http.StatusNotFound)
}

func TestDependencyValidation(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Логик")
	project := newProject(t, owner)
	a := newTask(t, owner.Token, project.ID, nil)
	b := newTask(t, owner.Token, project.ID, nil)
	c := newTask(t, owner.Token, project.ID, nil)
	foreign := newTask(t, owner.Token, newProject(t, owner).ID, nil)

	// link делает from предшественником to.
	link := func(from, to, depType, direction string) response {
		t.Helper()
		return post(t, owner.Token, "/tasks/"+to+"/dependencies", map[string]any{
			"relatedTaskId": from, "direction": direction, "type": depType,
		})
	}
	link(a.ID, b.ID, "FS", "predecessor").want(t, http.StatusCreated)
	link(b.ID, c.ID, "FS", "predecessor").want(t, http.StatusCreated)

	for _, tc := range []struct {
		name string
		r    response
	}{
		{"задача сама от себя", link(a.ID, a.ID, "FS", "predecessor")},
		{"обратная связь замыкает цикл", link(b.ID, a.ID, "FS", "predecessor")},
		{"цикл через три задачи", link(c.ID, a.ID, "FS", "predecessor")},
		{"задача из другого проекта", link(foreign.ID, a.ID, "FS", "predecessor")},
		{"неизвестный тип связи", link(a.ID, c.ID, "XX", "predecessor")},
		{"неизвестное направление", link(a.ID, c.ID, "FS", "sideways")},
	} {
		if tc.r.Status != http.StatusBadRequest {
			t.Errorf("%s: код %d, ожидался 400: %s", tc.name, tc.r.Status, tc.r.Body)
		}
	}
}

// Задача и её предшественники создаются в одной транзакции.
func TestCreateTaskWithPredecessorsIsAtomic(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Атомарный")
	project := newProject(t, owner)
	foreign := newTask(t, owner.Token, newProject(t, owner).ID, nil)
	before := newTask(t, owner.Token, project.ID, nil)

	task := newTask(t, owner.Token, project.ID, map[string]any{
		"predecessors": []map[string]any{{"taskId": before.ID, "type": "SS", "lagDays": -1}},
	})
	detail := decode[taskDetailJSON](t, get(t, owner.Token, "/tasks/"+task.ID).want(t, http.StatusOK))
	if len(detail.Predecessors) != 1 || detail.Predecessors[0].Type != "SS" || detail.Predecessors[0].LagDays != -1 {
		t.Fatalf("предшественники новой задачи = %+v", detail.Predecessors)
	}

	post(t, owner.Token, "/projects/"+project.ID+"/tasks", map[string]any{
		"title": "Не должна появиться", "startDate": day(1), "endDate": day(2),
		"predecessors": []map[string]any{{"taskId": foreign.ID, "type": "FS"}},
	}).want(t, http.StatusBadRequest)
	leftovers := decode[[]taskJSON](t, get(t, owner.Token, "/projects/"+project.ID+"/tasks?search="+urlQuery("Не должна")).want(t, http.StatusOK))
	if len(leftovers) != 0 {
		t.Fatalf("задача осталась после ошибки в предшественниках: %+v", leftovers)
	}
}

// ---------- Чек-лист, комментарии, журнал ----------

func TestChecklistLifecycle(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Приёмщик")
	project := newProject(t, owner)
	task := newTask(t, owner.Token, project.ID, nil)
	path := "/tasks/" + task.ID + "/checklist"

	first := decode[checklistJSON](t, post(t, owner.Token, path, map[string]any{"text": "Код прошёл ревью"}).want(t, http.StatusCreated))
	second := decode[checklistJSON](t, post(t, owner.Token, path, map[string]any{"text": "Тесты зелёные"}).want(t, http.StatusCreated))
	post(t, owner.Token, path, map[string]any{"text": "  "}).want(t, http.StatusBadRequest)

	done := decode[checklistJSON](t, patch(t, owner.Token, "/checklist/"+first.ID, map[string]any{"isDone": true}).want(t, http.StatusOK))
	if !done.IsDone || done.Text != "Код прошёл ревью" {
		t.Fatalf("после отметки пункт = %+v", done)
	}
	patch(t, owner.Token, "/checklist/"+first.ID, map[string]any{"text": ""}).want(t, http.StatusBadRequest)

	items := decode[[]checklistJSON](t, get(t, owner.Token, path).want(t, http.StatusOK))
	if len(items) != 2 || items[0].ID != first.ID || items[1].ID != second.ID || !items[0].IsDone {
		t.Fatalf("чек-лист = %+v, ожидались два пункта в порядке добавления", items)
	}

	del(t, owner.Token, "/checklist/"+second.ID).want(t, http.StatusNoContent)
	del(t, owner.Token, "/checklist/"+second.ID).want(t, http.StatusNotFound)
	if detail := decode[taskDetailJSON](t, get(t, owner.Token, "/tasks/"+task.ID).want(t, http.StatusOK)); len(detail.Checklist) != 1 {
		t.Fatalf("в карточке задачи %d пунктов чек-листа, ожидался 1", len(detail.Checklist))
	}
}

func TestCommentsAndHistory(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Автор")
	dev := newUser(t, "Исполнитель")
	project := newProject(t, owner)
	addMember(t, owner, project.ID, dev, "edit")
	task := newTask(t, owner.Token, project.ID, nil)
	path := "/tasks/" + task.ID

	comment := decode[commentJSON](t, post(t, dev.Token, path+"/comments", map[string]any{"text": "  Начинаю  "}).want(t, http.StatusCreated))
	if comment.Author.ID != dev.ID || comment.Text != "Начинаю" {
		t.Fatalf("комментарий = %+v", comment)
	}
	post(t, dev.Token, path+"/comments", map[string]any{"text": ""}).want(t, http.StatusBadRequest)

	if comments := decode[[]commentJSON](t, get(t, owner.Token, path+"/comments").want(t, http.StatusOK)); len(comments) != 1 {
		t.Fatalf("комментариев %d, ожидался 1", len(comments))
	}
	if detail := decode[taskDetailJSON](t, get(t, owner.Token, path).want(t, http.StatusOK)); detail.CommentsCount != 1 {
		t.Fatalf("commentsCount = %d, ожидался 1", detail.CommentsCount)
	}

	patch(t, owner.Token, path, map[string]any{"assigneeId": dev.ID}).want(t, http.StatusOK)
	patch(t, dev.Token, path, map[string]any{"status": "in_progress"}).want(t, http.StatusOK)

	entries := decode[[]historyJSON](t, get(t, owner.Token, path+"/history").want(t, http.StatusOK))
	actions := make([]string, 0, len(entries))
	for _, e := range entries {
		actions = append(actions, e.Action)
	}
	want := []string{"created", "commented", "assigned", "status_changed"}
	if len(actions) != len(want) {
		t.Fatalf("журнал = %v, ожидалось %v", actions, want)
	}
	for i := range want {
		if actions[i] != want[i] {
			t.Fatalf("журнал = %v, ожидалось %v", actions, want)
		}
	}
	last := entries[len(entries)-1]
	if last.Actor.ID != dev.ID || last.OldValue == nil || *last.OldValue != "planned" || last.NewValue == nil || *last.NewValue != "in_progress" {
		t.Fatalf("запись о смене статуса = %+v", last)
	}
}

// Журнал задуман как «что, кто, когда, с чего на что» по каждому полю, и коды
// date_shifted, progress_set, dependency_added объявлены в models. Но пишутся
// только создание, статус, ответственный, комментарии и чек-лист: перенос
// сроков, прогресс и связи в журнал не попадают.
func TestHistoryRecordsDatesProgressAndDependencies(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Аудитор")
	project := newProject(t, owner)
	other := newTask(t, owner.Token, project.ID, nil)
	task := newTask(t, owner.Token, project.ID, nil)

	patch(t, owner.Token, "/tasks/"+task.ID, map[string]any{"endDate": day(20), "progressPercent": 50}).want(t, http.StatusOK)
	post(t, owner.Token, "/tasks/"+task.ID+"/dependencies", map[string]any{
		"relatedTaskId": other.ID, "direction": "predecessor", "type": "FS",
	}).want(t, http.StatusCreated)

	actions := historyActions(t, owner.Token, task.ID)
	for _, want := range []string{"date_shifted", "progress_set", "dependency_added"} {
		if !contains(actions, want) {
			t.Errorf("в журнале нет записи %q, есть только %v", want, actions)
		}
	}
}

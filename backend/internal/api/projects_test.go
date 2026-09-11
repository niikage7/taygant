package api_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// ---------- Проекты ----------

func TestCreateProjectDefaultsAndCreatorMembership(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Владелец")

	project := newProject(t, owner)
	if project.Status != "active" || project.WorkingCalendarType != "5/2" {
		t.Fatalf("status=%q calendar=%q, ожидались active и 5/2 по умолчанию", project.Status, project.WorkingCalendarType)
	}
	if !regexp.MustCompile(`^PRJ-\d{4}-[0-9A-F]{4}$`).MatchString(project.Code) {
		t.Fatalf("код проекта %q не похож на PRJ-2026-A1B2", project.Code)
	}
	if project.DurationCalendarDays != 90 {
		t.Fatalf("durationCalendarDays = %d, ожидалось 90", project.DurationCalendarDays)
	}
	if project.CreatedBy == nil || project.CreatedBy.ID != owner.ID {
		t.Fatalf("createdBy = %+v, ожидался создатель проекта", project.CreatedBy)
	}
	if project.HealthIndex != 100 || project.ProgressPercent != 0 {
		t.Fatalf("у пустого проекта health=%d progress=%v, ожидались 100 и 0", project.HealthIndex, project.ProgressPercent)
	}

	// Создатель сразу в команде с полным доступом — иначе он не прошёл бы
	// собственную проверку доступа к проекту.
	members := decode[[]memberJSON](t, get(t, owner.Token, "/projects/"+project.ID+"/members").want(t, http.StatusOK))
	if len(members) != 1 || members[0].User.ID != owner.ID || members[0].AccessLevel != "full" || members[0].ProjectRole != "project_manager" {
		t.Fatalf("команда нового проекта = %+v, ожидался один создатель с full/project_manager", members)
	}

	draft := decode[projectJSON](t, post(t, owner.Token, "/projects", map[string]any{
		"name": "Черновик", "startDate": day(0), "deadline": day(10), "saveAsDraft": true,
	}).want(t, http.StatusCreated))
	if draft.Status != "draft" {
		t.Fatalf("saveAsDraft: статус %q, ожидался draft", draft.Status)
	}
}

func TestCreateProjectValidation(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Валидатор")

	cases := []struct {
		name string
		body map[string]any
	}{
		{"без названия", map[string]any{"name": " ", "startDate": day(0), "deadline": day(10)}},
		{"без сроков", map[string]any{"name": "Проект"}},
		{"дедлайн раньше старта", map[string]any{"name": "Проект", "startDate": day(10), "deadline": day(0)}},
		{"неизвестный календарь", map[string]any{"name": "Проект", "startDate": day(0), "deadline": day(10), "workingCalendarType": "7/0"}},
		{"дата в чужом формате", map[string]any{"name": "Проект", "startDate": "10.09.2026", "deadline": day(10)}},
		// Спецификация ограничивает name 120 символами, в БД varchar(120),
		// а в коде проверки нет — длинное название долетает до INSERT и даёт 500.
		{"название длиннее 120 символов", map[string]any{"name": strings.Repeat("Я", 121), "startDate": day(0), "deadline": day(10)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			post(t, owner.Token, "/projects", tc.body).want(t, http.StatusBadRequest)
		})
	}
}

// POST и PATCH возвращают createdBy, а GET того же проекта — null: хендлер
// отдаёт проект из access.Load, где связь CreatedBy не подгружается.
func TestGetProjectReturnsCreator(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Автор")
	project := newProject(t, owner)

	got := decode[projectJSON](t, get(t, owner.Token, "/projects/"+project.ID).want(t, http.StatusOK))
	if got.CreatedBy == nil || got.CreatedBy.ID != owner.ID {
		t.Fatalf("GET /projects/{id}: createdBy = %v, а POST для того же проекта вернул автора", got.CreatedBy)
	}
}

func TestProjectListShowsOnlyMemberProjectsAndFilters(t *testing.T) {
	requireDB(t)
	alice := newUser(t, "Алиса")
	bob := newUser(t, "Боб")

	marker := fmt.Sprintf("Ромашка-%d", uniq())
	active := decode[projectJSON](t, post(t, alice.Token, "/projects", map[string]any{
		"name": marker + " основной", "startDate": day(0), "deadline": day(30),
	}).want(t, http.StatusCreated))
	draft := decode[projectJSON](t, post(t, alice.Token, "/projects", map[string]any{
		"name": marker + " черновик", "startDate": day(0), "deadline": day(30), "saveAsDraft": true,
	}).want(t, http.StatusCreated))
	foreign := newProject(t, bob)

	list := func(query string) map[string]bool {
		t.Helper()
		projects := decode[[]projectJSON](t, get(t, alice.Token, "/projects"+query).want(t, http.StatusOK))
		ids := make(map[string]bool, len(projects))
		for _, p := range projects {
			ids[p.ID] = true
		}
		return ids
	}

	all := list("")
	if !all[active.ID] || !all[draft.ID] {
		t.Fatal("в списке нет собственных проектов")
	}
	if all[foreign.ID] {
		t.Fatal("в списке оказался чужой проект")
	}
	if drafts := list("?status=draft"); !drafts[draft.ID] || drafts[active.ID] {
		t.Fatalf("фильтр status=draft вернул %v", drafts)
	}
	if found := list("?search=" + urlQuery(strings.ToLower(marker)+" черн")); len(found) != 1 || !found[draft.ID] {
		t.Fatalf("поиск по названию вернул %v, ожидался один черновик", found)
	}
	get(t, alice.Token, "/projects?status=bogus").want(t, http.StatusBadRequest)
}

// Не участнику отвечаем 404, а не 403: иначе код ответа выдавал бы,
// что проект с таким id существует.
func TestProjectIsInvisibleToOutsiders(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Хозяин")
	outsider := newUser(t, "Посторонний")
	project := newProject(t, owner)
	path := "/projects/" + project.ID

	get(t, outsider.Token, path).want(t, http.StatusNotFound)
	patch(t, outsider.Token, path, map[string]any{"name": "Взлом"}).want(t, http.StatusNotFound)
	del(t, outsider.Token, path).want(t, http.StatusNotFound)
	get(t, outsider.Token, path+"/members").want(t, http.StatusNotFound)
	get(t, outsider.Token, path+"/tasks").want(t, http.StatusNotFound)

	get(t, owner.Token, "/projects/"+uuid.NewString()).want(t, http.StatusNotFound)
	get(t, owner.Token, "/projects/not-a-uuid").want(t, http.StatusBadRequest)
}

func TestUpdateProjectIsPartialAndRequiresFullAccess(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Руководитель")
	editor := newUser(t, "Редактор")
	viewer := newUser(t, "Наблюдатель")
	project := newProject(t, owner)
	addMember(t, owner, project.ID, editor, "edit")
	addMember(t, owner, project.ID, viewer, "view")
	path := "/projects/" + project.ID

	updated := decode[projectJSON](t, patch(t, owner.Token, path, map[string]any{"phase": "Фаза 2: Разработка"}).want(t, http.StatusOK))
	if updated.Phase != "Фаза 2: Разработка" || updated.Name != project.Name || updated.Deadline != project.Deadline {
		t.Fatalf("частичное обновление задело другие поля: %+v", updated)
	}

	patch(t, owner.Token, path, map[string]any{"status": "bogus"}).want(t, http.StatusBadRequest)
	patch(t, owner.Token, path, map[string]any{"deadline": day(-30)}).want(t, http.StatusBadRequest)
	patch(t, owner.Token, path, map[string]any{"name": ""}).want(t, http.StatusBadRequest)

	// Параметры проекта меняет только full; edit — это правка плана, а не проекта.
	patch(t, editor.Token, path, map[string]any{"phase": "x"}).want(t, http.StatusForbidden)
	patch(t, viewer.Token, path, map[string]any{"phase": "x"}).want(t, http.StatusForbidden)
	del(t, editor.Token, path).want(t, http.StatusForbidden)
}

func TestDeleteProjectArchivesInsteadOfRemoving(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Архивариус")
	project := newProject(t, owner)

	del(t, owner.Token, "/projects/"+project.ID).want(t, http.StatusNoContent)
	got := decode[projectJSON](t, get(t, owner.Token, "/projects/"+project.ID).want(t, http.StatusOK))
	if got.Status != "archived" {
		t.Fatalf("после DELETE статус %q, ожидался archived", got.Status)
	}

	// Архивирование — это «удаление» с точки зрения пользователя: из общего
	// списка проект пропадает, но остаётся доступен по явному фильтру.
	list := decode[[]projectJSON](t, get(t, owner.Token, "/projects").want(t, http.StatusOK))
	if len(list) != 0 {
		t.Fatalf("архивный проект остался в общем списке: %+v", list)
	}

	archived := decode[[]projectJSON](t, get(t, owner.Token, "/projects?status=archived").want(t, http.StatusOK))
	if len(archived) != 1 || archived[0].ID != project.ID {
		t.Fatalf("список status=archived = %+v, ожидался единственный архивный проект", archived)
	}
}

// Проект и стартовая команда создаются одним запросом (шаг 2 мастера).
func TestCreateProjectWithInitialTeam(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Инициатор")
	dev := newUser(t, "Разработчик")
	curator := newUser(t, "Куратор")

	project := decode[projectJSON](t, post(t, owner.Token, "/projects", map[string]any{
		"name": "С командой", "startDate": day(0), "deadline": day(30),
		"members": []map[string]any{
			{"userId": dev.ID, "projectRole": "developer", "sprintRole": "Бэкенд"},
			{"userId": curator.ID, "projectRole": "curator", "accessLevel": "view"},
			// Создатель в списке не дублируется и не теряет полный доступ.
			{"userId": owner.ID, "projectRole": "analyst", "accessLevel": "view"},
		},
	}).want(t, http.StatusCreated))

	members := decode[[]memberJSON](t, get(t, owner.Token, "/projects/"+project.ID+"/members").want(t, http.StatusOK))
	got := map[string]string{}
	for _, m := range members {
		got[m.User.ID] = m.AccessLevel
	}
	want := map[string]string{owner.ID: "full", dev.ID: "edit", curator.ID: "view"}
	if len(members) != len(want) {
		t.Fatalf("в команде %d записей, ожидалось %d: %+v", len(members), len(want), members)
	}
	for id, level := range want {
		if got[id] != level {
			t.Fatalf("уровень доступа %s = %q, ожидался %q", id, got[id], level)
		}
	}
}

// Проект и команда пишутся в одной транзакции: ошибка в составе команды
// не должна оставлять проект без участников.
func TestCreateProjectWithUnknownMemberIsRolledBack(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Откатчик")
	name := fmt.Sprintf("Откат-%d", uniq())

	post(t, owner.Token, "/projects", map[string]any{
		"name": name, "startDate": day(0), "deadline": day(30),
		"members": []map[string]any{{"userId": uuid.NewString(), "projectRole": "developer"}},
	}).want(t, http.StatusBadRequest)

	projects := decode[[]projectJSON](t, get(t, owner.Token, "/projects?search="+urlQuery(name)).want(t, http.StatusOK))
	if len(projects) != 0 {
		t.Fatalf("после ошибки остался проект: %+v", projects)
	}
}

// ---------- Команда проекта ----------

func TestMembersManagement(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Тимлид")
	bob := newUser(t, "Боб")
	editor := newUser(t, "Редактор")
	project := newProject(t, owner)
	path := "/projects/" + project.ID + "/members"
	addMember(t, owner, project.ID, editor, "edit")

	member := decode[memberJSON](t, post(t, owner.Token, path, map[string]any{
		"userId": bob.ID, "projectRole": "analyst", "sprintRole": "Аналитик",
	}).want(t, http.StatusCreated))
	if member.User.ID != bob.ID || member.AccessLevel != "edit" || member.SprintRole != "Аналитик" {
		t.Fatalf("добавлен %+v, ожидался Боб с доступом edit по умолчанию", member)
	}

	post(t, owner.Token, path, map[string]any{"userId": bob.ID, "projectRole": "analyst"}).want(t, http.StatusConflict)
	post(t, owner.Token, path, map[string]any{"userId": uuid.NewString(), "projectRole": "analyst"}).want(t, http.StatusBadRequest)
	post(t, owner.Token, path, map[string]any{"userId": bob.ID, "projectRole": "boss"}).want(t, http.StatusBadRequest)
	post(t, editor.Token, path, map[string]any{"userId": newUser(t, "Кто-то").ID, "projectRole": "other"}).want(t, http.StatusForbidden)

	updated := decode[memberJSON](t, patch(t, owner.Token, path+"/"+member.ID, map[string]any{
		"accessLevel": "view", "projectRole": "curator",
	}).want(t, http.StatusOK))
	if updated.AccessLevel != "view" || updated.ProjectRole != "curator" || updated.SprintRole != "Аналитик" {
		t.Fatalf("после PATCH участник = %+v", updated)
	}
	patch(t, owner.Token, path+"/"+member.ID, map[string]any{"accessLevel": "admin"}).want(t, http.StatusBadRequest)

	del(t, owner.Token, path+"/"+member.ID).want(t, http.StatusNoContent)
	get(t, bob.Token, "/projects/"+project.ID).want(t, http.StatusNotFound)
	del(t, owner.Token, path+"/"+member.ID).want(t, http.StatusNotFound)
}

func TestLastFullAccessMemberCannotBeDemotedOrRemoved(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Единственный")
	project := newProject(t, owner)
	path := "/projects/" + project.ID + "/members"

	self := decode[[]memberJSON](t, get(t, owner.Token, path).want(t, http.StatusOK))[0].ID

	patch(t, owner.Token, path+"/"+self, map[string]any{"accessLevel": "edit"}).want(t, http.StatusBadRequest)
	del(t, owner.Token, path+"/"+self).want(t, http.StatusBadRequest)

	// Когда есть второй участник с full, понизить себя можно.
	addMember(t, owner, project.ID, newUser(t, "Заместитель"), "full")
	patch(t, owner.Token, path+"/"+self, map[string]any{"accessLevel": "edit"}).want(t, http.StatusOK)
}

// Уязвимость: PATCH /projects/{projectId}/members/{memberId} проверяет права
// в проекте из URL, а участника ищет по memberId во всех проектах сразу.
// Любой может создать свой проект (там он full) и через его URL править
// участников чужих проектов — например, поднять себе доступ до full.
func TestMemberUpdateIsScopedToProjectInURL(t *testing.T) {
	requireDB(t)
	alice := newUser(t, "Алиса")
	mallory := newUser(t, "Мэллори")
	victim := newProject(t, alice)
	malloryInVictim := addMember(t, alice, victim.ID, mallory, "view")
	own := newProject(t, mallory)

	r := patch(t, mallory.Token, "/projects/"+own.ID+"/members/"+malloryInVictim.ID, map[string]any{"accessLevel": "full"})
	if r.Status != http.StatusNotFound {
		t.Errorf("PATCH участника чужого проекта через свой projectId: код %d, ожидался 404", r.Status)
	}

	members := decode[[]memberJSON](t, get(t, alice.Token, "/projects/"+victim.ID+"/members").want(t, http.StatusOK))
	for _, m := range members {
		if m.User.ID == mallory.ID && m.AccessLevel != "view" {
			t.Fatalf("наблюдатель поднял себе доступ в чужом проекте: view → %s", m.AccessLevel)
		}
	}
}

// Та же дыра в DELETE: наблюдатель чужого проекта исключает из него любого участника.
func TestMemberDeleteIsScopedToProjectInURL(t *testing.T) {
	requireDB(t)
	alice := newUser(t, "Алиса")
	bob := newUser(t, "Боб")
	mallory := newUser(t, "Мэллори")
	victim := newProject(t, alice)
	bobInVictim := addMember(t, alice, victim.ID, bob, "edit")
	addMember(t, alice, victim.ID, mallory, "view")
	own := newProject(t, mallory)

	r := del(t, mallory.Token, "/projects/"+own.ID+"/members/"+bobInVictim.ID)
	if r.Status != http.StatusNotFound {
		t.Errorf("DELETE участника чужого проекта через свой projectId: код %d, ожидался 404", r.Status)
	}
	if r := get(t, bob.Token, "/projects/"+victim.ID); r.Status != http.StatusOK {
		t.Fatalf("Боба исключили из проекта чужими руками: на свой проект он теперь получает %d", r.Status)
	}
}

// ---------- Спринты и вехи ----------

func TestSprints(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Скрам-мастер")
	viewer := newUser(t, "Зритель")
	project := newProject(t, owner)
	addMember(t, owner, project.ID, viewer, "view")
	path := "/projects/" + project.ID + "/sprints"

	second := decode[sprintJSON](t, post(t, owner.Token, path, map[string]any{
		"name": "Спринт 2", "startDate": day(15), "endDate": day(28),
	}).want(t, http.StatusCreated))
	first := decode[sprintJSON](t, post(t, owner.Token, path, map[string]any{
		"name": "Спринт 1", "startDate": day(1), "endDate": day(14), "releaseVersion": "1.0",
	}).want(t, http.StatusCreated))

	list := decode[[]sprintJSON](t, get(t, viewer.Token, path).want(t, http.StatusOK))
	if len(list) != 2 || list[0].ID != first.ID || list[1].ID != second.ID {
		t.Fatalf("спринты = %+v, ожидался порядок по дате начала", list)
	}

	post(t, owner.Token, path, map[string]any{"name": "Кривой", "startDate": day(10), "endDate": day(1)}).want(t, http.StatusBadRequest)
	post(t, viewer.Token, path, map[string]any{"name": "Чужой", "startDate": day(1), "endDate": day(2)}).want(t, http.StatusForbidden)

	renamed := decode[sprintJSON](t, patch(t, owner.Token, "/sprints/"+first.ID, map[string]any{
		"name": "Спринт 1 · MVP", "startDate": day(1), "endDate": day(14),
	}).want(t, http.StatusOK))
	if renamed.Name != "Спринт 1 · MVP" {
		t.Fatalf("после PATCH название %q", renamed.Name)
	}

	// Удаление спринта отвязывает его задачи (ON DELETE SET NULL), но не удаляет их.
	task := newTask(t, owner.Token, project.ID, map[string]any{"sprintId": first.ID})
	del(t, owner.Token, "/sprints/"+first.ID).want(t, http.StatusNoContent)
	got := decode[taskJSON](t, get(t, owner.Token, "/tasks/"+task.ID).want(t, http.StatusOK))
	if got.SprintID != nil {
		t.Fatalf("после удаления спринта sprintId = %s, ожидался null", *got.SprintID)
	}
	del(t, owner.Token, "/sprints/"+first.ID).want(t, http.StatusNotFound)
}

func TestMilestoneStatusesAreDerivedFromDates(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Планировщик")
	project := newProject(t, owner)
	path := "/projects/" + project.ID + "/milestones"

	for _, body := range []map[string]any{
		{"code": "КТ-3", "name": "Сдача", "plannedDate": day(30)},
		{"code": "КТ-1", "name": "Старт", "plannedDate": day(-3)},
		{"code": "КТ-2", "name": "Демо", "plannedDate": day(5)},
	} {
		post(t, owner.Token, path, body).want(t, http.StatusCreated)
	}

	list := decode[[]milestoneJSON](t, get(t, owner.Token, path).want(t, http.StatusOK))
	if len(list) != 3 {
		t.Fatalf("вех %d, ожидалось 3", len(list))
	}
	// Ближайшая недостигнутая — current, последняя — final, между ними — planned.
	for i, want := range []string{"current", "planned", "final"} {
		if list[i].Status != want {
			t.Fatalf("веха %s: статус %q, ожидался %q", list[i].Code, list[i].Status, want)
		}
	}
	if r := list[0].RiskDays; r == nil || *r != 3 {
		t.Fatalf("у вехи, просроченной на 3 дня, riskDays = %s, ожидалось 3", intOrNull(r))
	}
	if r := list[1].RiskDays; r != nil {
		t.Fatalf("у будущей вехи riskDays = %s, ожидался null", intOrNull(r))
	}

	post(t, owner.Token, path, map[string]any{"name": "Без даты"}).want(t, http.StatusBadRequest)
	patch(t, owner.Token, "/milestones/"+list[1].ID, map[string]any{"name": "Демо заказчику", "plannedDate": day(6)}).want(t, http.StatusOK)
	del(t, owner.Token, "/milestones/"+list[1].ID).want(t, http.StatusNoContent)
	if left := decode[[]milestoneJSON](t, get(t, owner.Token, path).want(t, http.StatusOK)); len(left) != 2 {
		t.Fatalf("после удаления осталось %d вех, ожидалось 2", len(left))
	}
}

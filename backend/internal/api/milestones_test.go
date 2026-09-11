package api_test

import (
	"net/http"
	"testing"
)

// ---------- Вехи ----------

func TestMilestoneActualDateMarksAchieved(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Владелец")
	project := newProject(t, owner)
	milestone := newMilestone(t, owner.Token, project.ID, nil)

	if milestone.Status != "current" || milestone.ActualDate != nil {
		t.Fatalf("свежая веха = %+v, ожидались status=current и пустая actualDate", milestone)
	}

	achieved := decode[milestoneJSON](t, patch(t, owner.Token, "/milestones/"+milestone.ID, map[string]any{
		"code":        milestone.Code,
		"name":        milestone.Name,
		"plannedDate": milestone.PlannedDate,
		"actualDate":  day(0),
	}).want(t, http.StatusOK))
	if achieved.ActualDate == nil || achieved.Status != "done" {
		t.Fatalf("после actualDate = %+v, ожидались заполненная actualDate и status=done", achieved)
	}

	// Переименование без actualDate не должно сбрасывать отметку о достижении.
	renamed := decode[milestoneJSON](t, patch(t, owner.Token, "/milestones/"+milestone.ID, map[string]any{
		"code":        milestone.Code,
		"name":        "Переименованная веха",
		"plannedDate": milestone.PlannedDate,
	}).want(t, http.StatusOK))
	if renamed.ActualDate == nil || renamed.Status != "done" {
		t.Fatalf("после переименования = %+v, ожидалось сохранение actualDate/status=done", renamed)
	}

	// Явный null снимает отметку о достижении.
	reopened := decode[milestoneJSON](t, patch(t, owner.Token, "/milestones/"+milestone.ID, map[string]any{
		"code":        milestone.Code,
		"name":        renamed.Name,
		"plannedDate": milestone.PlannedDate,
		"actualDate":  nil,
	}).want(t, http.StatusOK))
	if reopened.ActualDate != nil || reopened.Status == "done" {
		t.Fatalf("после actualDate=null = %+v, ожидались пустая actualDate и status != done", reopened)
	}
}

func TestMilestoneTaskCounters(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Владелец")
	project := newProject(t, owner)
	milestone := newMilestone(t, owner.Token, project.ID, nil)

	fresh := listMilestones(t, owner.Token, project.ID)[0]
	if fresh.TasksTotal != 0 || fresh.TasksDone != 0 {
		t.Fatalf("без задач = %+v, ожидались нулевые счётчики", fresh)
	}

	t1 := newTask(t, owner.Token, project.ID, map[string]any{"milestoneId": milestone.ID})
	newTask(t, owner.Token, project.ID, map[string]any{"milestoneId": milestone.ID})
	patch(t, owner.Token, "/tasks/"+t1.ID, map[string]any{"status": "done"}).want(t, http.StatusOK)

	got := listMilestones(t, owner.Token, project.ID)[0]
	if got.TasksTotal != 2 || got.TasksDone != 1 {
		t.Fatalf("счётчики = %+v, ожидались total=2 done=1", got)
	}
}

func listMilestones(t *testing.T, token, projectID string) []milestoneJSON {
	t.Helper()
	return decode[[]milestoneJSON](t, get(t, token, "/projects/"+projectID+"/milestones").want(t, http.StatusOK))
}

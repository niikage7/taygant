package mcpserver

import (
	"strings"
	"testing"

	"taygant_backend/internal/models"
)

func TestProjectSummary(t *testing.T) {
	project := models.Project{Deadline: models.NewDate(2026, 10, 1)}
	tasks := taskBreakdown{Total: 10, Done: 4, InProgress: 3, Overdue: 2, Blocked: 1}

	cases := []struct {
		name     string
		tasks    taskBreakdown
		schedule scheduleOut
		parts    []string
	}{
		{"нет задач", taskBreakdown{}, scheduleOut{}, []string{"нет задач", "2026-10-01"}},
		{"отстаёт", tasks, scheduleOut{Status: "delayed", ForecastFinish: "2026-10-05", DeviationDays: 4},
			[]string{"отстаёт", "2026-10-05", "на 4 дн. позже", "Готово 4 из 10", "просрочено 2", "ждут других 1"}},
		{"под угрозой", tasks, scheduleOut{Status: "at_risk", ForecastFinish: "2026-09-29", BufferDays: 2},
			[]string{"успевает", "запас почти исчерпан: 2 дн."}},
		{"по плану", taskBreakdown{Total: 3, Done: 3}, scheduleOut{Status: "on_track", ForecastFinish: "2026-09-20", BufferDays: 11},
			[]string{"идёт по плану", "2026-09-20", "запас 11 дн.", "Готово 3 из 3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := projectSummary(project, tc.tasks, tc.schedule)
			for _, part := range tc.parts {
				if !strings.Contains(got, part) {
					t.Fatalf("в сводке нет %q: %q", part, got)
				}
			}
		})
	}
	// Про просрочки молчим, когда их нет: иначе «просрочено 0» пугает зря.
	if got := projectSummary(project, taskBreakdown{Total: 3, Done: 3}, scheduleOut{Status: "on_track"}); strings.Contains(got, "просрочено") {
		t.Fatalf("сводка без просрочек: %q", got)
	}
}

func TestRisksSummary(t *testing.T) {
	if got := risksSummary(nil); !strings.HasPrefix(got, "Ничего не горит") {
		t.Fatalf("без рисков: %q", got)
	}
	got := risksSummary([]riskItem{{Severity: "high"}, {Severity: "critical"}, {Severity: "high"}})
	if got != "Проблем: 3 (1 критичных, 2 серьёзных). Самые срочные — первыми в списке." {
		t.Fatalf("сводка рисков: %q", got)
	}
}

func TestDescribeChange(t *testing.T) {
	str := func(s string) *string { return &s }
	cases := []struct {
		entry models.TaskHistoryEntry
		want  string
	}{
		{models.TaskHistoryEntry{Action: models.HistoryActionCreated}, "создал(а) задачу"},
		{models.TaskHistoryEntry{Action: models.HistoryActionStatusChanged, Field: str("status"), OldValue: str("in_progress"), NewValue: str("done")},
			"статус: «В работе» → «Готово»"},
		{models.TaskHistoryEntry{Action: models.HistoryActionDateShifted, Field: str("endDate"), OldValue: str("2026-09-10"), NewValue: str("2026-09-12")},
			"окончание: 2026-09-10 → 2026-09-12"},
		{models.TaskHistoryEntry{Action: models.HistoryActionUpdated, Field: str("title")}, "изменил(а) название"},
		{models.TaskHistoryEntry{Action: models.HistoryActionCommented}, "оставил(а) комментарий"},
		{models.TaskHistoryEntry{Action: "something_new"}, "something_new"},
	}
	for _, tc := range cases {
		if got := describeChange(tc.entry); got != tc.want {
			t.Errorf("describeChange(%s) = %q, ожидалось %q", tc.entry.Action, got, tc.want)
		}
	}
}

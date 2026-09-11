package mcpserver

// Тесты чистой логики MCP-сервера — без базы, выполняются за миллисекунды.
// Сценарии целиком (HTTP, SDK-клиент, PostgreSQL) проверяет
// internal/api/mcp_test.go.

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"taygant_backend/internal/access"
	"taygant_backend/internal/models"
)

func TestNormalizeTaskCode(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"TASK-005", "TASK-005", true},
		{"task-5", "TASK-005", true},
		{"Task 12", "TASK-012", true},
		{" 7 ", "TASK-007", true},
		{"TASK_0042", "TASK-042", true},
		{"TASK-1234", "TASK-1234", true},
		{"Настроить CI", "", false},
		{"TASK-", "", false},
		{"PRJ-2026-A1B2", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		got, ok := normalizeTaskCode(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("normalizeTaskCode(%q) = %q, %v; ожидалось %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func membershipFor(name, code string, status models.ProjectStatus, created time.Time) access.Membership {
	project := models.Project{Name: name, Code: code, Status: status}
	project.ID = uuid.New()
	project.CreatedAt = created
	return access.Membership{Project: project}
}

func TestPickProject(t *testing.T) {
	now := time.Now()
	site := membershipFor("Сайт", "PRJ-2026-AAAA", models.ProjectStatusActive, now)
	site2 := membershipFor("Сайт 2.0", "PRJ-2026-BBBB", models.ProjectStatusActive, now)
	crm := membershipFor("CRM для отдела продаж", "PRJ-2026-CCCC", models.ProjectStatusActive, now)
	old := membershipFor("Старый портал", "PRJ-2025-DDDD", models.ProjectStatusArchived, now)
	all := []access.Membership{site, site2, crm, old}

	cases := []struct {
		name    string
		all     []access.Membership
		ref     string
		want    access.Membership
		errPart string
	}{
		{"по id", all, crm.Project.ID.String(), crm, ""},
		{"по коду без учёта регистра", all, "prj-2026-cccc", crm, ""},
		{"точное название важнее частичного", all, "сайт", site, ""},
		{"часть названия", all, "продаж", crm, ""},
		{"архивный можно назвать явно", all, "Старый портал", old, ""},
		{"без ссылки, когда активный проект один", []access.Membership{crm, old}, "", crm, ""},
		{"неоднозначная часть названия", all, "айт", access.Membership{}, "несколько проектов"},
		{"без ссылки при нескольких", all, "", access.Membership{}, "несколько проектов"},
		{"неизвестное название", all, "склад", access.Membership{}, "не найден"},
		{"чужой id", all, uuid.NewString(), access.Membership{}, "не найден"},
		{"нет проектов вовсе", nil, "", access.Membership{}, "ни в одном проекте"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := pickProject(tc.all, tc.ref)
			if tc.errPart != "" {
				if err == nil || !strings.Contains(err.Error(), tc.errPart) {
					t.Fatalf("ошибка %v, ожидалась с «%s»", err, tc.errPart)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Project.ID != tc.want.Project.ID {
				t.Fatalf("выбран «%s», ожидался «%s»", got.Project.Name, tc.want.Project.Name)
			}
		})
	}
}

// Сообщение о неоднозначности перечисляет варианты: по нему нейронка
// переспросит человека, а не будет гадать.
func TestPickProjectListsCandidates(t *testing.T) {
	a := membershipFor("Сайт", "PRJ-1", models.ProjectStatusActive, time.Now())
	b := membershipFor("Сайт 2.0", "PRJ-2", models.ProjectStatusActive, time.Now())
	_, err := pickProject([]access.Membership{a, b}, "")
	if err == nil || !strings.Contains(err.Error(), "«Сайт» (PRJ-1)") || !strings.Contains(err.Error(), "«Сайт 2.0» (PRJ-2)") {
		t.Fatalf("ошибка = %v", err)
	}
}

func TestEscapeLike(t *testing.T) {
	if got := escapeLike(`50%_\`); got != `50\%\_\\` {
		t.Fatalf("escapeLike = %q", got)
	}
}

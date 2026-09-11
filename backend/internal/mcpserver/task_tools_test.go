package mcpserver

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"taygant_backend/internal/models"
)

// toolRequest собирает вызов инструмента так, как его видит обработчик после
// разбора сообщения клиента версии 2026-07-28: версия и возможности клиента
// лежат прямо в _meta запроса.
func toolRequest(protocolVersion string, capabilities map[string]any) *mcp.CallToolRequest {
	meta := mcp.Meta{}
	if protocolVersion != "" {
		meta[mcp.MetaKeyProtocolVersion] = protocolVersion
	}
	if capabilities != nil {
		meta[mcp.MetaKeyClientCapabilities] = capabilities
	}
	return &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{Meta: meta}}
}

func TestCanAskConfirmation(t *testing.T) {
	cases := []struct {
		name         string
		version      string
		capabilities map[string]any
		want         bool
	}{
		{"новый протокол, elicitation без режимов — это форма", "2026-07-28", map[string]any{"elicitation": map[string]any{}}, true},
		{"новый протокол, явная форма", "2026-07-28", map[string]any{"elicitation": map[string]any{"form": map[string]any{}}}, true},
		{"новый протокол, только ссылки", "2026-07-28", map[string]any{"elicitation": map[string]any{"url": map[string]any{}}}, false},
		{"новый протокол без elicitation", "2026-07-28", map[string]any{}, false},
		{"прошлый протокол с elicitation", "2025-11-25", map[string]any{"elicitation": map[string]any{}}, false},
		{"клиент ничего не сообщил", "", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := canAskConfirmation(toolRequest(tc.version, tc.capabilities)); got != tc.want {
				t.Fatalf("canAskConfirmation = %v, ожидалось %v", got, tc.want)
			}
		})
	}
}

func TestConfirmationAnswerAppliesOnlyToTheSameTransition(t *testing.T) {
	taskID := uuid.New()
	asked := transitionState(taskID, models.TaskStatusInProgress, models.TaskStatusDone)

	answered := func(state string, responses mcp.InputResponseMap) (*mcp.ElicitResult, bool) {
		req := toolRequest("2026-07-28", nil)
		req.Params.RequestState = state
		req.Params.InputResponses = responses
		return confirmationAnswer(req, asked)
	}
	accept := mcp.InputResponseMap{confirmInputKey: &mcp.ElicitResult{Action: "accept"}}

	if answer, ok := answered(asked, accept); !ok || answer.Action != "accept" {
		t.Fatalf("ответ на тот же вопрос не принят: %v, %v", answer, ok)
	}
	if _, ok := answered(asked, nil); ok {
		t.Fatal("без ответа человека подтверждение засчитано")
	}
	// Пока человек думал, статус успел смениться: согласие давали на другое.
	changedMeanwhile := transitionState(taskID, models.TaskStatusPlanned, models.TaskStatusDone)
	if _, ok := answered(changedMeanwhile, accept); ok {
		t.Fatal("согласие на другую смену статуса засчитано")
	}
	if transitionState(uuid.New(), models.TaskStatusInProgress, models.TaskStatusDone) == asked {
		t.Fatal("отпечаток не зависит от задачи")
	}
}

func TestConfirmationMessageNamesTaskProjectAndStatuses(t *testing.T) {
	task := models.Task{Code: "TASK-005", Title: "Настроить CI", Status: models.TaskStatusOverdue}
	project := models.Project{Name: "Сайт"}

	msg := confirmationMessage(task, project, models.TaskStatusDone, "")
	for _, part := range []string{"TASK-005", "«Настроить CI»", "«Сайт»", "«Просрочена» → «Готово»"} {
		if !strings.Contains(msg, part) {
			t.Fatalf("в вопросе человеку нет %q: %q", part, msg)
		}
	}
	if withComment := confirmationMessage(task, project, models.TaskStatusDone, "PR #42 влит"); !strings.Contains(withComment, "«PR #42 влит»") {
		t.Fatalf("комментарий не показан человеку: %q", withComment)
	}
}

func TestStatusWarnings(t *testing.T) {
	today := models.Today()
	blocker := models.Task{Code: "TASK-001", Title: "Блокер", Status: models.TaskStatusInProgress}
	blocker.ID = uuid.New()
	late := models.Task{Code: "TASK-002", Title: "Опоздавшая", Status: models.TaskStatusOverdue, EndDate: today.AddDays(-1)}
	late.ID = uuid.New()
	fine := models.Task{Code: "TASK-003", Title: "В порядке", Status: models.TaskStatusPlanned, EndDate: today.AddDays(5)}
	fine.ID = uuid.New()
	b := &board{
		byID:  map[uuid.UUID]models.Task{blocker.ID: blocker, late.ID: late, fine.ID: fine},
		preds: map[uuid.UUID][]uuid.UUID{late.ID: {blocker.ID}},
	}
	checklist := []models.ChecklistItem{
		{Text: "Покрыто тестами"}, {Text: "Код-ревью", IsDone: true},
		{Text: "Документация"}, {Text: "Демо"}, {Text: "Метрики"},
	}

	started := strings.Join(statusWarnings(b, late, models.TaskStatusInProgress, checklist, today), " | ")
	if !strings.Contains(started, "TASK-001") || !strings.Contains(started, "просроченной") {
		t.Fatalf("взятие в работу: %q", started)
	}
	if strings.Contains(started, "чек-лист") {
		t.Fatalf("чек-лист важен только при закрытии: %q", started)
	}

	closed := strings.Join(statusWarnings(b, late, models.TaskStatusDone, checklist, today), " | ")
	if !strings.Contains(closed, "не отмечено 4 из 5") || !strings.Contains(closed, "закрыта раньше") {
		t.Fatalf("закрытие: %q", closed)
	}
	// В тексте не больше трёх пунктов — дальше это уже не предупреждение, а список.
	if strings.Contains(closed, "«Метрики»") {
		t.Fatalf("перечислено больше трёх пунктов чек-листа: %q", closed)
	}

	if warnings := statusWarnings(b, fine, models.TaskStatusInProgress, nil, today); len(warnings) != 0 {
		t.Fatalf("для задачи без проблем: %v", warnings)
	}
}

func TestTruncateRunesKeepsUTF8Valid(t *testing.T) {
	if got := truncateRunes("Привет, мир", 6); got != "Привет…" {
		t.Fatalf("truncateRunes = %q", got)
	}
	if got := truncateRunes("коротко", 100); got != "коротко" {
		t.Fatalf("короткий текст изменён: %q", got)
	}
	if got := truncateRunes(strings.Repeat("я", 10), 3); !utf8.ValidString(got) {
		t.Fatalf("обрезка сломала UTF-8: %q", got)
	}
}

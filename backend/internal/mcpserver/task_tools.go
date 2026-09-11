package mcpserver

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"taygant_backend/internal/access"
	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// taskRefInput — как инструменты принимают ссылку на задачу.
type taskRefInput struct {
	Task    string `json:"task" jsonschema:"задача: код (TASK-005), id или часть названия"`
	Project string `json:"project,omitempty" jsonschema:"проект: id, код или часть названия — нужен, если такой код задачи есть в нескольких ваших проектах"`
}

// ---------- get_task ----------

type checklistOut struct {
	Text string `json:"text"`
	Done bool   `json:"done"`
}

type linkOut struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	StatusLabel string `json:"statusLabel"`
	Rule        string `json:"rule" jsonschema:"как связаны задачи, например «начать после окончания»"`
	LagDays     int    `json:"lagDays,omitempty" jsonschema:"задержка между задачами в днях; отрицательная — опережение"`
}

type changeOut struct {
	At   string `json:"at"`
	Who  string `json:"who"`
	What string `json:"what"`
	Via  string `json:"via,omitempty" jsonschema:"заполнено, если правку сделала нейронка: чем она была подключена"`
}

type commentOut struct {
	At     string `json:"at"`
	Author string `json:"author"`
	Text   string `json:"text"`
}

type taskDetailOutput struct {
	Task            taskItem       `json:"task"`
	Description     string         `json:"description"`
	ActualStartDate string         `json:"actualStartDate,omitempty"`
	ActualEndDate   string         `json:"actualEndDate,omitempty"`
	BufferDays      int            `json:"bufferDays" jsonschema:"на сколько дней задачу можно задержать без сдвига срока проекта"`
	Checklist       []checklistOut `json:"checklist" jsonschema:"критерии готовности (Definition of Done)"`
	Predecessors    []linkOut      `json:"predecessors" jsonschema:"задачи, от которых зависит эта"`
	Successors      []linkOut      `json:"successors" jsonschema:"задачи, которые ждут эту"`
	RecentChanges   []changeOut    `json:"recentChanges" jsonschema:"последние изменения, новые первыми"`
	RecentComments  []commentOut   `json:"recentComments" jsonschema:"последние комментарии команды, новые первыми"`
	CanChangeStatus bool           `json:"canChangeStatus" jsonschema:"может ли владелец ключа менять статус этой задачи"`
	StatusHint      string         `json:"statusHint"`
}

// Ограничения размера ответа: описание и комментарии пишут люди, и
// случайная простыня текста не должна съедать контекст нейронки.
const (
	maxDescriptionRunes = 4000
	maxCommentRunes     = 1000
	recentChangesLimit  = 10
	recentCommentsLimit = 3
)

func (s *Server) getTask(ctx context.Context, req *mcp.CallToolRequest, in taskRefInput) (*mcp.CallToolResult, taskDetailOutput, error) {
	c, err := callerFrom(req)
	if err != nil {
		return nil, taskDetailOutput{}, err
	}
	stored, m, err := s.findTask(ctx, c.User.ID, in.Task, in.Project)
	if err != nil {
		return nil, taskDetailOutput{}, toolError(err)
	}
	b, err := s.loadBoard(ctx, m.Project)
	if err != nil {
		return nil, taskDetailOutput{}, toolError(err)
	}
	t, ok := b.byID[stored.ID]
	if !ok {
		return nil, taskDetailOutput{}, toolError(service.ErrNotFound)
	}

	checklist, err := s.checklist.List(ctx, t.ID)
	if err != nil {
		return nil, taskDetailOutput{}, toolError(err)
	}
	preds, succs, err := s.dependencies.List(ctx, t.ID)
	if err != nil {
		return nil, taskDetailOutput{}, toolError(err)
	}
	entries, err := s.history.List(ctx, t.ID)
	if err != nil {
		return nil, taskDetailOutput{}, toolError(err)
	}
	comments, err := s.comments.List(ctx, t.ID)
	if err != nil {
		return nil, taskDetailOutput{}, toolError(err)
	}

	today := models.Today()
	canChange, hint := statusPermission(m, t)
	out := taskDetailOutput{
		Task:            b.itemWithProject(t, today),
		Description:     truncateRunes(t.Description, maxDescriptionRunes),
		BufferDays:      t.BufferDays,
		Checklist:       make([]checklistOut, 0, len(checklist)),
		Predecessors:    make([]linkOut, 0, len(preds)),
		Successors:      make([]linkOut, 0, len(succs)),
		RecentChanges:   make([]changeOut, 0, recentChangesLimit),
		RecentComments:  make([]commentOut, 0, recentCommentsLimit),
		CanChangeStatus: canChange,
		StatusHint:      hint,
	}
	if t.ActualStartDate != nil {
		out.ActualStartDate = t.ActualStartDate.String()
	}
	if t.ActualEndDate != nil {
		out.ActualEndDate = t.ActualEndDate.String()
	}
	for _, item := range checklist {
		out.Checklist = append(out.Checklist, checklistOut{Text: item.Text, Done: item.IsDone})
	}
	for _, d := range preds {
		out.Predecessors = append(out.Predecessors, b.link(d, d.PredecessorTaskID))
	}
	for _, d := range succs {
		out.Successors = append(out.Successors, b.link(d, d.SuccessorTaskID))
	}
	for i := len(entries) - 1; i >= 0 && len(out.RecentChanges) < recentChangesLimit; i-- {
		out.RecentChanges = append(out.RecentChanges, newChangeOut(entries[i]))
	}
	for i := len(comments) - 1; i >= 0 && len(out.RecentComments) < recentCommentsLimit; i-- {
		out.RecentComments = append(out.RecentComments, newCommentOut(comments[i]))
	}
	return nil, out, nil
}

// link описывает связь с другой задачей того же проекта. Статус берётся с
// доски — эффективный, как на диаграмме, а не сохранённый.
func (b *board) link(d models.TaskDependency, otherID uuid.UUID) linkOut {
	other, ok := b.byID[otherID]
	if !ok {
		return linkOut{Code: "?", Title: "задача недоступна", Rule: dependencyTypeLabels[d.Type], LagDays: d.LagDays}
	}
	return linkOut{
		Code:        other.Code,
		Title:       other.Title,
		Status:      string(other.Status),
		StatusLabel: taskStatusLabel(other.Status),
		Rule:        dependencyTypeLabels[d.Type],
		LagDays:     d.LagDays,
	}
}

func newChangeOut(e models.TaskHistoryEntry) changeOut {
	out := changeOut{At: e.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"), What: describeChange(e)}
	if e.Actor != nil {
		out.Who = e.Actor.FullName
	}
	if e.Source == models.HistorySourceAssistant {
		out.Via = "через ассистента"
		if e.Via != "" {
			out.Via += " (" + e.Via + ")"
		}
	}
	return out
}

func newCommentOut(c models.TaskComment) commentOut {
	out := commentOut{At: c.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"), Text: truncateRunes(c.Text, maxCommentRunes)}
	if c.Author != nil {
		out.Author = c.Author.FullName
	}
	return out
}

// statusPermission повторяет правило прав из internal/access: статус меняет
// участник с полным доступом или исполнитель задачи с доступом edit.
func statusPermission(m access.Membership, t models.Task) (bool, string) {
	if m.CanEditTask(t) {
		return true, "Статус можно сменить инструментом set_task_status на planned, in_progress или done — только после явного согласия человека."
	}
	if m.Member.AccessLevel == models.AccessLevelView {
		return false, "У владельца ключа в этом проекте доступ только на просмотр: статус меняют исполнитель задачи или менеджер проекта."
	}
	return false, "Статус этой задачи меняют её исполнитель или менеджер проекта, а владелец ключа не исполнитель."
}

// truncateRunes обрезает текст по символам (не байтам: кириллица занимает
// два байта, и обрезка посреди символа сломала бы UTF-8).
func truncateRunes(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit]) + "…"
}

// ---------- set_task_status ----------

type setStatusInput struct {
	Task    string `json:"task" jsonschema:"задача: код (TASK-005), id или часть названия"`
	Status  string `json:"status" jsonschema:"новый статус: planned — запланирована, in_progress — в работе, done — готово"`
	Project string `json:"project,omitempty" jsonschema:"проект: id, код или часть названия — нужен, если такой код задачи есть в нескольких ваших проектах"`
	Comment string `json:"comment,omitempty" jsonschema:"короткое пояснение для команды, появится в комментариях задачи — например, «PR #42 влит в main»"`
}

type setStatusOutput struct {
	Changed        bool     `json:"changed" jsonschema:"изменился ли статус"`
	Task           taskItem `json:"task" jsonschema:"задача после вызова"`
	PreviousStatus string   `json:"previousStatus"`
	Message        string   `json:"message"`
	Warnings       []string `json:"warnings,omitempty" jsonschema:"на что обратить внимание человека"`
}

// userSettableStatuses — статусы, которые можно поставить руками. overdue и
// blocked система вычисляет сама (см. models.TaskStatus).
var userSettableStatuses = []any{
	string(models.TaskStatusPlanned),
	string(models.TaskStatusInProgress),
	string(models.TaskStatusDone),
}

// maxStatusCommentRunes — предел пояснения к смене статуса. Это заметка для
// команды, а не отчёт: длинный текст — скорее ошибка нейронки.
const maxStatusCommentRunes = 2000

func (s *Server) setTaskStatus(ctx context.Context, req *mcp.CallToolRequest, in setStatusInput) (*mcp.CallToolResult, setStatusOutput, error) {
	c, err := callerFrom(req)
	if err != nil {
		return nil, setStatusOutput{}, err
	}

	target := models.TaskStatus(in.Status)
	if target != models.TaskStatusPlanned && target != models.TaskStatusInProgress && target != models.TaskStatusDone {
		return nil, setStatusOutput{}, failf("статус «%s» поставить нельзя: доступны planned, in_progress и done, а overdue и blocked приложение вычисляет само", in.Status)
	}
	comment := strings.TrimSpace(in.Comment)
	if utf8.RuneCountInString(comment) > maxStatusCommentRunes {
		return nil, setStatusOutput{}, failf("комментарий длиннее %d символов — сократите его", maxStatusCommentRunes)
	}

	stored, m, err := s.findTask(ctx, c.User.ID, in.Task, in.Project)
	if err != nil {
		return nil, setStatusOutput{}, toolError(err)
	}
	if ok, hint := statusPermission(m, stored); !ok {
		return nil, setStatusOutput{}, failf("%s «%s»: статус не изменён. %s", stored.Code, stored.Title, hint)
	}

	b, err := s.loadBoard(ctx, m.Project)
	if err != nil {
		return nil, setStatusOutput{}, toolError(err)
	}
	current, ok := b.byID[stored.ID]
	if !ok {
		return nil, setStatusOutput{}, toolError(service.ErrNotFound)
	}
	today := models.Today()

	if stored.Status == target {
		msg := fmt.Sprintf("Статус %s «%s» уже «%s» — ничего не менял.", stored.Code, stored.Title, taskStatusLabel(target))
		if current.Status != stored.Status {
			msg += fmt.Sprintf(" На диаграмме задача сейчас «%s».", taskStatusLabel(current.Status))
		}
		return nil, setStatusOutput{
			Changed:        false,
			Task:           b.itemWithProject(current, today),
			PreviousStatus: string(current.Status),
			Message:        msg,
		}, nil
	}

	// Подтверждение от самого сервера — если клиент умеет его показать.
	// Иначе согласием служит окно разрешения на вызов инструмента, которое
	// клиент (Claude Code, Cursor) показывает перед вызовом.
	if canAskConfirmation(req) {
		state := transitionState(stored.ID, stored.Status, target)
		answer, answered := confirmationAnswer(req, state)
		if !answered {
			return &mcp.CallToolResult{
				InputRequests: mcp.InputRequestMap{confirmInputKey: &mcp.ElicitParams{
					Mode:            "form",
					Message:         confirmationMessage(current, m.Project, target, comment),
					RequestedSchema: &jsonschema.Schema{Type: "object", Properties: map[string]*jsonschema.Schema{}},
				}},
				RequestState: state,
			}, setStatusOutput{}, nil
		}
		if answer.Action != "accept" {
			return nil, setStatusOutput{
				Changed:        false,
				Task:           b.itemWithProject(current, today),
				PreviousStatus: string(current.Status),
				Message:        fmt.Sprintf("Человек не подтвердил смену статуса %s — ничего не изменено.", stored.Code),
			}, nil
		}
	}

	checklist, err := s.checklist.List(ctx, stored.ID)
	if err != nil {
		return nil, setStatusOutput{}, toolError(err)
	}
	warnings := statusWarnings(b, current, target, checklist, today)

	// Всё, что изменится ниже, журнал задачи запишет как «через ассистента».
	ctx = service.WithChangeSource(ctx, service.ChangeSource{Channel: models.HistorySourceAssistant, Via: c.TokenName})
	if _, err := s.tasks.Update(ctx, stored.ID, c.User.ID, service.TaskInput{Status: &target}); err != nil {
		return nil, setStatusOutput{}, toolError(err)
	}
	commented := false
	if comment != "" {
		if _, err := s.comments.Add(ctx, stored.ID, c.User.ID, comment); err != nil {
			// Статус уже сохранён — откатывать его из-за заметки неправильно.
			log.Printf("mcp: комментарий к %s: %v", stored.Code, err)
			warnings = append(warnings, "статус изменён, но комментарий сохранить не удалось — добавьте его в приложении")
		} else {
			commented = true
		}
	}

	after, err := s.loadBoard(ctx, m.Project)
	if err != nil {
		return nil, setStatusOutput{}, toolError(err)
	}
	updated := after.byID[stored.ID]

	msg := fmt.Sprintf("Статус %s «%s» изменён: «%s» → «%s». В истории задачи отмечено, что это сделано через ассистента.",
		stored.Code, stored.Title, taskStatusLabel(current.Status), taskStatusLabel(target))
	// Просрочку и блокировку приложение вычисляет само: поставленный статус
	// сохранён, но на диаграмме задача может выглядеть иначе.
	if updated.Status != target {
		msg += fmt.Sprintf(" На диаграмме задача пока «%s» — причина в предупреждениях.", taskStatusLabel(updated.Status))
	}
	if commented {
		msg += " Комментарий добавлен."
	}
	return nil, setStatusOutput{
		Changed:        true,
		Task:           after.itemWithProject(updated, today),
		PreviousStatus: string(current.Status),
		Message:        msg,
		Warnings:       warnings,
	}, nil
}

// statusWarnings — что человеку стоит знать о только что сделанной смене:
// приложение её примет, но картина на диаграмме может оказаться не той,
// которую он ждёт.
func statusWarnings(b *board, t models.Task, target models.TaskStatus, checklist []models.ChecklistItem, today models.Date) []string {
	var warnings []string
	waiting := b.blockedBy(t)

	if target == models.TaskStatusDone {
		var open []string
		for _, item := range checklist {
			if !item.IsDone {
				open = append(open, "«"+item.Text+"»")
			}
		}
		if len(open) > 0 {
			shown := open
			if len(shown) > 3 {
				shown = shown[:3]
			}
			warnings = append(warnings, fmt.Sprintf("в чек-листе готовности не отмечено %d из %d: %s",
				len(open), len(checklist), strings.Join(shown, ", ")))
		}
		if len(waiting) > 0 {
			warnings = append(warnings, "задача закрыта раньше тех, от которых зависит: "+strings.Join(waiting, "; "))
		}
		return warnings
	}

	if len(waiting) > 0 {
		warnings = append(warnings, "задача ждёт незавершённые: "+strings.Join(waiting, "; ")+
			" — пока их не закроют, на диаграмме она будет «Заблокирована»")
	}
	if t.EndDate.Before(today) {
		warnings = append(warnings, fmt.Sprintf("плановый срок окончания (%s) уже прошёл — задача остаётся просроченной; если срок сдвинулся, поправьте его в приложении",
			t.EndDate.String()))
	}
	return warnings
}

// ---------- подтверждение смены статуса ----------

const (
	// multiRoundTripVersion — первая версия MCP, в которой сервер может
	// попросить клиента задать человеку вопрос прямо в ответе на вызов
	// инструмента (SEP-2322, multi round-trip requests) — без открытого
	// соединения, то есть и в stateless-режиме, в котором работает сервер.
	multiRoundTripVersion = "2026-07-28"
	// confirmInputKey — метка запроса подтверждения; клиент возвращает ответ
	// человека под этим же ключом.
	confirmInputKey = "confirm"
)

// canAskConfirmation сообщает, что клиент сам покажет человеку окно
// подтверждения по просьбе сервера: говорит на версии протокола с
// многошаговыми запросами и объявил elicitation в режиме формы. Пустой
// объект elicitation по спецификации тоже означает поддержку формы.
func canAskConfirmation(req *mcp.CallToolRequest) bool {
	// Даты версий в формате YYYY-MM-DD сравниваются как строки.
	if req.ProtocolVersion() < multiRoundTripVersion {
		return false
	}
	caps := req.ClientCapabilities()
	if caps == nil || caps.Elicitation == nil {
		return false
	}
	return caps.Elicitation.Form != nil || caps.Elicitation.URL == nil
}

// transitionState — отпечаток смены, на которую спрашивали согласие. Клиент
// возвращает его вместе с ответом человека. Если статус задачи за это время
// успел поменяться, отпечаток не совпадёт и сервер спросит заново, а не
// применит согласие, данное на другую ситуацию.
func transitionState(taskID uuid.UUID, from, to models.TaskStatus) string {
	return fmt.Sprintf("set_task_status:%s:%s:%s", taskID, from, to)
}

// confirmationAnswer достаёт ответ человека, если он дан именно на эту смену.
func confirmationAnswer(req *mcp.CallToolRequest, state string) (*mcp.ElicitResult, bool) {
	if req.Params == nil || req.Params.RequestState != state {
		return nil, false
	}
	answer, ok := req.Params.InputResponses[confirmInputKey].(*mcp.ElicitResult)
	return answer, ok && answer != nil
}

// confirmationMessage — вопрос, который клиент покажет человеку. Статус «с
// какого» — тот, что человек видит на диаграмме, а не сохранённый.
func confirmationMessage(t models.Task, project models.Project, target models.TaskStatus, comment string) string {
	msg := fmt.Sprintf("Сменить статус задачи %s «%s» (проект «%s»): «%s» → «%s»?",
		t.Code, t.Title, project.Name, taskStatusLabel(t.Status), taskStatusLabel(target))
	if comment != "" {
		msg += fmt.Sprintf("\nКомментарий для команды: «%s»", comment)
	}
	return msg
}

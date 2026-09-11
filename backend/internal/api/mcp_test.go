package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---------- MCP: подключение нейронок (docs/mcp.md) ----------
//
// В /mcp ходим так же, как настоящие клиенты, двумя способами:
//   - SDK-клиентом на последней версии протокола — так будут подключаться
//     свежие версии Claude Code, Cursor и других редакторов;
//   - «сырым» JSON-RPC версии 2025-06-18 — так разговаривает большинство
//     клиентов сегодня. Заодно это проверка формата на проводе.

// mcpEndpoint — адрес для SDK-клиента. Хост условный: в сеть запросы не
// уходят, их принимает testApp через app.Test.
const mcpEndpoint = "http://taygant.test/mcp"

// legacyProtocol — версия протокола «сегодняшних» клиентов.
const legacyProtocol = "2025-06-18"

// appTransport отдаёт запросы SDK-клиента тестовому приложению напрямую и
// подставляет личный ключ — как его подставил бы редактор из настроек.
type appTransport struct{ secret string }

func (tr appTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	if tr.secret != "" {
		req.Header.Set("Authorization", "Bearer "+tr.secret)
	}
	return testApp.Test(req, -1)
}

// connectMCP подключает SDK-клиент от имени владельца ключа.
func connectMCP(t *testing.T, secret string, opts *mcp.ClientOptions) *mcp.ClientSession {
	t.Helper()
	client := mcp.NewClient(&mcp.Implementation{Name: "taygant-tests", Version: "1.0.0"}, opts)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{
		Endpoint:             mcpEndpoint,
		HTTPClient:           &http.Client{Transport: appTransport{secret: secret}},
		DisableStandaloneSSE: true,
	}, nil)
	if err != nil {
		t.Fatalf("подключение к MCP: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func callToolRaw(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s(%v): %v", name, args, err)
	}
	return res
}

// callTool вызывает инструмент и разбирает структурированный ответ в T.
// Отказ инструмента (isError) останавливает тест.
func callTool[T any](t *testing.T, session *mcp.ClientSession, name string, args map[string]any) T {
	t.Helper()
	res := callToolRaw(t, session, name, args)
	if res.IsError {
		t.Fatalf("%s(%v): инструмент отказал: %s", name, args, toolText(res))
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("%s: structuredContent: %v", name, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("%s: разбор ответа %s: %v", name, raw, err)
	}
	return out
}

// callToolError вызывает инструмент, который должен отказать, и возвращает
// текст отказа — тот, что увидит нейронка.
func callToolError(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) string {
	t.Helper()
	res := callToolRaw(t, session, name, args)
	if !res.IsError {
		t.Fatalf("%s(%v): ожидался отказ, получено %s", name, args, toolText(res))
	}
	return toolText(res)
}

func toolText(res *mcp.CallToolResult) string {
	var parts []string
	for _, c := range res.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, text.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// mcpPost отправляет JSON-RPC сообщение так, как это делает клиент версии
// 2025-06-18: POST с обоими типами в Accept и заголовком версии протокола.
func mcpPost(t *testing.T, secret, protocolVersion string, message any) response {
	t.Helper()
	body, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if protocolVersion != "" {
		req.Header.Set("Mcp-Protocol-Version", protocolVersion)
	}
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	return doMCP(t, req)
}

func doMCP(t *testing.T, req *http.Request) response {
	t.Helper()
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatalf("%s /mcp: %v", req.Method, err)
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("чтение ответа /mcp: %v", err)
	}
	return response{Status: resp.StatusCode, Body: payload}
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type rpcToolResult struct {
	IsError           bool            `json:"isError"`
	StructuredContent json.RawMessage `json:"structuredContent"`
	Content           []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// legacyCall вызывает инструмент по протоколу 2025-06-18 и возвращает result.
func legacyCall(t *testing.T, secret, tool string, args map[string]any) rpcToolResult {
	t.Helper()
	r := mcpPost(t, secret, legacyProtocol, map[string]any{
		"jsonrpc": "2.0", "id": uniq(), "method": "tools/call",
		"params": map[string]any{"name": tool, "arguments": args},
	}).want(t, http.StatusOK)
	rpc := decode[rpcResponse](t, r)
	if rpc.Error != nil {
		t.Fatalf("%s: JSON-RPC ошибка %d: %s", tool, rpc.Error.Code, rpc.Error.Message)
	}
	var result rpcToolResult
	if err := json.Unmarshal(rpc.Result, &result); err != nil {
		t.Fatalf("%s: разбор result %s: %v", tool, rpc.Result, err)
	}
	return result
}

// ---------- Ответы инструментов (свои структуры — это контракт с нейронкой) ----------

type mcpProjectRefJSON struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type mcpTaskJSON struct {
	ID              string             `json:"id"`
	Code            string             `json:"code"`
	Title           string             `json:"title"`
	Project         *mcpProjectRefJSON `json:"project"`
	Status          string             `json:"status"`
	StatusLabel     string             `json:"statusLabel"`
	StartDate       string             `json:"startDate"`
	EndDate         string             `json:"endDate"`
	DaysLeft        int                `json:"daysLeft"`
	OnCriticalPath  bool               `json:"onCriticalPath"`
	Assignee        string             `json:"assignee"`
	BlockedBy       []string           `json:"blockedBy"`
	ProgressPercent int                `json:"progressPercent"`
}

type mcpMyTasksJSON struct {
	Today   string `json:"today"`
	Summary struct {
		Open       int `json:"open"`
		Overdue    int `json:"overdue"`
		InProgress int `json:"inProgress"`
		Blocked    int `json:"blocked"`
		Planned    int `json:"planned"`
		DueSoon    int `json:"dueSoon"`
	} `json:"summary"`
	Tasks []mcpTaskJSON `json:"tasks"`
	Note  string        `json:"note"`
}

type mcpTaskDetailJSON struct {
	Task      mcpTaskJSON `json:"task"`
	Checklist []struct {
		Text string `json:"text"`
		Done bool   `json:"done"`
	} `json:"checklist"`
	Predecessors []struct {
		Code   string `json:"code"`
		Status string `json:"status"`
		Rule   string `json:"rule"`
	} `json:"predecessors"`
	Successors []struct {
		Code string `json:"code"`
	} `json:"successors"`
	RecentChanges []struct {
		Who  string `json:"who"`
		What string `json:"what"`
		Via  string `json:"via"`
	} `json:"recentChanges"`
	RecentComments []struct {
		Author string `json:"author"`
		Text   string `json:"text"`
	} `json:"recentComments"`
	CanChangeStatus bool   `json:"canChangeStatus"`
	StatusHint      string `json:"statusHint"`
}

type mcpSetStatusJSON struct {
	Changed        bool        `json:"changed"`
	Task           mcpTaskJSON `json:"task"`
	PreviousStatus string      `json:"previousStatus"`
	Message        string      `json:"message"`
	Warnings       []string    `json:"warnings"`
}

type mcpProjectStatusJSON struct {
	Project mcpProjectRefJSON `json:"project"`
	Summary string            `json:"summary"`
	Tasks   struct {
		Total      int `json:"total"`
		Done       int `json:"done"`
		InProgress int `json:"inProgress"`
		Planned    int `json:"planned"`
		Blocked    int `json:"blocked"`
		Overdue    int `json:"overdue"`
	} `json:"tasks"`
	Schedule struct {
		Status         string `json:"status"`
		ForecastFinish string `json:"forecastFinish"`
		DeviationDays  int    `json:"deviationDays"`
		BufferDays     int    `json:"bufferDays"`
	} `json:"schedule"`
	InProgress    []mcpTaskJSON `json:"inProgress"`
	StatusUpdates struct {
		PeriodDays   int `json:"periodDays"`
		Total        int `json:"total"`
		ViaAssistant int `json:"viaAssistant"`
	} `json:"statusUpdates"`
}

type mcpRisksJSON struct {
	Summary string `json:"summary"`
	Items   []struct {
		Severity string       `json:"severity"`
		Kind     string       `json:"kind"`
		Title    string       `json:"title"`
		Details  string       `json:"details"`
		Task     *mcpTaskJSON `json:"task"`
	} `json:"items"`
}

type mcpProjectsJSON struct {
	Projects []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Status      string `json:"status"`
		MyAccess    string `json:"myAccess"`
		MyOpenTasks int    `json:"myOpenTasks"`
	} `json:"projects"`
}

type statusChangesJSON struct {
	PeriodDays   int `json:"periodDays"`
	Total        int `json:"total"`
	ViaAssistant int `json:"viaAssistant"`
}

// ---------- Фикстуры ----------

// mcpTeam — проект с менеджером, разработчиком (edit) и наблюдателем (view);
// у разработчика есть личный ключ для ассистента.
type mcpTeam struct {
	manager, developer, viewer testUser
	project                    projectJSON
	devKey                     mcpTokenCreatedJSON
}

func newMCPTeam(t *testing.T) mcpTeam {
	t.Helper()
	team := mcpTeam{
		manager:   newUser(t, "Менеджер проекта"),
		developer: newUser(t, "Разработчик"),
		viewer:    newUser(t, "Продакт"),
	}
	team.project = newProject(t, team.manager)
	addMember(t, team.manager, team.project.ID, team.developer, "edit")
	addMember(t, team.manager, team.project.ID, team.viewer, "view")
	team.devKey = newMCPToken(t, team.developer, "Claude Code")
	return team
}

// setStatusREST меняет статус через приложение — «вручную».
func setStatusREST(t *testing.T, token, taskID, status string) {
	t.Helper()
	patch(t, token, "/tasks/"+taskID, map[string]any{"status": status}).want(t, http.StatusOK)
}

func taskStatus(t *testing.T, token, taskID string) string {
	t.Helper()
	return decode[taskJSON](t, get(t, token, "/tasks/"+taskID).want(t, http.StatusOK)).Status
}

func codes(tasks []mcpTaskJSON) []string {
	out := make([]string, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, task.Code+"/"+task.Status)
	}
	return out
}

// ---------- Доступ ----------

func TestMCPRequiresPersonalToken(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Владелец ключей")
	initialize := map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{
		"protocolVersion": legacyProtocol,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "1"},
	}}

	revoked := newMCPToken(t, user, "Отозванный")
	del(t, user.Token, mcpTokensPath+"/"+revoked.Token.ID).want(t, http.StatusNoContent)

	for name, secret := range map[string]string{
		"без ключа":                 "",
		"JWT браузера вместо ключа": user.Token,
		"выдуманный ключ":           "tgn_AAAAAAAAAAAAAAAAAAAAAAAAAA",
		"отозванный ключ":           revoked.Secret,
	} {
		t.Run(name, func(t *testing.T) {
			mcpPost(t, secret, "", initialize).want(t, http.StatusUnauthorized)
		})
	}

	valid := newMCPToken(t, user, "Рабочий")
	mcpPost(t, valid.Secret, "", initialize).want(t, http.StatusOK)

	// По lastUsedAt в приложении видно, какой ключ забыт и его пора отозвать.
	for _, token := range decode[[]mcpTokenJSON](t, get(t, user.Token, mcpTokensPath).want(t, http.StatusOK)) {
		if token.ID == valid.Token.ID && token.LastUsedAt == nil {
			t.Fatal("после подключения lastUsedAt ключа не заполнен")
		}
	}
}

func TestMCPRejectsCrossSiteBrowserRequests(t *testing.T) {
	requireDB(t)
	key := newMCPToken(t, newUser(t, "Жертва чужого сайта"), "Ключ")
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+key.Secret)
	// Так браузер помечает запрос, который страница evil.example шлёт на наш адрес.
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Origin", "https://evil.example")
	doMCP(t, req).want(t, http.StatusForbidden)
}

// ---------- Протокол глазами сегодняшних клиентов (2025-06-18) ----------

func TestMCPHandshakeForLegacyClients(t *testing.T) {
	requireDB(t)
	key := newMCPToken(t, newUser(t, "Пользователь Cursor"), "Cursor")

	r := mcpPost(t, key.Secret, "", map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{
		"protocolVersion": legacyProtocol,
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "cursor", "version": "1"},
	}}).want(t, http.StatusOK)
	var init struct {
		ProtocolVersion string `json:"protocolVersion"`
		Instructions    string `json:"instructions"`
		ServerInfo      struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
		Capabilities map[string]any `json:"capabilities"`
	}
	if err := json.Unmarshal(decode[rpcResponse](t, r).Result, &init); err != nil {
		t.Fatalf("разбор initialize: %v; тело %s", err, r.Body)
	}
	if init.ProtocolVersion != legacyProtocol || init.ServerInfo.Name != "taygant" {
		t.Fatalf("initialize: версия %q, сервер %q", init.ProtocolVersion, init.ServerInfo.Name)
	}
	if _, ok := init.Capabilities["tools"]; !ok {
		t.Fatalf("сервер не объявил инструменты: %v", init.Capabilities)
	}
	if !strings.Contains(init.Instructions, "без явного согласия") {
		t.Fatalf("инструкции для нейронки не содержат главного правила: %q", init.Instructions)
	}

	mcpPost(t, key.Secret, legacyProtocol, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"}).
		want(t, http.StatusAccepted)

	// Потока событий сервер не держит — клиенты спецификации это понимают по 405.
	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+key.Secret)
	resp, err := testApp.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed || resp.Header.Get("Allow") != http.MethodPost {
		t.Fatalf("GET /mcp: код %d, Allow %q — ожидались 405 и POST", resp.StatusCode, resp.Header.Get("Allow"))
	}

	list := decode[rpcResponse](t, mcpPost(t, key.Secret, legacyProtocol, map[string]any{
		"jsonrpc": "2.0", "id": 2, "method": "tools/list",
	}).want(t, http.StatusOK))
	var tools struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(list.Result, &tools); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range tools.Tools {
		names = append(names, tool.Name)
	}
	for _, want := range []string{"my_tasks", "get_task", "set_task_status", "project_status", "project_risks", "list_my_projects"} {
		if !contains(names, want) {
			t.Fatalf("в tools/list нет %s: %v", want, names)
		}
	}
}

func TestMCPToolAnnotationsTellClientsWhatChangesData(t *testing.T) {
	requireDB(t)
	session := connectMCP(t, newMCPToken(t, newUser(t, "Любопытный"), "Ключ").Secret, nil)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range res.Tools {
		if tool.Annotations == nil {
			t.Fatalf("%s без аннотаций: клиент не поймёт, меняет ли инструмент данные", tool.Name)
		}
		writes := tool.Name == "set_task_status"
		if tool.Annotations.ReadOnlyHint == writes {
			t.Fatalf("%s: readOnlyHint = %v", tool.Name, tool.Annotations.ReadOnlyHint)
		}
		if writes {
			schema, _ := json.Marshal(tool.InputSchema)
			if !strings.Contains(string(schema), `"enum":["planned","in_progress","done"]`) {
				t.Fatalf("схема set_task_status не ограничивает статусы: %s", schema)
			}
		}
	}
}

// ---------- «Что у меня сегодня?» ----------

func TestMCPMyTasksListsOnlyMyOpenTasksByUrgency(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	dev := team.developer
	path := team.project.ID

	inWork := newTask(t, team.manager.Token, path, map[string]any{"title": "В работе", "assigneeId": dev.ID, "startDate": day(-2), "endDate": day(2)})
	setStatusREST(t, dev.Token, inWork.ID, "in_progress")
	late := newTask(t, team.manager.Token, path, map[string]any{"title": "Сорвана", "assigneeId": dev.ID, "startDate": day(-10), "endDate": day(-3)})
	foreign := newTask(t, team.manager.Token, path, map[string]any{"title": "Чужая", "assigneeId": team.manager.ID})
	waiting := newTask(t, team.manager.Token, path, map[string]any{
		"title": "Ждёт чужую", "assigneeId": dev.ID, "startDate": day(11), "endDate": day(20),
		"predecessors": []map[string]any{{"taskId": foreign.ID, "type": "FS"}},
	})
	finished := newTask(t, team.manager.Token, path, map[string]any{"title": "Уже готова", "assigneeId": dev.ID, "status": "done"})

	// Второй проект: задачи из разных проектов собираются в один список.
	otherManager := newUser(t, "Другой менеджер")
	other := newProject(t, otherManager)
	addMember(t, otherManager, other.ID, dev, "edit")
	elsewhere := newTask(t, otherManager.Token, other.ID, map[string]any{"title": "В другом проекте", "assigneeId": dev.ID, "startDate": day(3), "endDate": day(8)})

	session := connectMCP(t, team.devKey.Secret, nil)
	got := callTool[mcpMyTasksJSON](t, session, "my_tasks", nil)

	want := []string{late.Code + "/overdue", inWork.Code + "/in_progress", waiting.Code + "/blocked", elsewhere.Code + "/planned"}
	if strings.Join(codes(got.Tasks), " ") != strings.Join(want, " ") {
		t.Fatalf("my_tasks = %v, ожидалось %v (просроченные → в работе → ждут → план; чужих и готовых нет)", codes(got.Tasks), want)
	}
	s := got.Summary
	if s.Open != 4 || s.Overdue != 1 || s.InProgress != 1 || s.Blocked != 1 || s.Planned != 1 || s.DueSoon != 1 {
		t.Fatalf("summary = %+v", s)
	}
	if got.Today != day(0) {
		t.Fatalf("today = %q, ожидалось %q", got.Today, day(0))
	}

	byTitle := map[string]mcpTaskJSON{}
	for _, task := range got.Tasks {
		byTitle[task.Title] = task
	}
	if w := byTitle["Ждёт чужую"]; len(w.BlockedBy) != 1 || !strings.Contains(w.BlockedBy[0], foreign.Code) {
		t.Fatalf("blockedBy = %v, ожидалась ссылка на %s", w.BlockedBy, foreign.Code)
	}
	if l := byTitle["Сорвана"]; l.DaysLeft != -3 {
		t.Fatalf("daysLeft у просроченной = %d, ожидалось -3", l.DaysLeft)
	}
	if e := byTitle["В другом проекте"]; e.Project == nil || e.Project.ID != other.ID {
		t.Fatalf("у задачи нет проекта, по которому её отличить: %+v", e.Project)
	}

	withDone := callTool[mcpMyTasksJSON](t, session, "my_tasks", map[string]any{"includeDone": true})
	if last := withDone.Tasks[len(withDone.Tasks)-1]; last.ID != finished.ID {
		t.Fatalf("с includeDone готовая задача должна быть в конце, последняя — %s", last.Code)
	}

	onlyOther := callTool[mcpMyTasksJSON](t, session, "my_tasks", map[string]any{"project": other.Code})
	if len(onlyOther.Tasks) != 1 || onlyOther.Tasks[0].ID != elsewhere.ID {
		t.Fatalf("фильтр по проекту вернул %v", codes(onlyOther.Tasks))
	}
}

// ---------- Карточка задачи и как ассистент на неё ссылается ----------

func TestMCPGetTaskFindsTaskByCodeTitleOrID(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	design := newTask(t, team.manager.Token, team.project.ID, map[string]any{"title": "Макет CI-панели", "assigneeId": team.developer.ID})
	setStatusREST(t, team.manager.Token, design.ID, "done")
	build := newTask(t, team.manager.Token, team.project.ID, map[string]any{
		"title": "Настроить CI", "assigneeId": team.developer.ID,
		"predecessors": []map[string]any{{"taskId": design.ID, "type": "FS"}},
	})
	post(t, team.manager.Token, "/tasks/"+build.ID+"/checklist", map[string]any{"text": "Сборка зелёная"}).want(t, http.StatusCreated)
	post(t, team.manager.Token, "/tasks/"+build.ID+"/comments", map[string]any{"text": "Раннер уже поднят"}).want(t, http.StatusCreated)

	// Во втором проекте разработчика тот же код TASK-001.
	otherManager := newUser(t, "Менеджер соседнего проекта")
	other := newProject(t, otherManager)
	addMember(t, otherManager, other.ID, team.developer, "edit")
	twin := newTask(t, otherManager.Token, other.ID, map[string]any{"title": "Написать тесты"})
	if twin.Code != design.Code {
		t.Fatalf("фикстура: коды %s и %s должны совпасть", twin.Code, design.Code)
	}

	session := connectMCP(t, team.devKey.Secret, nil)

	msg := callToolError(t, session, "get_task", map[string]any{"task": strings.ToLower(design.Code)})
	if !strings.Contains(msg, team.project.Name) || !strings.Contains(msg, other.Name) {
		t.Fatalf("неоднозначный код: ответ должен перечислить оба проекта, получено %q", msg)
	}
	if got := callTool[mcpTaskDetailJSON](t, session, "get_task", map[string]any{"task": "task-1", "project": other.Code}); got.Task.ID != twin.ID {
		t.Fatalf("с уточнением проекта найдена %s «%s»", got.Task.Code, got.Task.Title)
	}

	detail := callTool[mcpTaskDetailJSON](t, session, "get_task", map[string]any{"task": "настроить ci"})
	if detail.Task.ID != build.ID {
		t.Fatalf("поиск по названию нашёл %s «%s»", detail.Task.Code, detail.Task.Title)
	}
	if len(detail.Checklist) != 1 || detail.Checklist[0].Text != "Сборка зелёная" || detail.Checklist[0].Done {
		t.Fatalf("checklist = %+v", detail.Checklist)
	}
	if len(detail.Predecessors) != 1 || detail.Predecessors[0].Code != design.Code || detail.Predecessors[0].Status != "done" {
		t.Fatalf("predecessors = %+v", detail.Predecessors)
	}
	if len(detail.RecentComments) != 1 || detail.RecentComments[0].Text != "Раннер уже поднят" {
		t.Fatalf("recentComments = %+v", detail.RecentComments)
	}
	if len(detail.RecentChanges) == 0 || !detail.CanChangeStatus {
		t.Fatalf("recentChanges = %+v, canChangeStatus = %v", detail.RecentChanges, detail.CanChangeStatus)
	}
	if detail.Task.Project == nil || detail.Task.Project.ID != team.project.ID {
		t.Fatalf("в карточке нет проекта задачи: %+v", detail.Task.Project)
	}

	// Точное название не путается с частичным совпадением: «Настроить CI»
	// не должна становиться неоднозначной из-за «Макет CI-панели».
	if got := callTool[mcpTaskDetailJSON](t, session, "get_task", map[string]any{"task": "Настроить CI"}); got.Task.ID != build.ID {
		t.Fatalf("по точному названию найдена %s", got.Task.Code)
	}
	if got := callTool[mcpTaskDetailJSON](t, session, "get_task", map[string]any{"task": build.ID}); got.Task.ID != build.ID {
		t.Fatalf("по id найдена %s", got.Task.Code)
	}

	// Чужие проекты для ассистента не существуют: даже по точному id.
	outsiderUser := newUser(t, "Посторонний")
	newProject(t, outsiderUser)
	outsiderSession := connectMCP(t, newMCPToken(t, outsiderUser, "Ключ постороннего").Secret, nil)
	if msg := callToolError(t, outsiderSession, "get_task", map[string]any{"task": build.ID}); strings.Contains(msg, build.Title) {
		t.Fatalf("постороннему раскрыто название чужой задачи: %q", msg)
	}
}

// ---------- Смена статуса ----------

func TestMCPSetTaskStatusMarksHistoryAsAssistant(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	task := newTask(t, team.manager.Token, team.project.ID, map[string]any{"title": "Интеграция с оплатой", "assigneeId": team.developer.ID})

	// Клиент прошлой версии протокола: сервер не может сам спросить человека,
	// согласием служит окно разрешения на вызов инструмента в редакторе.
	result := legacyCall(t, team.devKey.Secret, "set_task_status", map[string]any{
		"task": task.Code, "status": "in_progress", "comment": "Начал, ветка feature/payments",
	})
	if result.IsError {
		t.Fatalf("set_task_status отказал: %+v", result.Content)
	}
	var got mcpSetStatusJSON
	if err := json.Unmarshal(result.StructuredContent, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Changed || got.Task.Status != "in_progress" || got.PreviousStatus != "planned" {
		t.Fatalf("ответ = %+v", got)
	}
	if !strings.Contains(got.Message, "через ассистента") {
		t.Fatalf("сообщение не говорит, как записана смена: %q", got.Message)
	}
	// Структурированный ответ продублирован текстом для клиентов, которые
	// structuredContent не читают.
	if len(result.Content) == 0 || !strings.Contains(result.Content[0].Text, `"changed":true`) {
		t.Fatalf("content = %+v", result.Content)
	}

	if status := taskStatus(t, team.manager.Token, task.ID); status != "in_progress" {
		t.Fatalf("статус в приложении = %q", status)
	}
	history := decode[[]historyJSON](t, get(t, team.manager.Token, "/tasks/"+task.ID+"/history").want(t, http.StatusOK))
	if len(history) != 3 {
		t.Fatalf("журнал: %+v, ожидались создание, смена статуса и комментарий", history)
	}
	created, changed, commented := history[0], history[1], history[2]
	if created.Source != "app" || created.Via != nil {
		t.Fatalf("создание в приложении записано как %q/%v", created.Source, created.Via)
	}
	if changed.Action != "status_changed" || changed.Source != "assistant" || changed.Via == nil || *changed.Via != "Claude Code" {
		t.Fatalf("смена статуса записана как %+v", changed)
	}
	if changed.Actor.ID != team.developer.ID || *changed.OldValue != "planned" || *changed.NewValue != "in_progress" {
		t.Fatalf("смена статуса: автор %s, %v → %v", changed.Actor.FullName, *changed.OldValue, *changed.NewValue)
	}
	if commented.Action != "commented" || commented.Source != "assistant" {
		t.Fatalf("комментарий записан как %+v", commented)
	}
	comments := decode[[]commentJSON](t, get(t, team.manager.Token, "/tasks/"+task.ID+"/comments").want(t, http.StatusOK))
	if len(comments) != 1 || comments[0].Text != "Начал, ветка feature/payments" || comments[0].Author.ID != team.developer.ID {
		t.Fatalf("комментарии = %+v", comments)
	}

	// Метрика из docs/mcp.md: доля смен статуса через ассистента.
	type dashboardJSON struct {
		StatusChanges statusChangesJSON `json:"statusChanges"`
	}
	stats := func() statusChangesJSON {
		t.Helper()
		return decode[dashboardJSON](t, get(t, team.manager.Token, "/projects/"+team.project.ID+"/dashboard").want(t, http.StatusOK)).StatusChanges
	}
	if s := stats(); s != (statusChangesJSON{PeriodDays: 30, Total: 1, ViaAssistant: 1}) {
		t.Fatalf("statusChanges = %+v", s)
	}
	setStatusREST(t, team.manager.Token, task.ID, "done")
	if s := stats(); s != (statusChangesJSON{PeriodDays: 30, Total: 2, ViaAssistant: 1}) {
		t.Fatalf("после ручной смены statusChanges = %+v", s)
	}

	// Повтор уже стоящего статуса ничего не меняет и в журнал не пишется.
	again := legacyCall(t, team.devKey.Secret, "set_task_status", map[string]any{"task": task.Code, "status": "done"})
	var repeat mcpSetStatusJSON
	if err := json.Unmarshal(again.StructuredContent, &repeat); err != nil {
		t.Fatal(err)
	}
	if again.IsError || repeat.Changed || !strings.Contains(repeat.Message, "уже") {
		t.Fatalf("повторная смена = %+v", repeat)
	}
	if s := stats(); s.Total != 2 {
		t.Fatalf("пустая смена попала в журнал: %+v", s)
	}
}

func TestMCPSetTaskStatusFollowsProjectRoles(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	own := newTask(t, team.manager.Token, team.project.ID, map[string]any{"title": "Своя", "assigneeId": team.developer.ID})
	managers := newTask(t, team.manager.Token, team.project.ID, map[string]any{"title": "Менеджерская", "assigneeId": team.manager.ID})

	viewerSession := connectMCP(t, newMCPToken(t, team.viewer, "Ключ продакта").Secret, nil)
	if msg := callToolError(t, viewerSession, "set_task_status", map[string]any{"task": own.Code, "status": "done"}); !strings.Contains(msg, "просмотр") {
		t.Fatalf("наблюдателю отказано без объяснения: %q", msg)
	}

	devSession := connectMCP(t, team.devKey.Secret, nil)
	if msg := callToolError(t, devSession, "set_task_status", map[string]any{"task": managers.Code, "status": "done"}); !strings.Contains(msg, "исполнитель") {
		t.Fatalf("отказ по чужой задаче без объяснения: %q", msg)
	}
	// overdue и blocked вычисляет система — схема инструмента их не пропускает.
	callToolError(t, devSession, "set_task_status", map[string]any{"task": own.Code, "status": "blocked"})
	callToolError(t, devSession, "set_task_status", map[string]any{"task": own.Code, "status": "done", "comment": strings.Repeat("я", 2001)})

	for _, task := range []taskJSON{own, managers} {
		if status := taskStatus(t, team.manager.Token, task.ID); status != "planned" {
			t.Fatalf("%s после отказов = %q", task.Code, status)
		}
	}

	// Менеджер с полным доступом меняет статус любой задачи проекта.
	managerSession := connectMCP(t, newMCPToken(t, team.manager, "Ключ менеджера").Secret, nil)
	callTool[mcpSetStatusJSON](t, managerSession, "set_task_status", map[string]any{"task": own.Code, "status": "in_progress"})
	if status := taskStatus(t, team.manager.Token, own.ID); status != "in_progress" {
		t.Fatalf("статус после смены менеджером = %q", status)
	}
}

func TestMCPSetTaskStatusWarnsAboutWhatStaysOnTheChart(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	blocker := newTask(t, team.manager.Token, team.project.ID, map[string]any{"title": "Блокер", "assigneeId": team.manager.ID})
	late := newTask(t, team.manager.Token, team.project.ID, map[string]any{
		"title": "Опоздавшая", "assigneeId": team.developer.ID, "startDate": day(-6), "endDate": day(-1),
		"predecessors": []map[string]any{{"taskId": blocker.ID, "type": "FS"}},
	})
	post(t, team.manager.Token, "/tasks/"+late.ID+"/checklist", map[string]any{"text": "Покрыто тестами"}).want(t, http.StatusCreated)

	session := connectMCP(t, team.devKey.Secret, nil)
	started := callTool[mcpSetStatusJSON](t, session, "set_task_status", map[string]any{"task": late.Code, "status": "in_progress"})
	joined := strings.Join(started.Warnings, " | ")
	if !strings.Contains(joined, blocker.Code) || !strings.Contains(joined, "просроченной") {
		t.Fatalf("предупреждения при взятии в работу = %v", started.Warnings)
	}
	// Статус сохранён, но на диаграмме задача остаётся просроченной — ответ
	// говорит об этом прямо, а не только «изменён на В работе».
	if started.Task.Status != "overdue" || !strings.Contains(started.Message, "На диаграмме задача пока «Просрочена»") {
		t.Fatalf("ответ при взятии в работу: статус %q, сообщение %q", started.Task.Status, started.Message)
	}

	closed := callTool[mcpSetStatusJSON](t, session, "set_task_status", map[string]any{"task": late.Code, "status": "done"})
	joined = strings.Join(closed.Warnings, " | ")
	if !strings.Contains(joined, "Покрыто тестами") || !strings.Contains(joined, blocker.Code) {
		t.Fatalf("предупреждения при закрытии = %v", closed.Warnings)
	}
}

// elicitationRecorder — «человек» перед экраном клиента: отвечает на вопросы
// сервера заданным действием и запоминает, о чём его спросили.
type elicitationRecorder struct {
	mu       sync.Mutex
	action   string
	messages []string
}

func (r *elicitationRecorder) handle(_ context.Context, req *mcp.ElicitRequest) (*mcp.ElicitResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages = append(r.messages, req.Params.Message)
	return &mcp.ElicitResult{Action: r.action}, nil
}

func (r *elicitationRecorder) asked() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.messages...)
}

func TestMCPAsksHumanToConfirmWhenClientSupportsIt(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	task := newTask(t, team.manager.Token, team.project.ID, map[string]any{"title": "Выгрузка отчёта", "assigneeId": team.developer.ID})

	// Человек отказался — статус не трогаем.
	refuser := &elicitationRecorder{action: "decline"}
	declined := callTool[mcpSetStatusJSON](t, connectMCP(t, team.devKey.Secret, &mcp.ClientOptions{ElicitationHandler: refuser.handle}),
		"set_task_status", map[string]any{"task": task.Code, "status": "done"})
	if declined.Changed || !strings.Contains(declined.Message, "не подтвердил") {
		t.Fatalf("после отказа человека: %+v", declined)
	}
	if status := taskStatus(t, team.manager.Token, task.ID); status != "planned" {
		t.Fatalf("статус изменён без согласия человека: %q", status)
	}
	asked := refuser.asked()
	if len(asked) != 1 || !strings.Contains(asked[0], task.Code) || !strings.Contains(asked[0], "«План» → «Готово»") {
		t.Fatalf("вопрос человеку = %q", asked)
	}

	// Человек согласился — статус меняется, в журнале «через ассистента».
	approver := &elicitationRecorder{action: "accept"}
	approverSession := connectMCP(t, team.devKey.Secret, &mcp.ClientOptions{ElicitationHandler: approver.handle})
	accepted := callTool[mcpSetStatusJSON](t, approverSession, "set_task_status", map[string]any{"task": task.Code, "status": "done"})
	if !accepted.Changed || taskStatus(t, team.manager.Token, task.ID) != "done" {
		t.Fatalf("после согласия человека: %+v", accepted)
	}
	history := decode[[]historyJSON](t, get(t, team.manager.Token, "/tasks/"+task.ID+"/history").want(t, http.StatusOK))
	if last := history[len(history)-1]; last.Action != "status_changed" || last.Source != "assistant" {
		t.Fatalf("последняя запись журнала = %+v", last)
	}

	// Если менять нечего, человека не беспокоим.
	callTool[mcpSetStatusJSON](t, approverSession, "set_task_status", map[string]any{"task": task.Code, "status": "done"})
	if n := len(approver.asked()); n != 1 {
		t.Fatalf("человека спросили %d раз(а), ожидался один вопрос", n)
	}
}

// ---------- «Как дела у проекта?» и «Что горит?» ----------

func TestMCPProjectStatusAndRisks(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	m, dev, id := team.manager.Token, team.developer.ID, team.project.ID

	newTask(t, m, id, map[string]any{"title": "Сделано", "status": "done", "startDate": day(-20), "endDate": day(-15)})
	overdue := newTask(t, m, id, map[string]any{"title": "Просроченная", "assigneeId": dev, "startDate": day(-10), "endDate": day(-2)})
	lateStart := newTask(t, m, id, map[string]any{"title": "Не начатая", "startDate": day(-3), "endDate": day(6)})
	blocked := newTask(t, m, id, map[string]any{
		"title": "Застрявшая", "startDate": day(-1), "endDate": day(8),
		"predecessors": []map[string]any{{"taskId": lateStart.ID, "type": "FS"}},
	})
	inWork := newTask(t, m, id, map[string]any{"title": "Идёт", "assigneeId": dev, "startDate": day(0), "endDate": day(4)})
	setStatusREST(t, m, inWork.ID, "in_progress")

	session := connectMCP(t, newMCPToken(t, team.viewer, "Ключ продакта").Secret, nil)

	status := callTool[mcpProjectStatusJSON](t, session, "project_status", nil)
	tasks := status.Tasks
	if tasks.Total != 5 || tasks.Done != 1 || tasks.InProgress != 1 || tasks.Planned != 1 || tasks.Blocked != 1 || tasks.Overdue != 1 {
		t.Fatalf("разбивка задач = %+v", tasks)
	}
	if status.Project.ID != id || status.Summary == "" || status.Schedule.ForecastFinish == "" {
		t.Fatalf("сводка = %+v", status)
	}
	// Задачи заканчиваются задолго до дедлайна (90 дней): запас реальный, а не ноль.
	if status.Schedule.Status != "on_track" || status.Schedule.BufferDays < 30 {
		t.Fatalf("schedule = %+v, ожидался on_track с запасом больше месяца", status.Schedule)
	}
	if len(status.InProgress) != 1 || status.InProgress[0].ID != inWork.ID || status.InProgress[0].Project != nil {
		t.Fatalf("inProgress = %+v (проект в ответе о проекте не повторяется)", status.InProgress)
	}
	if status.StatusUpdates.PeriodDays != 30 || status.StatusUpdates.Total != 1 || status.StatusUpdates.ViaAssistant != 0 {
		t.Fatalf("statusUpdates = %+v", status.StatusUpdates)
	}

	risks := callTool[mcpRisksJSON](t, session, "project_risks", map[string]any{"project": team.project.Name})
	kinds := map[string]string{}
	for i, item := range risks.Items {
		if item.Task != nil {
			kinds[item.Task.ID] = item.Kind
		}
		if i > 0 && severityOrder(item.Severity) < severityOrder(risks.Items[i-1].Severity) {
			t.Fatalf("риски не отсортированы по серьёзности: %+v", risks.Items)
		}
	}
	if kinds[overdue.ID] != "overdue" || kinds[lateStart.ID] != "late_start" || kinds[blocked.ID] != "blocked" {
		t.Fatalf("виды рисков по задачам = %v", kinds)
	}
	if _, ok := kinds[inWork.ID]; ok {
		t.Fatal("задача, которая идёт по плану, попала в «что горит»")
	}
	if !strings.Contains(risks.Summary, "Проблем: 3") {
		t.Fatalf("summary = %q", risks.Summary)
	}
}

func severityOrder(s string) int {
	return map[string]int{"critical": 0, "high": 1, "medium": 2}[s]
}

func TestMCPProjectToolsNeedProjectWhenThereAreSeveral(t *testing.T) {
	requireDB(t)
	manager := newUser(t, "Менеджер двух проектов")
	first := newProject(t, manager)
	second := newProject(t, manager)
	session := connectMCP(t, newMCPToken(t, manager, "Ключ").Secret, nil)

	msg := callToolError(t, session, "project_status", nil)
	if !strings.Contains(msg, first.Name) || !strings.Contains(msg, second.Name) {
		t.Fatalf("без проекта ответ должен перечислить варианты: %q", msg)
	}
	empty := callTool[mcpProjectStatusJSON](t, session, "project_status", map[string]any{"project": second.Code})
	if empty.Project.ID != second.ID || !strings.Contains(empty.Summary, "нет задач") {
		t.Fatalf("пустой проект: %+v", empty)
	}
	if msg := callToolError(t, session, "project_risks", map[string]any{"project": "нет такого"}); !strings.Contains(msg, "не найден") {
		t.Fatalf("неизвестный проект: %q", msg)
	}
}

func TestMCPListMyProjects(t *testing.T) {
	requireDB(t)
	team := newMCPTeam(t)
	newTask(t, team.manager.Token, team.project.ID, map[string]any{"assigneeId": team.developer.ID})
	own := newProject(t, team.developer)
	del(t, team.developer.Token, "/projects/"+own.ID).want(t, http.StatusNoContent) // в архив

	session := connectMCP(t, team.devKey.Secret, nil)
	active := callTool[mcpProjectsJSON](t, session, "list_my_projects", nil)
	if len(active.Projects) != 1 || active.Projects[0].ID != team.project.ID {
		t.Fatalf("без архивных: %+v", active.Projects)
	}
	if p := active.Projects[0]; p.MyAccess != "edit" || p.MyOpenTasks != 1 {
		t.Fatalf("проект команды: %+v", p)
	}

	all := callTool[mcpProjectsJSON](t, session, "list_my_projects", map[string]any{"includeArchived": true})
	statuses := map[string]string{}
	for _, p := range all.Projects {
		statuses[p.ID] = p.Status + "/" + p.MyAccess
	}
	if statuses[own.ID] != "archived/full" || len(statuses) != 2 {
		t.Fatalf("с архивными: %v", statuses)
	}
}

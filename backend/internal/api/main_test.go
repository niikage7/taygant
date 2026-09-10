// Интеграционные тесты HTTP API: настоящее Fiber-приложение (server.New)
// поверх настоящего PostgreSQL. Запросы идут через app.Test — без сети и порта,
// но через те же middleware, маршруты, сервисы и SQL, что и в проде.
//
// Нужна база. Строка подключения берётся из TEST_DATABASE_URL (формат URL);
// тесты создают в этой базе отдельную временную схему, прогоняют миграции
// и в конце удаляют схему — данные разработки не затрагиваются. Без переменной
// тесты этого пакета пропускаются, и `go test ./...` работает без PostgreSQL.
//
// Запуск против БД из docker compose (из каталога backend):
//
//	docker compose up -d postgres
//	TEST_DATABASE_URL="postgres://taygant:taygant@localhost:5432/taygant?sslmode=disable" go test ./...
package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"taygant_backend/internal/config"
	"taygant_backend/internal/database"
	"taygant_backend/internal/models"
	"taygant_backend/internal/server"
)

// testApp — приложение, собранное так же, как в main.go. nil, если БД не задана.
var testApp *fiber.App

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

// runTests вынесен из TestMain, чтобы defer'ы (удаление схемы) успели
// выполниться до os.Exit.
func runTests(m *testing.M) int {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL не задан — интеграционные тесты API пропущены")
		return m.Run()
	}

	schema := fmt.Sprintf("api_test_%d", time.Now().UnixNano())
	schemaDSN, err := withSearchPath(dsn, schema)
	if err != nil {
		fmt.Println("TEST_DATABASE_URL:", err)
		return 1
	}

	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		fmt.Println("подключение к тестовой БД:", err)
		return 1
	}
	adminSQL, _ := admin.DB()
	defer adminSQL.Close()

	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		fmt.Println("создание тестовой схемы:", err)
		return 1
	}
	defer admin.Exec("DROP SCHEMA " + schema + " CASCADE")

	db, err := database.Connect(schemaDSN, 10*time.Second)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	appSQL, _ := db.DB()
	defer appSQL.Close()

	if err := database.Migrate(db); err != nil {
		fmt.Println(err)
		return 1
	}

	// Часть тестов намеренно проверяет ошибочные запросы. Логи GORM и Fiber
	// превратили бы вывод упавшего теста в сотни строк чужого шума.
	db.Logger = logger.Discard
	restoreStdout := silenceStdout()

	testApp = server.New(config.Config{
		AllowedOrigins: []string{"http://localhost:3000"},
		// У всех запросов app.Test один и тот же IP, лимитер здесь только мешал бы.
		RateLimitMax:    1_000_000,
		RateLimitWindow: time.Minute,
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	}, db)
	restoreStdout()

	return m.Run()
}

// silenceStdout подменяет os.Stdout на /dev/null и возвращает функцию отката.
// Логгер запросов Fiber запоминает os.Stdout в момент server.New, поэтому
// подмены на время сборки приложения хватает, чтобы заглушить его насовсем.
func silenceStdout() (restore func()) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return func() {}
	}
	original := os.Stdout
	os.Stdout = devNull
	return func() { os.Stdout = original }
}

// withSearchPath направляет все соединения приложения во временную схему:
// pgx передаёт неизвестные параметры URL серверу как настройки сессии.
func withSearchPath(dsn, schema string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return "", fmt.Errorf("ожидается URL вида postgres://user:pass@host:port/db, получено %q", dsn)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// requireDB пропускает тест, если PostgreSQL для интеграционных тестов не задан.
func requireDB(t *testing.T) {
	t.Helper()
	if testApp == nil {
		t.Skip("нужен PostgreSQL: задайте TEST_DATABASE_URL (см. комментарий в main_test.go)")
	}
}

// ---------- HTTP ----------

// response — ответ API. Тело хранится целиком, чтобы сообщение об ошибке
// показывало, что именно вернул сервер.
type response struct {
	Status int
	Body   []byte
}

// want останавливает тест, если код ответа не совпал с ожидаемым.
func (r response) want(t *testing.T, status int) response {
	t.Helper()
	if r.Status != status {
		t.Fatalf("ожидался код %d, получен %d: %s", status, r.Status, r.Body)
	}
	return r
}

func get(t *testing.T, token, path string) response {
	t.Helper()
	return send(t, http.MethodGet, path, token, nil)
}

func post(t *testing.T, token, path string, body any) response {
	t.Helper()
	return send(t, http.MethodPost, path, token, body)
}

func patch(t *testing.T, token, path string, body any) response {
	t.Helper()
	return send(t, http.MethodPatch, path, token, body)
}

func del(t *testing.T, token, path string) response {
	t.Helper()
	return send(t, http.MethodDelete, path, token, nil)
}

// send кодирует body в JSON и выполняет запрос к /api/v1 + path.
func send(t *testing.T, method, path, token string, body any) response {
	t.Helper()
	var raw string
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		raw = string(encoded)
	}
	return sendRaw(t, method, path, token, raw)
}

// sendRaw отправляет тело как есть — для проверки реакции на битый JSON.
func sendRaw(t *testing.T, method, path, token, body string) response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, "/api/v1"+path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := testApp.Test(req, 10_000)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: чтение ответа: %v", method, path, err)
	}
	return response{Status: resp.StatusCode, Body: payload}
}

// newRawRequest собирает запрос с произвольным заголовком Authorization —
// для проверок, которым нужен заведомо некорректный заголовок.
func newRawRequest(method, path, authorization string) *http.Request {
	req := httptest.NewRequest(method, "/api/v1"+path, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	return req
}

// urlQuery экранирует значение query-параметра (кириллица, %, пробелы).
func urlQuery(s string) string { return url.QueryEscape(s) }

// decode разбирает JSON-ответ в T.
func decode[T any](t *testing.T, r response) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(r.Body, &v); err != nil {
		t.Fatalf("не удалось разобрать ответ %s: %v", r.Body, err)
	}
	return v
}

// ---------- Контракт ответов ----------
//
// Тесты разбирают ответы в собственные структуры, а не в DTO пакета api:
// так переименование JSON-поля в DTO (которое сломает фронт) сломает и тесты.

type errorJSON struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

type userJSON struct {
	ID         string `json:"id"`
	FullName   string `json:"fullName"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

type authJSON struct {
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	User         userJSON `json:"user"`
}

type projectJSON struct {
	ID                   string    `json:"id"`
	Code                 string    `json:"code"`
	Name                 string    `json:"name"`
	Status               string    `json:"status"`
	Phase                string    `json:"phase"`
	Description          string    `json:"description"`
	StartDate            string    `json:"startDate"`
	Deadline             string    `json:"deadline"`
	WorkingCalendarType  string    `json:"workingCalendarType"`
	ProgressPercent      float64   `json:"progressPercent"`
	HealthIndex          int       `json:"healthIndex"`
	CriticalRisksCount   int       `json:"criticalRisksCount"`
	DurationCalendarDays int       `json:"durationCalendarDays"`
	CreatedBy            *userJSON `json:"createdBy"`
}

type memberJSON struct {
	ID          string   `json:"id"`
	User        userJSON `json:"user"`
	ProjectRole string   `json:"projectRole"`
	SprintRole  string   `json:"sprintRole"`
	AccessLevel string   `json:"accessLevel"`
}

type taskJSON struct {
	ID                        string    `json:"id"`
	ProjectID                 string    `json:"projectId"`
	SprintID                  *string   `json:"sprintId"`
	ParentTaskID              *string   `json:"parentTaskId"`
	MilestoneID               *string   `json:"milestoneId"`
	Code                      string    `json:"code"`
	WBSNumber                 string    `json:"wbsNumber"`
	Title                     string    `json:"title"`
	Assignee                  *userJSON `json:"assignee"`
	Status                    string    `json:"status"`
	StartDate                 string    `json:"startDate"`
	EndDate                   string    `json:"endDate"`
	DurationCalendarDays      int       `json:"durationCalendarDays"`
	DurationWorkingDays       int       `json:"durationWorkingDays"`
	ProgressPercent           int       `json:"progressPercent"`
	WeightPercent             float64   `json:"weightPercent"`
	PlanVsActualDeviationDays int       `json:"planVsActualDeviationDays"`
}

type taskDetailJSON struct {
	taskJSON
	Description   string           `json:"description"`
	Checklist     []checklistJSON  `json:"checklist"`
	Predecessors  []dependencyJSON `json:"predecessors"`
	Successors    []dependencyJSON `json:"successors"`
	CommentsCount int              `json:"commentsCount"`
}

type dependencyJSON struct {
	ID                string `json:"id"`
	PredecessorTaskID string `json:"predecessorTaskId"`
	SuccessorTaskID   string `json:"successorTaskId"`
	PredecessorTitle  string `json:"predecessorTitle"`
	SuccessorTitle    string `json:"successorTitle"`
	Type              string `json:"type"`
	LagDays           int    `json:"lagDays"`
}

type dependencyListJSON struct {
	Predecessors []dependencyJSON `json:"predecessors"`
	Successors   []dependencyJSON `json:"successors"`
}

type checklistJSON struct {
	ID     string `json:"id"`
	TaskID string `json:"taskId"`
	Text   string `json:"text"`
	IsDone bool   `json:"isDone"`
}

type commentJSON struct {
	ID     string   `json:"id"`
	TaskID string   `json:"taskId"`
	Author userJSON `json:"author"`
	Text   string   `json:"text"`
}

type historyJSON struct {
	Action   string   `json:"action"`
	Field    *string  `json:"field"`
	OldValue *string  `json:"oldValue"`
	NewValue *string  `json:"newValue"`
	Actor    userJSON `json:"actor"`
}

type sprintJSON struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type milestoneJSON struct {
	ID          string  `json:"id"`
	ProjectID   string  `json:"projectId"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	PlannedDate string  `json:"plannedDate"`
	ActualDate  *string `json:"actualDate"`
	Status      string  `json:"status"`
	RiskDays    *int    `json:"riskDays"`
	TasksTotal  int     `json:"tasksTotal"`
	TasksDone   int     `json:"tasksDone"`
}

// ---------- Фикстуры ----------

// testUser — зарегистрированный пользователь с действующими токенами.
type testUser struct {
	ID           string
	Email        string
	Token        string
	RefreshToken string
}

const testPassword = "password1"

// seq делает email'ы и названия уникальными в пределах прогона, в том числе
// при `go test -count=N`, когда тесты повторяются в одной и той же схеме.
var seq atomic.Int64

func uniq() int64 { return seq.Add(1) }

// newUser регистрирует пользователя через API.
func newUser(t *testing.T, name string) testUser {
	t.Helper()
	email := fmt.Sprintf("user%d@test.dev", uniq())
	r := post(t, "", "/auth/register", map[string]any{
		"email":    email,
		"password": testPassword,
		"fullName": name,
	}).want(t, http.StatusCreated)
	auth := decode[authJSON](t, r)
	return testUser{ID: auth.User.ID, Email: email, Token: auth.AccessToken, RefreshToken: auth.RefreshToken}
}

// newProject создаёт проект на 90 дней от сегодняшнего; владелец получает полный доступ.
func newProject(t *testing.T, owner testUser) projectJSON {
	t.Helper()
	r := post(t, owner.Token, "/projects", map[string]any{
		"name":      fmt.Sprintf("Проект %d", uniq()),
		"startDate": day(0),
		"deadline":  day(90),
	}).want(t, http.StatusCreated)
	return decode[projectJSON](t, r)
}

// addMember добавляет пользователя в команду проекта с указанным уровнем доступа.
func addMember(t *testing.T, owner testUser, projectID string, user testUser, accessLevel string) memberJSON {
	t.Helper()
	r := post(t, owner.Token, "/projects/"+projectID+"/members", map[string]any{
		"userId":      user.ID,
		"projectRole": "developer",
		"accessLevel": accessLevel,
	}).want(t, http.StatusCreated)
	return decode[memberJSON](t, r)
}

// newTask создаёт задачу в будущем (дни 1–10); fields дополняют и переопределяют тело запроса.
func newTask(t *testing.T, token, projectID string, fields map[string]any) taskJSON {
	t.Helper()
	body := map[string]any{"title": fmt.Sprintf("Задача %d", uniq()), "startDate": day(1), "endDate": day(10)}
	for k, v := range fields {
		body[k] = v
	}
	r := post(t, token, "/projects/"+projectID+"/tasks", body).want(t, http.StatusCreated)
	return decode[taskJSON](t, r)
}

// newMilestone создаёт веху проекта; fields дополняют и переопределяют тело запроса.
func newMilestone(t *testing.T, token, projectID string, fields map[string]any) milestoneJSON {
	t.Helper()
	body := map[string]any{"name": fmt.Sprintf("Веха %d", uniq()), "plannedDate": day(5)}
	for k, v := range fields {
		body[k] = v
	}
	r := post(t, token, "/projects/"+projectID+"/milestones", body).want(t, http.StatusCreated)
	return decode[milestoneJSON](t, r)
}

// day возвращает дату «сегодня + offset дней» в формате API (YYYY-MM-DD).
func day(offset int) string {
	return models.Today().AddDays(offset).String()
}

// nextMonday — ближайший понедельник не раньше чем через неделю: от дня
// недели зависит число рабочих дней, и тесту нужна предсказуемая точка отсчёта.
func nextMonday() models.Date {
	d := models.Today().AddDays(7)
	for d.Weekday() != time.Monday {
		d = d.AddDays(1)
	}
	return d
}

// historyActions возвращает коды действий из журнала задачи.
func historyActions(t *testing.T, token, taskID string) []string {
	t.Helper()
	entries := decode[[]historyJSON](t, get(t, token, "/tasks/"+taskID+"/history").want(t, http.StatusOK))
	actions := make([]string, 0, len(entries))
	for _, e := range entries {
		actions = append(actions, e.Action)
	}
	return actions
}

// intOrNull печатает nullable-число так, как оно выглядит в JSON.
func intOrNull(p *int) string {
	if p == nil {
		return "null"
	}
	return fmt.Sprint(*p)
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// mustNotContainPassword проверяет, что в ответ не утекло ничего похожего на пароль.
func mustNotContainPassword(t *testing.T, r response) {
	t.Helper()
	if bytes.Contains(bytes.ToLower(r.Body), []byte("password")) {
		t.Fatalf("в ответе есть поле с паролем: %s", r.Body)
	}
}

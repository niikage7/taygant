package server_test

// Тесты «обвязки» приложения: health-check, формат ошибок, CORS и лимит
// запросов. PostgreSQL не нужен — подключение к БД ленивое и указывает на
// заведомо закрытый порт, а всё проверяемое здесь до базы не доходит.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"taygant_backend/internal/config"
	"taygant_backend/internal/server"
)

const frontendOrigin = "http://localhost:3000"

func newApp(t *testing.T, tweak func(*config.Config)) *fiber.App {
	t.Helper()
	cfg := config.Config{
		AllowedOrigins:  []string{frontendOrigin},
		RateLimitMax:    100,
		RateLimitWindow: time.Minute,
		JWTSecret:       "test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: time.Hour,
	}
	if tweak != nil {
		tweak(&cfg)
	}

	db, err := gorm.Open(postgres.Open("postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=1"),
		&gorm.Config{DisableAutomaticPing: true, Logger: logger.Discard})
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	// Логгер запросов Fiber запоминает os.Stdout в момент сборки приложения.
	if devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0); err == nil {
		original := os.Stdout
		os.Stdout = devNull
		defer func() { os.Stdout = original }()
	}
	return server.New(cfg, db)
}

type errorBody struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

func do(t *testing.T, app *fiber.App, req *http.Request) (*http.Response, errorBody) {
	t.Helper()
	resp, err := app.Test(req, 5_000)
	if err != nil {
		t.Fatalf("%s %s: %v", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var body errorBody
	_ = json.Unmarshal(raw, &body)
	return resp, body
}

func request(method, path string, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

// docker compose считает бэкенд живым по /ping — без базы он живым не считается.
func TestPingReportsUnavailableDatabase(t *testing.T) {
	resp, body := do(t, newApp(t, nil), request(http.MethodGet, "/ping", nil))
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("/ping без БД: код %d, ожидался 503", resp.StatusCode)
	}
	if body.Message == "" || body.Path != "/ping" {
		t.Fatalf("тело ошибки = %+v", body)
	}
}

// Фронт разбирает любые ошибки одинаково: JSON с error, message и path.
func TestErrorsUseUnifiedJSONFormat(t *testing.T) {
	app := newApp(t, nil)
	for _, tc := range []struct {
		path string
		want int
	}{
		{"/no-such-route", http.StatusNotFound},
		{"/api/v1/projects", http.StatusUnauthorized},
		{"/api/v1/no-such-route", http.StatusUnauthorized}, // закрыто по умолчанию, даже чего нет
	} {
		resp, body := do(t, app, request(http.MethodGet, tc.path, nil))
		if resp.StatusCode != tc.want {
			t.Errorf("%s: код %d, ожидался %d", tc.path, resp.StatusCode, tc.want)
		}
		if body.Error == "" || body.Error != body.Message || body.Path != tc.path {
			t.Errorf("%s: тело ошибки = %+v", tc.path, body)
		}
		if ct := resp.Header.Get(fiber.HeaderContentType); !strings.HasPrefix(ct, fiber.MIMEApplicationJSON) {
			t.Errorf("%s: Content-Type = %q", tc.path, ct)
		}
	}
}

func TestCORSAllowsOnlyConfiguredOrigins(t *testing.T) {
	app := newApp(t, nil)
	preflight := func(origin string) *http.Response {
		resp, _ := do(t, app, request(http.MethodOptions, "/api/v1/projects", map[string]string{
			"Origin":                         origin,
			"Access-Control-Request-Method":  http.MethodPatch,
			"Access-Control-Request-Headers": "authorization,content-type",
		}))
		return resp
	}

	allowed := preflight(frontendOrigin)
	if got := allowed.Header.Get("Access-Control-Allow-Origin"); got != frontendOrigin {
		t.Fatalf("для фронта Access-Control-Allow-Origin = %q", got)
	}
	if !strings.Contains(allowed.Header.Get("Access-Control-Allow-Methods"), http.MethodPatch) {
		t.Fatalf("PATCH не разрешён: %q", allowed.Header.Get("Access-Control-Allow-Methods"))
	}

	evil := preflight("https://evil.example")
	if got := evil.Header.Get("Access-Control-Allow-Origin"); got == "https://evil.example" || got == "*" {
		t.Fatalf("чужому сайту разрешён доступ: Access-Control-Allow-Origin = %q", got)
	}
}

func TestRateLimit(t *testing.T) {
	app := newApp(t, func(cfg *config.Config) { cfg.RateLimitMax = 2 })
	status := func(req *http.Request) int {
		resp, _ := do(t, app, req)
		return resp.StatusCode
	}

	// Preflight браузера и health-check квоту не тратят.
	for i := 0; i < 5; i++ {
		preflight := request(http.MethodOptions, "/api/v1/projects", map[string]string{
			"Origin": frontendOrigin, "Access-Control-Request-Method": http.MethodGet,
		})
		if code := status(preflight); code == http.StatusTooManyRequests {
			t.Fatal("preflight-запрос съел квоту")
		}
		if code := status(request(http.MethodGet, "/ping", nil)); code == http.StatusTooManyRequests {
			t.Fatal("health-check попал под лимит")
		}
	}

	for i := 1; i <= 2; i++ {
		if code := status(request(http.MethodGet, "/api/v1/projects", nil)); code != http.StatusUnauthorized {
			t.Fatalf("запрос %d в пределах лимита: код %d, ожидался 401", i, code)
		}
	}

	// Подделанный X-Forwarded-For от недоверенного адреса новую квоту не даёт.
	resp, body := do(t, app, request(http.MethodGet, "/api/v1/projects", map[string]string{"X-Forwarded-For": "203.0.113.7"}))
	if resp.StatusCode != http.StatusTooManyRequests || body.Message == "" {
		t.Fatalf("сверх лимита: код %d, тело %+v, ожидался 429 с сообщением", resp.StatusCode, body)
	}
}

// /mcp смонтирован и закрыт личным ключом ещё до обращения к базе: без ключа —
// 401, а не 500 из-за недоступного PostgreSQL. Лимит запросов у него свой.
func TestMCPEndpointRequiresPersonalToken(t *testing.T) {
	app := newApp(t, func(cfg *config.Config) { cfg.RateLimitMax = 1 })
	mcpRequest := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		return req
	}

	if resp, _ := do(t, app, mcpRequest()); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("POST /mcp без ключа: код %d, ожидался 401", resp.StatusCode)
	}
	// Квоту /api/v1 запросы ассистента не тратят.
	if resp, _ := do(t, app, request(http.MethodGet, "/api/v1/projects", nil)); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("после запроса к /mcp /api/v1 ответил %d, ожидался 401", resp.StatusCode)
	}
	if resp, _ := do(t, app, mcpRequest()); resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("второй запрос к /mcp при лимите 1: код %d, ожидался 429", resp.StatusCode)
	}
}

// За доверенным прокси (Next.js в compose) квота считается по реальному IP клиента.
func TestRateLimitUsesClientIPBehindTrustedProxy(t *testing.T) {
	app := newApp(t, func(cfg *config.Config) {
		cfg.RateLimitMax = 1
		cfg.TrustedProxies = []string{"0.0.0.0"} // адрес, с которого «приходит» app.Test
	})
	status := func(clientIP string) int {
		resp, _ := do(t, app, request(http.MethodGet, "/api/v1/projects", map[string]string{"X-Forwarded-For": clientIP}))
		return resp.StatusCode
	}

	if status("203.0.113.1") != http.StatusUnauthorized || status("203.0.113.2") != http.StatusUnauthorized {
		t.Fatal("разные клиенты за прокси делят одну квоту")
	}
	if code := status("203.0.113.1"); code != http.StatusTooManyRequests {
		t.Fatalf("повторный запрос того же клиента: код %d, ожидался 429", code)
	}
}

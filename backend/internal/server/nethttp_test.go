package server

// Тесты моста между Fiber и net/http, через который работает /mcp.
// PostgreSQL не нужен: мост ничего не знает о базе.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func bridgeApp(t *testing.T, h http.Handler, timeout time.Duration) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.All("/bridge", netHTTPHandler(h, timeout))
	return app
}

func doBridge(t *testing.T, app *fiber.App, req *http.Request) (*http.Response, string) {
	t.Helper()
	resp, err := app.Test(req, 5_000)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, string(body)
}

func TestNetHTTPHandlerPassesRequestAndResponseThrough(t *testing.T) {
	app := bridgeApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		switch {
		case r.Method != http.MethodPost, r.URL.Path != "/bridge", r.URL.Query().Get("q") != "1":
			http.Error(w, "метод или адрес: "+r.Method+" "+r.URL.String(), http.StatusTeapot)
		case r.Header.Get("Authorization") != "Bearer tgn_X", r.Host != "api.test":
			http.Error(w, "заголовки не дошли", http.StatusTeapot)
		case string(body) != `{"привет":"мир"}` || r.ContentLength != int64(len(body)):
			http.Error(w, "тело не дошло: "+string(body), http.StatusTeapot)
		case r.RemoteAddr == "":
			http.Error(w, "нет адреса клиента", http.StatusTeapot)
		default:
			if _, ok := r.Context().Deadline(); !ok {
				http.Error(w, "у контекста нет таймаута", http.StatusTeapot)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Add("X-Multi", "a")
			w.Header().Add("X-Multi", "b")
			// Неверная длина от обработчика не должна попасть в ответ.
			w.Header().Set("Content-Length", "999")
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusInternalServerError) // второй вызов, как в net/http, игнорируется
			_, _ = w.Write([]byte(`{"ok":true}`))
		}
	}), time.Second)

	req := httptest.NewRequest(http.MethodPost, "http://api.test/bridge?q=1", strings.NewReader(`{"привет":"мир"}`))
	req.Header.Set("Authorization", "Bearer tgn_X")
	resp, body := doBridge(t, app, req)

	if resp.StatusCode != http.StatusCreated || body != `{"ok":true}` {
		t.Fatalf("ответ: %d %s", resp.StatusCode, body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if got := resp.Header.Values("X-Multi"); strings.Join(got, ",") != "a,b" {
		t.Fatalf("многозначный заголовок = %v", got)
	}
	if resp.ContentLength != int64(len(body)) {
		t.Fatalf("Content-Length = %d при теле из %d байт", resp.ContentLength, len(body))
	}
}

func TestNetHTTPHandlerDefaultsAndFlusher(t *testing.T) {
	app := bridgeApp(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// SDK требует http.Flusher для потоковых ответов — мост его предоставляет.
		if _, ok := w.(http.Flusher); !ok {
			http.Error(w, "нет Flusher", http.StatusTeapot)
			return
		}
		_, _ = w.Write([]byte("без WriteHeader"))
	}), time.Second)

	resp, body := doBridge(t, app, httptest.NewRequest(http.MethodGet, "/bridge", nil))
	if resp.StatusCode != http.StatusOK || body != "без WriteHeader" {
		t.Fatalf("ответ: %d %s", resp.StatusCode, body)
	}

	empty := bridgeApp(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), time.Second)
	if resp, body := doBridge(t, empty, httptest.NewRequest(http.MethodPost, "/bridge", nil)); resp.StatusCode != http.StatusAccepted || body != "" {
		t.Fatalf("пустой ответ: %d %q", resp.StatusCode, body)
	}
}

// Обработчик, забывший ответить, не держит запрос дольше таймаута.
func TestNetHTTPHandlerCancelsContextOnTimeout(t *testing.T) {
	app := bridgeApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		http.Error(w, "время вышло", http.StatusServiceUnavailable)
	}), 50*time.Millisecond)

	start := time.Now()
	resp, _ := doBridge(t, app, httptest.NewRequest(http.MethodPost, "/bridge", nil))
	if resp.StatusCode != http.StatusServiceUnavailable || time.Since(start) > 2*time.Second {
		t.Fatalf("код %d за %v", resp.StatusCode, time.Since(start))
	}
}

package server

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// netHTTPHandler встраивает обработчик стандартного net/http в Fiber.
//
// MCP SDK написан под net/http, а приложение — на Fiber поверх fasthttp.
// Готовый адаптер Fiber (middleware/adaptor) не подходит: он отдаёт
// обработчику в роли context.Context сам fasthttp.RequestCtx, а fasthttp
// переиспользует этот объект для следующих запросов. Любая горутина SDK,
// пережившая ответ, читала бы уже чужой запрос. Здесь запрос копируется
// целиком, контекст — собственный, с таймаутом, а ответ копится в буфере и
// передаётся Fiber, когда обработчик уже завершился.
//
// Буферизация означает, что потоковые ответы (SSE) уходят одним куском;
// MCP-сервер настроен отвечать обычным JSON, поэтому для него это не важно.
func netHTTPHandler(h http.Handler, timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		body := bytes.Clone(c.Body())
		req, err := http.NewRequestWithContext(ctx, c.Method(), c.OriginalURL(), bytes.NewReader(body))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный запрос")
		}
		c.Request().Header.VisitAll(func(key, value []byte) {
			req.Header.Add(string(key), string(value))
		})
		req.Host = string(c.Request().Host())
		req.RemoteAddr = c.Context().RemoteAddr().String()
		req.RequestURI = c.OriginalURL()
		req.ContentLength = int64(len(body))

		w := &bufferedResponseWriter{header: make(http.Header)}
		h.ServeHTTP(w, req)

		for key, values := range w.header {
			// Длину тела посчитает fasthttp: заголовок от обработчика мог
			// разойтись с тем, что реально лежит в буфере.
			if strings.EqualFold(key, fiber.HeaderContentLength) {
				continue
			}
			for _, value := range values {
				c.Response().Header.Add(key, value)
			}
		}
		c.Status(w.statusCode())
		return c.Send(w.body.Bytes())
	}
}

// bufferedResponseWriter — http.ResponseWriter, который копит ответ в памяти.
type bufferedResponseWriter struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (w *bufferedResponseWriter) Header() http.Header { return w.header }

// WriteHeader, как и в net/http, учитывает только первый вызов.
func (w *bufferedResponseWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *bufferedResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}

// Flush ничего не делает: ответ всё равно уходит целиком после обработчика.
// Метод нужен, чтобы обработчики, которые требуют http.Flusher, не падали.
func (w *bufferedResponseWriter) Flush() {}

func (w *bufferedResponseWriter) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

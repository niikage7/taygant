// Package mcpserver — MCP-сервер taygant: подключение приложения к нейронке,
// которой человек уже пользуется (Claude Code, Cursor…). Нейронка отвечает на
// «что у меня сегодня?», «как дела у проекта?», «что горит?» и меняет статус
// задачи, когда человек это подтвердил. Зачем это нужно — docs/mcp.md.
//
// Протокол реализует официальный Go SDK (github.com/modelcontextprotocol/go-sdk):
// версии MCP выходят несколько раз в год, и согласование версий, формат
// сообщений и схемы инструментов надёжнее доверить ему. Здесь — только наша
// часть: вход по личным ключам и инструменты поверх сервисного слоя. Сервисы те
// же, что у REST API, поэтому права, эффективные статусы и расчёт сроков
// совпадают с тем, что человек видит в приложении.
package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
	"taygant_backend/internal/service"
)

// serverVersion — версия набора инструментов. Клиенты показывают её рядом с
// именем сервера; менять при несовместимых изменениях инструментов.
const serverVersion = "1.0.0"

// Server — инструменты MCP и сервисы, на которые они опираются.
type Server struct {
	db           *gorm.DB
	tokens       *service.MCPTokens
	tasks        *service.Tasks
	dependencies *service.Dependencies
	checklist    *service.Checklist
	comments     *service.Comments
	history      *service.History
	dashboard    *service.Dashboard
}

func newServer(db *gorm.DB) *Server {
	tasks := service.NewTasks(db)
	milestones := service.NewMilestones(db)
	return &Server{
		db:           db,
		tokens:       service.NewMCPTokens(db),
		tasks:        tasks,
		dependencies: service.NewDependencies(db),
		checklist:    service.NewChecklist(db),
		comments:     service.NewComments(db),
		history:      service.NewHistory(db),
		dashboard:    service.NewDashboard(db, tasks, milestones, service.NewWorkload(db)),
	}
}

// NewHandler собирает http.Handler, обслуживающий MCP по Streamable HTTP.
//
// Порядок обёрток: сначала защита от запросов со сторонних сайтов, потом
// проверка личного ключа, и только потом сам протокол — до инструментов не
// доходит ни один запрос без действующего ключа.
func NewHandler(db *gorm.DB) http.Handler {
	s := newServer(db)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "taygant",
		Title:   "taygant — проекты и диаграмма Ганта",
		Version: serverVersion,
	}, &mcp.ServerOptions{
		Instructions: instructions,
		// Набор инструментов не меняется на лету, поэтому listChanged не
		// объявляем. Logging тоже: ответы уходят одним JSON, и серверным
		// логам по ходу вызова деваться некуда.
		Capabilities: &mcp.ServerCapabilities{Tools: &mcp.ToolCapabilities{}},
	})
	s.registerTools(server)

	streamable := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{
		// Stateless — каждый запрос самодостаточен, сервер не хранит сессий:
		// так работает любая версия протокола (с 2026-07-28 — только так), а
		// перезапуск бэкенда не рвёт подключения ассистентов.
		Stateless: true,
		// JSONResponse — ответ одним JSON вместо потока SSE. Инструменты
		// отвечают за миллисекунды, стримить нечего, а обычный JSON без
		// сюрпризов проходит через Fiber и прокси Next.js (/backend/mcp).
		JSONResponse: true,
		// Оборвался HTTP-запрос — ответ уже некому отдать, работу можно бросить.
		PropagateRequestCancellation: true,
	})

	withAuth := auth.RequireBearerToken(s.verifyToken, &auth.RequireBearerTokenOptions{
		// У личного ключа нет срока действия: его отзывают в приложении.
		AllowMissingExpiration: true,
	})(streamable)

	// Запросы со сторонних сайтов из браузера отклоняются (защита от DNS
	// rebinding, которую требует спецификация MCP). Настоящие MCP-клиенты —
	// не браузеры, заголовков Origin/Sec-Fetch-Site не шлют и проходят.
	return http.NewCrossOriginProtection().Handler(withAuth)
}

// callerKey — ключ в TokenInfo.Extra, под которым verifyToken оставляет
// инструментам владельца ключа.
const callerKey = "taygant/caller"

// caller — от чьего имени работает ассистент и каким ключом он подключён.
type caller struct {
	User models.User
	// TokenName — подпись ключа («Claude Code, ноутбук»); попадает в журнал
	// задачи рядом с пометкой «через ассистента».
	TokenName string
}

// verifyToken проверяет личный ключ из заголовка Authorization.
//
// Пользователь загружается здесь, один раз на запрос, и передаётся
// инструментам через TokenInfo — так же, как requireAuth в REST API кладёт
// его в контекст Fiber.
func (s *Server) verifyToken(ctx context.Context, secret string, _ *http.Request) (*auth.TokenInfo, error) {
	token, err := s.tokens.Authenticate(ctx, secret)
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		return nil, fmt.Errorf("%w: ключ не найден или отозван — создайте новый в приложении, раздел «Ассистент»", auth.ErrInvalidToken)
	case err != nil:
		// Текст ошибки ушёл бы клиенту в тело ответа, а в нём может оказаться
		// фрагмент SQL — наружу только общая фраза.
		log.Printf("mcp: проверка ключа: %v", err)
		return nil, errors.New("не удалось проверить ключ, попробуйте позже")
	}
	return &auth.TokenInfo{
		UserID: token.UserID.String(),
		Extra:  map[string]any{callerKey: caller{User: *token.User, TokenName: token.Name}},
	}, nil
}

// callerFrom достаёт владельца ключа из запроса к инструменту. Без него
// запрос не прошёл бы verifyToken, так что ошибка здесь — ошибка сборки
// обработчика, а не клиента.
func callerFrom(req *mcp.CallToolRequest) (caller, error) {
	if req.Extra != nil && req.Extra.TokenInfo != nil {
		if c, ok := req.Extra.TokenInfo.Extra[callerKey].(caller); ok {
			return c, nil
		}
	}
	return caller{}, errors.New("mcp: запрос к инструменту без проверенного ключа")
}

// instructions получает нейронка при подключении. Главное здесь — правило про
// подтверждение: docs/mcp.md обещает, что без согласия человека ассистент
// ничего не меняет.
const instructions = `taygant — приложение для планирования проектов с диаграммой Ганта. Вы работаете от имени владельца ключа и видите только проекты, в команде которых он состоит.

Что спрашивают и чем отвечать:
- «Что у меня сегодня?», «над чем я работаю?» — my_tasks.
- «Как дела у проекта?» — project_status. «Что горит?» — project_risks.
- Подробности задачи: описание, критерии готовности, связи, история — get_task.
- Список проектов и права в них — list_my_projects.
- Смена статуса — set_task_status.

Главное правило: ничего не меняйте без явного согласия человека. Вызывайте set_task_status, только если человек в этом разговоре сам попросил сменить статус («беру TASK-005 в работу») или ответил «да» на ваше предложение. Если по работе в коде видно, что задача, похоже, готова, — не закрывайте её сами, а спросите: «Похоже, TASK-005 готова. Закрыть?» — и дождитесь ответа. Перед закрытием полезно свериться с критериями готовности из get_task.

Статусы, которые можно поставить: planned (запланирована), in_progress (в работе), done (готово). «Просрочена» (overdue) и «заблокирована» (blocked) приложение вычисляет само по датам и связям — выставить их нельзя.

Называйте задачи кодом и названием, например: TASK-005 «Настроить CI». Коды уникальны внутри проекта; если код встречается в нескольких проектах, инструмент попросит уточнить проект.`

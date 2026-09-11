package mcpserver

import (
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// registerTools объявляет инструменты. Описания адресованы нейронке: по ним
// она решает, какой инструмент позвать на вопрос человека, поэтому в них
// сказано, на какие вопросы инструмент отвечает.
//
// Аннотации (readOnlyHint и др.) — подсказки клиенту: инструменты для чтения
// клиент может разрешить без вопросов, а перед set_task_status спросит
// человека. Схемы входа и выхода SDK выводит из Go-структур.
func (s *Server) registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "my_tasks",
		Title:       "Мои задачи",
		Description: "Задачи, где владелец ключа — исполнитель: отвечает на «что у меня сегодня?», «над чем я работаю?», «что у меня горит?». Сначала просроченные, потом начатые, ждущие других и будущие; внутри группы — по сроку. По умолчанию из всех проектов и без завершённых.",
		Annotations: readOnly("Мои задачи"),
	}, s.myTasks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Title:       "Карточка задачи",
		Description: "Всё о задаче: описание, критерии готовности (чек-лист), связи с другими задачами, последние изменения и комментарии, можно ли владельцу ключа менять её статус. Полезно, чтобы проверить, готова ли задача, прежде чем предлагать её закрыть.",
		Annotations: readOnly("Карточка задачи"),
	}, s.getTask)

	mcp.AddTool(server, &mcp.Tool{
		Name:  "set_task_status",
		Title: "Сменить статус задачи",
		Description: "Меняет статус задачи от имени владельца ключа: planned, in_progress или done. " +
			"Вызывайте только после явного согласия человека в этом разговоре: он сам попросил («беру TASK-005 в работу») или ответил «да» на ваш вопрос «Похоже, TASK-005 готова. Закрыть?». " +
			"Сами, без спроса, статусы не меняйте. В истории задачи смена будет отмечена как сделанная через ассистента. " +
			"Если клиент это умеет, сервер дополнительно покажет человеку окно подтверждения.",
		InputSchema: inputSchemaWithEnum[setStatusInput]("status", userSettableStatuses),
		Annotations: &mcp.ToolAnnotations{
			Title: "Сменить статус задачи",
			// Статус можно вернуть обратно — это не удаление данных.
			DestructiveHint: ptr(false),
			IdempotentHint:  true,
			OpenWorldHint:   ptr(false),
		},
	}, s.setTaskStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_status",
		Title:       "Как дела у проекта",
		Description: "Сводка по проекту: успевает ли к дедлайну (прогноз по критическому пути), прогресс, сколько задач готово, в работе, просрочено и ждёт других, ближайшие контрольные точки, что сейчас в работе. Отвечает на «как дела у проекта?».",
		Annotations: readOnly("Как дела у проекта"),
	}, s.projectStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "project_risks",
		Title:       "Что горит",
		Description: "Проблемы проекта, самые срочные первыми: прогноз срыва дедлайна, просроченные задачи, задачи, которые уже должны идти, но ждут других или не начаты, просроченные контрольные точки. Отвечает на «что горит?», «где риски?».",
		Annotations: readOnly("Что горит"),
	}, s.projectRisks)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_my_projects",
		Title:       "Мои проекты",
		Description: "Проекты, в команде которых состоит владелец ключа: код, сроки, прогресс, число просрочек, его уровень доступа и сколько на нём открытых задач. Нужен, чтобы выбрать проект, если их несколько.",
		Annotations: readOnly("Мои проекты"),
	}, s.listMyProjects)
}

// readOnly — аннотация инструмента, который только читает данные приложения.
func readOnly(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true, OpenWorldHint: ptr(false)}
}

// inputSchemaWithEnum выводит схему входа из структуры и ограничивает одно
// поле списком значений. Enum виден нейронке прямо в схеме, а SDK отклоняет
// другие значения ещё до вызова обработчика.
func inputSchemaWithEnum[T any](field string, values []any) *jsonschema.Schema {
	schema, err := jsonschema.For[T](nil)
	if err != nil {
		panic(fmt.Sprintf("mcp: схема входа: %v", err))
	}
	prop, ok := schema.Properties[field]
	if !ok {
		panic(fmt.Sprintf("mcp: в схеме нет поля %q", field))
	}
	prop.Enum = values
	return schema
}

func ptr[T any](v T) *T { return &v }

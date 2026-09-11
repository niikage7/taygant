package service

import (
	"context"

	"taygant_backend/internal/models"
)

// ChangeSource — через что пользователь внёс изменение: веб-приложение или
// нейронку, подключённую по MCP. Попадает в каждую запись журнала задачи.
type ChangeSource struct {
	// Channel — models.HistorySourceApp или models.HistorySourceAssistant.
	Channel string
	// Via — уточнение канала: для ассистента это имя личного ключа.
	Via string
}

type changeSourceKey struct{}

// WithChangeSource помечает все изменения, сделанные в рамках ctx, указанным
// каналом.
//
// Канал передаётся контекстом, а не параметром: журнал пишут все сервисы,
// меняющие задачи (задачи, связи, чек-лист, комментарии), и протаскивать
// «откуда пришла правка» через каждую их сигнатуру ради одного MCP-слоя
// значило бы менять весь сервисный API. Контекст и так несёт сведения
// о запросе, а recordHistory — единственное место, где канал читается.
func WithChangeSource(ctx context.Context, src ChangeSource) context.Context {
	return context.WithValue(ctx, changeSourceKey{}, src)
}

// changeSourceFrom возвращает канал из контекста. Без пометки изменение
// считается сделанным в приложении — так работают все REST-хендлеры.
func changeSourceFrom(ctx context.Context) ChangeSource {
	if ctx != nil {
		if src, ok := ctx.Value(changeSourceKey{}).(ChangeSource); ok && src.Channel != "" {
			return src
		}
	}
	return ChangeSource{Channel: models.HistorySourceApp}
}

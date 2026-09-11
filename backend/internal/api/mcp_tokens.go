package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"taygant_backend/internal/models"
)

// registerMCPTokensRoutes — личные ключи, которыми нейронка пользователя
// подключается к MCP-серверу (docs/mcp.md). Ключи всегда свои: чужие не
// видны и не отзываются, поэтому проверка прав — это сам requireAuth.
func (a *API) registerMCPTokensRoutes(r fiber.Router) {
	tokens := r.Group("/users/me/mcp-tokens")
	tokens.Get("/", a.mcpTokensList)
	tokens.Post("/", a.mcpTokensCreate)
	tokens.Delete("/:tokenId", a.mcpTokensRevoke)
}

// mcpTokenDTO — схема McpToken. Секрета здесь нет и быть не может: в БД
// хранится только его хеш.
type mcpTokenDTO struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Hint       string    `json:"hint"`
	CreatedAt  string    `json:"createdAt"`
	LastUsedAt *string   `json:"lastUsedAt"`
}

// newMCPTokenDTO собирает ответ. Время — в UTC: только что созданный ключ и
// тот же ключ, прочитанный из БД, иначе приходили бы в разных часовых поясах.
func newMCPTokenDTO(t models.MCPToken) mcpTokenDTO {
	dto := mcpTokenDTO{
		ID:        t.ID,
		Name:      t.Name,
		Hint:      t.Hint,
		CreatedAt: t.CreatedAt.UTC().Format(rfc3339),
	}
	if t.LastUsedAt != nil {
		used := t.LastUsedAt.UTC().Format(rfc3339)
		dto.LastUsedAt = &used
	}
	return dto
}

// mcpTokenCreatedDTO — схема McpTokenCreated: единственный ответ, в котором
// есть сам ключ.
type mcpTokenCreatedDTO struct {
	Token  mcpTokenDTO `json:"token"`
	Secret string      `json:"secret"`
}

// GET /users/me/mcp-tokens — ключи текущего пользователя, новые сверху.
// 200 -> []McpToken.
func (a *API) mcpTokensList(c *fiber.Ctx) error {
	tokens, err := a.mcpTokens.List(c.Context(), currentUserID(c))
	if err != nil {
		return fail(err)
	}
	out := make([]mcpTokenDTO, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, newMCPTokenDTO(t))
	}
	return c.JSON(out)
}

// mcpTokenCreateRequest — тело POST /users/me/mcp-tokens.
type mcpTokenCreateRequest struct {
	Name string `json:"name"`
}

// POST /users/me/mcp-tokens — выпустить ключ для подключения ассистента.
// Body: { name }. 201 -> McpTokenCreated, 400 -> пустое название или лимит ключей.
func (a *API) mcpTokensCreate(c *fiber.Ctx) error {
	body, err := parseBody[mcpTokenCreateRequest](c)
	if err != nil {
		return err
	}
	token, secret, err := a.mcpTokens.Issue(c.Context(), currentUserID(c), body.Name)
	if err != nil {
		return fail(err)
	}
	return c.Status(fiber.StatusCreated).JSON(mcpTokenCreatedDTO{Token: newMCPTokenDTO(token), Secret: secret})
}

// DELETE /users/me/mcp-tokens/{tokenId} — отозвать ключ: подключённый им
// ассистент теряет доступ со следующего запроса. 204, 404 -> ключа нет среди своих.
func (a *API) mcpTokensRevoke(c *fiber.Ctx) error {
	tokenID, err := pathUUID(c, "tokenId")
	if err != nil {
		return err
	}
	if err := a.mcpTokens.Revoke(c.Context(), currentUserID(c), tokenID); err != nil {
		return fail(err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

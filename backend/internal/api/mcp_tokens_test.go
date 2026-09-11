package api_test

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
)

// ---------- Личные ключи для подключения ассистента (MCP) ----------

type mcpTokenJSON struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Hint       string  `json:"hint"`
	CreatedAt  string  `json:"createdAt"`
	LastUsedAt *string `json:"lastUsedAt"`
}

type mcpTokenCreatedJSON struct {
	Token  mcpTokenJSON `json:"token"`
	Secret string       `json:"secret"`
}

const mcpTokensPath = "/users/me/mcp-tokens"

// newMCPToken выпускает личный ключ пользователю через API.
func newMCPToken(t *testing.T, user testUser, name string) mcpTokenCreatedJSON {
	t.Helper()
	r := post(t, user.Token, mcpTokensPath, map[string]any{"name": name}).want(t, http.StatusCreated)
	return decode[mcpTokenCreatedJSON](t, r)
}

func TestMCPTokenLifecycle(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Разработчик с ассистентом")

	created := newMCPToken(t, user, "  Claude Code, ноутбук ")
	if !strings.HasPrefix(created.Secret, "tgn_") || len(created.Secret) < 30 {
		t.Fatalf("секрет %q: ожидался префикс tgn_ и не меньше 30 символов", created.Secret)
	}
	if created.Token.Name != "Claude Code, ноутбук" {
		t.Fatalf("name = %q, ожидались обрезанные пробелы", created.Token.Name)
	}
	if created.Token.Hint != created.Secret[len(created.Secret)-4:] {
		t.Fatalf("hint = %q, ожидались последние 4 символа ключа", created.Token.Hint)
	}
	if created.Token.LastUsedAt != nil {
		t.Fatalf("новым ключом ещё не пользовались, а lastUsedAt = %v", *created.Token.LastUsedAt)
	}

	// Показать ключ повторно нельзя: в списке его нет — ни целиком, ни хешем.
	r := get(t, user.Token, mcpTokensPath).want(t, http.StatusOK)
	if bytes.Contains(r.Body, []byte(created.Secret)) || bytes.Contains(bytes.ToLower(r.Body), []byte("hash")) {
		t.Fatalf("в списке ключей утёк секрет или его хеш: %s", r.Body)
	}
	list := decode[[]mcpTokenJSON](t, r)
	if len(list) != 1 || list[0].ID != created.Token.ID || list[0].Hint != created.Token.Hint {
		t.Fatalf("список ключей = %+v, ожидался один выпущенный ключ", list)
	}

	del(t, user.Token, mcpTokensPath+"/"+created.Token.ID).want(t, http.StatusNoContent)
	if list := decode[[]mcpTokenJSON](t, get(t, user.Token, mcpTokensPath).want(t, http.StatusOK)); len(list) != 0 {
		t.Fatalf("после отзыва в списке остались ключи: %+v", list)
	}
	del(t, user.Token, mcpTokensPath+"/"+created.Token.ID).want(t, http.StatusNotFound)
}

func TestMCPTokensArePrivate(t *testing.T) {
	requireDB(t)
	owner := newUser(t, "Владелец ключа")
	other := newUser(t, "Посторонний")
	created := newMCPToken(t, owner, "Cursor")

	if list := decode[[]mcpTokenJSON](t, get(t, other.Token, mcpTokensPath).want(t, http.StatusOK)); len(list) != 0 {
		t.Fatalf("чужие ключи видны в списке: %+v", list)
	}
	// Чужой ключ неотличим от несуществующего.
	del(t, other.Token, mcpTokensPath+"/"+created.Token.ID).want(t, http.StatusNotFound)
	if list := decode[[]mcpTokenJSON](t, get(t, owner.Token, mcpTokensPath).want(t, http.StatusOK)); len(list) != 1 {
		t.Fatalf("ключ пропал после попытки чужого отзыва: %+v", list)
	}
}

func TestMCPTokenValidation(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Проверяющий ключи")

	post(t, "", mcpTokensPath, map[string]any{"name": "Без входа"}).want(t, http.StatusUnauthorized)
	get(t, "", mcpTokensPath).want(t, http.StatusUnauthorized)
	post(t, user.Token, mcpTokensPath, map[string]any{"name": "   "}).want(t, http.StatusBadRequest)
	post(t, user.Token, mcpTokensPath, map[string]any{"name": strings.Repeat("я", 121)}).want(t, http.StatusBadRequest)
	sendRaw(t, http.MethodPost, mcpTokensPath, user.Token, "{битый json").want(t, http.StatusBadRequest)
	del(t, user.Token, mcpTokensPath+"/не-uuid").want(t, http.StatusBadRequest)
}

func TestMCPTokenLimitPerUser(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Коллекционер ключей")
	for i := 0; i < 10; i++ {
		newMCPToken(t, user, "Ключ")
	}
	r := post(t, user.Token, mcpTokensPath, map[string]any{"name": "Одиннадцатый"}).want(t, http.StatusBadRequest)
	if msg := decode[errorJSON](t, r).Message; !strings.Contains(msg, "отзовите") {
		t.Fatalf("сообщение о лимите не подсказывает, что делать: %q", msg)
	}
}

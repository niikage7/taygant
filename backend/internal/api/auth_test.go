package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRegisterReturnsTokensAndHidesPassword(t *testing.T) {
	requireDB(t)
	n := uniq()

	r := post(t, "", "/auth/register", map[string]any{
		"email":      fmt.Sprintf("  New.User%d@Test.DEV ", n),
		"password":   "secret123",
		"fullName":   "  Новый Пользователь ",
		"department": "Отдел разработки",
		"position":   "Инженер",
	}).want(t, http.StatusCreated)
	mustNotContainPassword(t, r)

	auth := decode[authJSON](t, r)
	if auth.AccessToken == "" || auth.RefreshToken == "" {
		t.Fatalf("регистрация не выдала токены: %s", r.Body)
	}
	if want := fmt.Sprintf("new.user%d@test.dev", n); auth.User.Email != want {
		t.Fatalf("email = %q, ожидался нормализованный %q", auth.User.Email, want)
	}
	if auth.User.FullName != "Новый Пользователь" {
		t.Fatalf("fullName = %q, ожидались обрезанные пробелы", auth.User.FullName)
	}

	// После регистрации человек сразу в приложении — токен рабочий.
	me := decode[userJSON](t, get(t, auth.AccessToken, "/users/me").want(t, http.StatusOK))
	if me.ID != auth.User.ID || me.Department != "Отдел разработки" {
		t.Fatalf("/users/me вернул %+v, ожидался только что созданный пользователь", me)
	}
}

func TestRegisterRejectsDuplicateEmailIgnoringCase(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Первый")

	r := post(t, "", "/auth/register", map[string]any{
		"email":    strings.ToUpper(user.Email),
		"password": testPassword,
		"fullName": "Второй",
	})
	r.want(t, http.StatusConflict)
}

func TestRegisterValidatesInput(t *testing.T) {
	requireDB(t)

	cases := []struct {
		name string
		body map[string]any
	}{
		{"email без @", map[string]any{"email": "nobody", "password": testPassword, "fullName": "Имя"}},
		{"короткий пароль", map[string]any{"email": fmt.Sprintf("v%d@test.dev", uniq()), "password": "abc1", "fullName": "Имя"}},
		{"пароль без цифр", map[string]any{"email": fmt.Sprintf("v%d@test.dev", uniq()), "password": "onlyletters", "fullName": "Имя"}},
		{"пустое имя", map[string]any{"email": fmt.Sprintf("v%d@test.dev", uniq()), "password": testPassword, "fullName": "   "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := post(t, "", "/auth/register", tc.body).want(t, http.StatusBadRequest)
			if decode[errorJSON](t, r).Message == "" {
				t.Fatalf("400 без понятного сообщения: %s", r.Body)
			}
		})
	}
}

// Длина полей ограничена в БД (varchar(255)), но не проверяется в коде:
// слишком длинное имя долетает до INSERT и возвращается как 500.
func TestRegisterRejectsTooLongNameWith400(t *testing.T) {
	requireDB(t)

	r := post(t, "", "/auth/register", map[string]any{
		"email":    fmt.Sprintf("long%d@test.dev", uniq()),
		"password": testPassword,
		"fullName": strings.Repeat("Я", 256),
	})
	r.want(t, http.StatusBadRequest)
}

func TestLogin(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Входящий")

	t.Run("верный пароль, email в другом регистре", func(t *testing.T) {
		r := post(t, "", "/auth/login", map[string]any{"email": strings.ToUpper(user.Email), "password": testPassword}).
			want(t, http.StatusOK)
		mustNotContainPassword(t, r)
		if auth := decode[authJSON](t, r); auth.User.ID != user.ID || auth.AccessToken == "" {
			t.Fatalf("вход вернул %s", r.Body)
		}
	})

	// Ответы для «нет такого пользователя» и «неверный пароль» совпадают —
	// иначе форму входа можно использовать для перебора зарегистрированных адресов.
	t.Run("неверный пароль и неизвестный email неразличимы", func(t *testing.T) {
		wrongPassword := post(t, "", "/auth/login", map[string]any{"email": user.Email, "password": "wrong-password1"}).
			want(t, http.StatusUnauthorized)
		unknownEmail := post(t, "", "/auth/login", map[string]any{"email": "ghost@test.dev", "password": testPassword}).
			want(t, http.StatusUnauthorized)

		if a, b := decode[errorJSON](t, wrongPassword).Message, decode[errorJSON](t, unknownEmail).Message; a != b {
			t.Fatalf("сообщения различаются: %q и %q", a, b)
		}
	})
}

func TestRefreshToken(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Обновляющий")

	r := post(t, "", "/auth/refresh", map[string]any{"refreshToken": user.RefreshToken}).want(t, http.StatusOK)
	access := decode[authJSON](t, r).AccessToken
	if access == "" {
		t.Fatalf("refresh не вернул accessToken: %s", r.Body)
	}
	get(t, access, "/users/me").want(t, http.StatusOK)

	t.Run("access-токен не годится для refresh", func(t *testing.T) {
		post(t, "", "/auth/refresh", map[string]any{"refreshToken": user.Token}).want(t, http.StatusUnauthorized)
	})
	t.Run("refresh-токен не пускает в API", func(t *testing.T) {
		get(t, user.RefreshToken, "/users/me").want(t, http.StatusUnauthorized)
	})
	t.Run("мусор вместо токена", func(t *testing.T) {
		post(t, "", "/auth/refresh", map[string]any{"refreshToken": "not-a-token"}).want(t, http.StatusUnauthorized)
	})
}

func TestAuthorizationHeader(t *testing.T) {
	requireDB(t)
	user := newUser(t, "Заголовочный")

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"без заголовка", "", http.StatusUnauthorized},
		{"не та схема", "Basic " + user.Token, http.StatusUnauthorized},
		{"пустой токен", "Bearer   ", http.StatusUnauthorized},
		{"мусорный токен", "Bearer abc.def.ghi", http.StatusUnauthorized},
		{"схема в нижнем регистре", "bearer " + user.Token, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newRawRequest(http.MethodGet, "/users/me", tc.header)
			resp, err := testApp.Test(req, 10_000)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Fatalf("код %d, ожидался %d", resp.StatusCode, tc.want)
			}
		})
	}
}

func TestMalformedJSONIs400(t *testing.T) {
	requireDB(t)
	sendRaw(t, http.MethodPost, "/auth/login", "", `{"email": "a@b.c", "password": `).want(t, http.StatusBadRequest)
}

func TestUsersSearch(t *testing.T) {
	requireDB(t)
	marker := fmt.Sprintf("Уникум%d", uniq())
	target := newUser(t, "Гаврила "+marker)

	search := func(q string) []userJSON {
		t.Helper()
		return decode[[]userJSON](t, get(t, target.Token, "/users?search="+urlQuery(q)).want(t, http.StatusOK))
	}
	found := func(users []userJSON) bool {
		for _, u := range users {
			if u.ID == target.ID {
				return true
			}
		}
		return false
	}

	if !found(search(strings.ToLower(marker))) {
		t.Fatal("поиск по имени в нижнем регистре (кириллица) не нашёл пользователя")
	}
	if !found(search(target.Email[:8])) {
		t.Fatal("поиск по части email не нашёл пользователя")
	}
	// "%" — спецсимвол LIKE. Без экранирования поиск "%" вернул бы всех подряд.
	if found(search("%")) {
		t.Fatal(`поиск "%" нашёл пользователя без "%" в имени — спецсимволы LIKE не экранируются`)
	}
}

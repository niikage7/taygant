package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"нормальный пароль", "hackaton2026", false},
		{"кириллица считается буквами", "пароль2026", false},
		{"слишком короткий", "abc1", true},
		{"без цифр", "onlyletters", true},
		{"без букв", "12345678", true},
		{"пустой", "", true},
		{"длиннее ограничения bcrypt", string(make([]byte, 100)), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidatePassword(%q) = %v, ожидалась ошибка: %v", tc.password, err, tc.wantErr)
			}
		})
	}
}

func TestHashPasswordRoundTrip(t *testing.T) {
	const password = "hackaton2026"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == password {
		t.Fatal("хеш совпал с открытым паролем")
	}
	if !VerifyPassword(hash, password) {
		t.Fatal("верный пароль не прошёл проверку")
	}
	if VerifyPassword(hash, password+"x") {
		t.Fatal("неверный пароль прошёл проверку")
	}
}

// Одинаковые пароли должны давать разные хеши: bcrypt подмешивает случайную соль,
// иначе по совпадающим хешам было бы видно, у кого из пользователей общий пароль.
func TestHashPasswordUsesSalt(t *testing.T) {
	first, err := HashPassword("hackaton2026")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	second, err := HashPassword("hackaton2026")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if first == second {
		t.Fatal("два хеша одного пароля совпали — соль не применяется")
	}
}

func newTestIssuer() *TokenIssuer {
	return NewTokenIssuer("test-secret", time.Hour, 24*time.Hour)
}

func TestIssueAndParse(t *testing.T) {
	issuer := newTestIssuer()
	userID := uuid.New()

	access, refresh, err := issuer.IssuePair(userID)
	if err != nil {
		t.Fatalf("IssuePair: %v", err)
	}

	got, err := issuer.Parse(access, KindAccess)
	if err != nil {
		t.Fatalf("Parse(access): %v", err)
	}
	if got != userID {
		t.Fatalf("Parse вернул %s, ожидался %s", got, userID)
	}

	if _, err := issuer.Parse(refresh, KindRefresh); err != nil {
		t.Fatalf("Parse(refresh): %v", err)
	}
}

// Ключевая проверка: долгоживущий refresh не должен открывать доступ к API.
func TestRefreshTokenRejectedAsAccess(t *testing.T) {
	issuer := newTestIssuer()

	refresh, err := issuer.Issue(uuid.New(), KindRefresh)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := issuer.Parse(refresh, KindAccess); err == nil {
		t.Fatal("refresh-токен принят как access")
	}
}

func TestParseRejectsForeignSignature(t *testing.T) {
	token, err := newTestIssuer().Issue(uuid.New(), KindAccess)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}

	other := NewTokenIssuer("другой-секрет", time.Hour, 24*time.Hour)
	if _, err := other.Parse(token, KindAccess); err == nil {
		t.Fatal("токен, подписанный чужим ключом, прошёл проверку")
	}
}

func TestParseRejectsExpired(t *testing.T) {
	// Отрицательный TTL даёт заведомо просроченный токен без ожидания в тесте.
	issuer := NewTokenIssuer("test-secret", -time.Minute, time.Hour)

	token, err := issuer.Issue(uuid.New(), KindAccess)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	if _, err := issuer.Parse(token, KindAccess); err == nil {
		t.Fatal("просроченный токен прошёл проверку")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	issuer := newTestIssuer()

	for _, token := range []string{"", "не-токен", "a.b.c"} {
		if _, err := issuer.Parse(token, KindAccess); err == nil {
			t.Fatalf("мусорный токен %q прошёл проверку", token)
		}
	}
}

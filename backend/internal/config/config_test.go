package config

import (
	"reflect"
	"testing"
	"time"
)

// Пустой список origin'ов Fiber трактует как "*": опечатка в ALLOWED_ORIGINS
// не должна молча открывать API всем сайтам.
func TestEnvOriginsNeverEmpty(t *testing.T) {
	t.Setenv("TEST_ALLOWED_ORIGINS", " , ")
	if got := envOrigins("TEST_ALLOWED_ORIGINS", "http://localhost:3000"); !reflect.DeepEqual(got, []string{"http://localhost:3000"}) {
		t.Fatalf("envOrigins = %v, ожидался дефолт", got)
	}
}

// Браузер присылает Origin без завершающего слэша.
func TestEnvOriginsTrimsTrailingSlash(t *testing.T) {
	t.Setenv("TEST_ALLOWED_ORIGINS", "http://localhost:3000/, https://taygant.dev/ ")
	want := []string{"http://localhost:3000", "https://taygant.dev"}
	if got := envOrigins("TEST_ALLOWED_ORIGINS", "x"); !reflect.DeepEqual(got, want) {
		t.Fatalf("envOrigins = %v, ожидалось %v", got, want)
	}
}

// TRUSTED_PROXIES="" означает «не доверять никому», а не «взять дефолт».
func TestEnvListDistinguishesEmptyFromUnset(t *testing.T) {
	t.Setenv("TEST_TRUSTED_PROXIES", "")
	if got := envList("TEST_TRUSTED_PROXIES", "127.0.0.1"); len(got) != 0 {
		t.Fatalf("явно пустая переменная дала %v", got)
	}
	if got := envList("TEST_VARIABLE_THAT_IS_NOT_SET", "127.0.0.1, ::1"); !reflect.DeepEqual(got, []string{"127.0.0.1", "::1"}) {
		t.Fatalf("незаданная переменная дала %v, ожидался дефолт", got)
	}
}

func TestInvalidNumbersFallBackToDefaults(t *testing.T) {
	for _, value := range []string{"abc", "0", "-5"} {
		t.Setenv("TEST_INT", value)
		if got := envInt("TEST_INT", 300); got != 300 {
			t.Errorf("envInt(%q) = %d, ожидался дефолт 300", value, got)
		}
		t.Setenv("TEST_DURATION", value)
		if got := envDuration("TEST_DURATION", time.Minute); got != time.Minute {
			t.Errorf("envDuration(%q) = %s, ожидался дефолт 1m", value, got)
		}
	}

	t.Setenv("TEST_INT", "42")
	t.Setenv("TEST_DURATION", "90s")
	t.Setenv("TEST_BOOL", "false")
	if envInt("TEST_INT", 1) != 42 || envDuration("TEST_DURATION", 0) != 90*time.Second || envBool("TEST_BOOL", true) {
		t.Fatal("корректные значения не прочитались")
	}
	t.Setenv("TEST_BOOL", "maybe")
	if !envBool("TEST_BOOL", true) {
		t.Fatal("нераспознанный bool не заменился дефолтом")
	}
}

func TestUsesDevJWTSecret(t *testing.T) {
	if !(Config{JWTSecret: devJWTSecret}).UsesDevJWTSecret() {
		t.Fatal("дефолтный ключ не распознан")
	}
	if (Config{JWTSecret: "настоящий-секрет"}).UsesDevJWTSecret() {
		t.Fatal("собственный ключ принят за дефолтный")
	}
}

package service

import (
	"context"
	"regexp"
	"testing"

	"taygant_backend/internal/models"
)

func TestProjectBufferDays(t *testing.T) {
	cases := []struct{ deviation, want int }{
		{-10, 10}, // прогноз на 10 дней раньше дедлайна — 10 дней запаса
		{0, 0},    // финиш ровно в дедлайн — запаса нет
		{5, 0},    // уже не успеваем — запаса нет, отставание видно в deviation
	}
	for _, tc := range cases {
		if got := projectBufferDays(tc.deviation); got != tc.want {
			t.Errorf("projectBufferDays(%d) = %d, ожидалось %d", tc.deviation, got, tc.want)
		}
	}
	// С запасом больше порога проект не «в зоне риска»: раньше резерв считался
	// как минимальный slack и всегда выходил 0.
	if status := adherenceStatus(-20, projectBufferDays(-20)); status != "on_track" {
		t.Fatalf("проект с запасом 20 дней: %s", status)
	}
}

func TestChangeSourceFromContext(t *testing.T) {
	if src := changeSourceFrom(context.Background()); src.Channel != models.HistorySourceApp || src.Via != "" {
		t.Fatalf("без пометки канал = %+v, ожидалось приложение", src)
	}
	ctx := WithChangeSource(context.Background(), ChangeSource{Channel: models.HistorySourceAssistant, Via: "Cursor"})
	if src := changeSourceFrom(ctx); src.Channel != models.HistorySourceAssistant || src.Via != "Cursor" {
		t.Fatalf("канал = %+v", src)
	}
	// Пустая пометка не должна превращать правку в «ничью».
	if src := changeSourceFrom(WithChangeSource(context.Background(), ChangeSource{})); src.Channel != models.HistorySourceApp {
		t.Fatalf("пустая пометка дала канал %+v", src)
	}
}

func TestMCPSecretFormat(t *testing.T) {
	pattern := regexp.MustCompile(`^tgn_[A-Z2-7]{26}$`)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		secret := newMCPSecret()
		if !pattern.MatchString(secret) {
			t.Fatalf("ключ %q не похож на tgn_ + 26 символов base32", secret)
		}
		if seen[secret] {
			t.Fatalf("ключ %q выпущен дважды", secret)
		}
		seen[secret] = true
	}

	hash := hashMCPSecret("tgn_EXAMPLE")
	if len(hash) != 64 || hash != hashMCPSecret("tgn_EXAMPLE") || hash == hashMCPSecret("tgn_EXAMPLF") {
		t.Fatalf("хеш %q: ожидались стабильные 64 hex-символа, разные для разных ключей", hash)
	}
}

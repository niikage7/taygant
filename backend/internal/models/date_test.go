package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDateJSONRoundTrip(t *testing.T) {
	in := NewDate(2026, time.September, 10)

	encoded, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(encoded) != `"2026-09-10"` {
		t.Fatalf("Marshal вернул %s, ожидалось \"2026-09-10\"", encoded)
	}

	var out Date
	if err := json.Unmarshal(encoded, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !out.Equal(in) {
		t.Fatalf("после round-trip получено %s, ожидалось %s", out, in)
	}
}

// Клиенты нередко присылают Date.toISOString() вместо чистой даты —
// такой ввод должен приниматься, а не отвергаться.
func TestDateUnmarshalAcceptsTimestamp(t *testing.T) {
	var d Date
	if err := json.Unmarshal([]byte(`"2026-09-10T21:45:00Z"`), &d); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if d.String() != "2026-09-10" {
		t.Fatalf("получено %s, ожидалось 2026-09-10", d)
	}
}

func TestDateUnmarshalNull(t *testing.T) {
	d := NewDate(2026, time.September, 10)
	if err := json.Unmarshal([]byte(`null`), &d); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !d.IsZero() {
		t.Fatalf("null должен давать пустую дату, получено %s", d)
	}
}

// Арифметика сроков строится на том, что разность дат кратна суткам.
// Проверяем переход через границу месяца и через перевод часов в марте.
func TestDateDaysUntil(t *testing.T) {
	cases := []struct {
		name string
		from Date
		to   Date
		want int
	}{
		{"тот же день", NewDate(2026, time.September, 10), NewDate(2026, time.September, 10), 0},
		{"через границу месяца", NewDate(2026, time.September, 28), NewDate(2026, time.October, 3), 5},
		{"назад", NewDate(2026, time.October, 3), NewDate(2026, time.September, 28), -5},
		{"через високосное 29 февраля", NewDate(2028, time.February, 28), NewDate(2028, time.March, 1), 2},
		{"через переход на летнее время", NewDate(2026, time.March, 28), NewDate(2026, time.March, 30), 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.from.DaysUntil(tc.to); got != tc.want {
				t.Fatalf("DaysUntil = %d, ожидалось %d", got, tc.want)
			}
		})
	}
}

func TestDateAddDays(t *testing.T) {
	got := NewDate(2026, time.December, 30).AddDays(5)
	if want := NewDate(2027, time.January, 4); !got.Equal(want) {
		t.Fatalf("AddDays = %s, ожидалось %s", got, want)
	}
}

func TestParseDateRejectsGarbage(t *testing.T) {
	for _, s := range []string{"", "10.09.2026", "2026-13-01", "не дата"} {
		if _, err := ParseDate(s); err == nil {
			t.Fatalf("ParseDate(%q) не вернул ошибку", s)
		}
	}
}

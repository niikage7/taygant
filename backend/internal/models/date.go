package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// DateLayout — формат календарной даты, которым оперирует спецификация (format: date).
const DateLayout = "2006-01-02"

// Date — календарная дата без времени и часового пояса.
//
// Сроки задач, спринтов и вех — это дни, а не моменты времени. Отдельный тип
// закрывает два класса ошибок, на которые легко напороться с обычным time.Time:
// сериализацию в RFC3339 (фронт ждёт "2026-09-10") и сдвиг даты на сутки при
// пересчёте часовых поясов между машиной разработчика и контейнером.
//
// Значение всегда нормализовано к полуночи UTC, поэтому разность двух дат
// делится на 24 часа без остатка — на этом строится вся арифметика сроков.
type Date struct {
	time.Time
}

// NewDate собирает дату из компонентов.
func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// DateOf отбрасывает у момента времени всё, кроме календарного дня.
func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return NewDate(y, m, d)
}

// Today — сегодняшняя дата по UTC.
func Today() Date {
	return DateOf(time.Now().UTC())
}

// ParseDate разбирает дату в формате YYYY-MM-DD.
func ParseDate(s string) (Date, error) {
	t, err := time.ParseInLocation(DateLayout, strings.TrimSpace(s), time.UTC)
	if err != nil {
		return Date{}, fmt.Errorf("дата должна быть в формате %s: %w", DateLayout, err)
	}
	return Date{t}, nil
}

// IsZero сообщает, что дата не задана.
func (d Date) IsZero() bool { return d.Time.IsZero() }

// String возвращает дату в формате спецификации.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Format(DateLayout)
}

// AddDays сдвигает дату на указанное число календарных дней (может быть отрицательным).
func (d Date) AddDays(days int) Date {
	return Date{d.AddDate(0, 0, days)}
}

// DaysUntil возвращает число календарных дней от d до other.
// Положительное значение означает, что other позже d.
func (d Date) DaysUntil(other Date) int {
	return int(other.Sub(d.Time).Hours() / 24)
}

// Before сообщает, что d строго раньше other.
func (d Date) Before(other Date) bool { return d.Time.Before(other.Time) }

// After сообщает, что d строго позже other.
func (d Date) After(other Date) bool { return d.Time.After(other.Time) }

// Equal сообщает, что даты совпадают.
func (d Date) Equal(other Date) bool { return d.Time.Equal(other.Time) }

// MaxDate возвращает более позднюю из двух дат.
func MaxDate(a, b Date) Date {
	if a.After(b) {
		return a
	}
	return b
}

// MinDate возвращает более раннюю из двух дат.
func MinDate(a, b Date) Date {
	if a.Before(b) {
		return a
	}
	return b
}

// MarshalJSON отдаёт дату строкой "2006-01-02"; незаданная дата становится null.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format(DateLayout) + `"`), nil
}

// UnmarshalJSON принимает как чистую дату, так и полный timestamp:
// клиенты иногда присылают Date.toISOString(), и отвергать такой ввод незачем.
func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		*d = Date{}
		return nil
	}
	if parsed, err := ParseDate(s); err == nil {
		*d = parsed
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return fmt.Errorf("дата должна быть в формате %s, получено %q", DateLayout, s)
	}
	*d = DateOf(t)
	return nil
}

// Value сохраняет дату в колонку типа date.
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

// Scan читает дату из БД. Драйвер отдаёт date как time.Time, но текстовые
// представления тоже поддерживаем — они приходят из ручных SQL-выборок.
func (d *Date) Scan(value any) error {
	switch v := value.(type) {
	case nil:
		*d = Date{}
	case time.Time:
		*d = DateOf(v)
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = parsed
	case []byte:
		parsed, err := ParseDate(string(v))
		if err != nil {
			return err
		}
		*d = parsed
	default:
		return fmt.Errorf("невозможно прочитать %T как дату", value)
	}
	return nil
}

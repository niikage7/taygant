// Package database отвечает за подключение к PostgreSQL и применение схемы.
// Остальные пакеты получают готовый *gorm.DB и про драйвер ничего не знают.
package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect открывает пул соединений к PostgreSQL и дожидается его готовности.
//
// gorm.Open только конструирует объект — реальное соединение проверяется отдельным
// Ping'ом. В docker compose backend стартует одновременно с postgres, поэтому первые
// попытки закономерно падают: ждём до waitFor, а не падаем на первой ошибке.
func Connect(dsn string, waitFor time.Duration) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Дефолтный логгер GORM печатает каждый запрос; на хакатоне это забивает
		// вывод и мешает читать лог Fiber. Оставляем только ошибки и медленные запросы.
		Logger: logger.Default.LogMode(logger.Warn),
		// Все даты в API — календарные, без часовых поясов. Единый UTC внутри
		// избавляет от расхождений между машиной разработчика и контейнером.
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("открыть подключение к postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("получить пул соединений: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := waitReady(sqlDB.Ping, waitFor); err != nil {
		return nil, err
	}

	return db, nil
}

// waitReady повторяет ping, пока БД не ответит или не истечёт таймаут.
func waitReady(ping func() error, waitFor time.Duration) error {
	const retryInterval = time.Second

	deadline := time.Now().Add(waitFor)
	var lastErr error

	for attempt := 1; ; attempt++ {
		if lastErr = ping(); lastErr == nil {
			return nil
		}
		if time.Now().Add(retryInterval).After(deadline) {
			return fmt.Errorf("postgres не ответил за %s: %w", waitFor, lastErr)
		}
		log.Printf("postgres ещё не готов (попытка %d): %v", attempt, lastErr)
		time.Sleep(retryInterval)
	}
}

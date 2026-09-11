package main

import (
	"log"
	"time"

	"taygant_backend/internal/config"
	"taygant_backend/internal/database"
	"taygant_backend/internal/server"
)

// dbStartupTimeout — сколько ждём готовности postgres при старте.
// В docker compose backend поднимается вместе с БД, первые ping'и штатно падают.
const dbStartupTimeout = 30 * time.Second

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL, dbStartupTimeout)
	if err != nil {
		log.Fatalf("подключение к postgres: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("миграция схемы: %v", err)
	}
	log.Println("postgres подключён, схема применена")

	if cfg.SeedDemoData {
		if err := database.Seed(db); err != nil {
			log.Fatalf("сидинг демо-данных: %v", err)
		}
	}

	app := server.New(cfg, db)

	log.Printf("taygant api listening on :%s (allowed origins: %v)", cfg.Port, cfg.AllowedOrigins)
	log.Fatal(app.Listen(":" + cfg.Port))
}

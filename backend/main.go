package main

import (
	"log"

	"taygant_backend/internal/config"
	"taygant_backend/internal/server"
)

func main() {
	cfg := config.Load()
	app := server.New(cfg)

	log.Printf("taygant api listening on :%s (allowed origins: %v)", cfg.Port, cfg.AllowedOrigins)
	log.Fatal(app.Listen(":" + cfg.Port))
}

package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	//"github.com/gofiber/fiber/v2/middleware/cors"

	"taygant_backend/internal/api"
)

func main() {
	fmt.Println("Start main.go")

	app := fiber.New()

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	// Все эндпоинты из api-spec.yml, смонтированные на /api/v1.
	api.Register(app)

	app.Listen(":4000")
}

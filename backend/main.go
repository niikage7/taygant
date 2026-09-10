package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	//"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	fmt.Println("Start main.go")

	app := fiber.New()

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	
	app.Listen(":4000")
}
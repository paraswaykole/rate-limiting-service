package main

import (
	"log"
	"net/http"
	"rate-limiting-service/internal/config"
	"rate-limiting-service/internal/storage"

	"github.com/gofiber/fiber/v3"
)

func main() {
	storage.GetManager()
	app := fiber.New()
	app.Post("/check", func(c fiber.Ctx) error {
		allowed := true
		if allowed {
			c.Status(http.StatusOK)
			return nil
		}
		c.Status(http.StatusTooManyRequests)
		return nil
	})
	app.Post("/configure", func(c fiber.Ctx) error {
		return c.SendString("configured")
	})
	app.Get("/metrics", func(c fiber.Ctx) error {
		return nil
	})
	log.Fatal(app.Listen(":" + config.PORT))
}

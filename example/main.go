package main

import (
	"math/rand"
	"rate-limiting-service/pkg/sdk"
	"time"

	"github.com/gofiber/fiber/v3"
)

var userids = []string{"1234", "7891", "9990", "1123"}

var users = map[string]any{
	"1234": "Rob",
	"7891": "Alice",
	"9990": "Meet",
	"1123": "Rohan",
}

func main() {
	app := fiber.New()

	// mock userid header for simulating tests
	app.Use(func(c fiber.Ctx) error {
		min := 0
		max := len(userids)
		randomI := rand.Intn(max-min) + min
		userid := userids[randomI]
		c.Request().Header.Add("X-User-Id", userid)
		return c.Next()
	})

	app.Use(sdk.Middleware(sdk.Config{
		CheckURL: "http://localhost:3123/check",
		Timeout:  200 * time.Millisecond, // 200ms
		FailOpen: false,
		Key:      "helloworldservice",
		ArgsExtractor: func(c fiber.Ctx) (args []string, err error) {
			userid := c.Get("X-User-Id", "")
			return []string{c.Path(), userid}, nil
		},
	}))

	app.Get("/hello", func(c fiber.Ctx) error {
		userid := c.Get("X-User-Id", "")
		return c.SendString("hello " + users[userid].(string))
	})

	_ = app.Listen(":8080")
}

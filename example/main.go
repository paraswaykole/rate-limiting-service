package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"rate-limiting-service/internal/config"
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

	rateLimiterHost := config.GetConfig("RATE_LIMITER_HOST", "http://localhost:3123")
	setupRateLimiter(rateLimiterHost)

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

	middleware1 := sdk.Middleware(sdk.Config{
		CheckURL: rateLimiterHost + "/check",
		Timeout:  200 * time.Millisecond, // 200ms
		FailOpen: false,
		Key:      "api1",
		ArgsExtractor: func(c fiber.Ctx) (args []string, err error) {
			userid := c.Get("X-User-Id", "")
			return []string{c.Path(), userid}, nil
		},
	})

	middleware2 := sdk.Middleware(sdk.Config{
		CheckURL: rateLimiterHost + "/check",
		Timeout:  200 * time.Millisecond, // 200ms
		FailOpen: false,
		Key:      "api2",
		ArgsExtractor: func(c fiber.Ctx) (args []string, err error) {
			userid := c.Get("X-User-Id", "")
			return []string{c.Path(), userid}, nil
		},
	})

	api1 := app.Group("/api1", middleware1)
	{
		api1.Get("/hello", func(c fiber.Ctx) error {
			userid := c.Get("X-User-Id", "")
			return c.SendString("hello " + users[userid].(string))
		})
	}

	api2 := app.Group("/api2", middleware2)
	{
		api2.Get("/hello", func(c fiber.Ctx) error {
			userid := c.Get("X-User-Id", "")
			return c.SendString("hello " + users[userid].(string))
		})
	}

	_ = app.Listen(":8080")
}

func setupRateLimiter(rateLimiterHost string) {
	url := rateLimiterHost + "/configure"
	payload := []byte(`{
        "key": "api1",
        "limiterType": 20,
        "configuration": {
            "capacity": 2,
            "windowSize": 5
        }
    }`)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)

	payload = []byte(`{
		"key": "api2",
	    "limiterType": 20,
	    "configuration": {
	      "capacity": 10,
	      "refillRate": 5
	    }
    }`)

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
}

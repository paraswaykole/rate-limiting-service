package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rate-limiting-service/internal/config"
	"rate-limiting-service/internal/limiter"
	"rate-limiting-service/internal/services"
	"rate-limiting-service/internal/storage"
	"rate-limiting-service/internal/utils"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

func main() {
	storage.GetManager()
	startSyncJob()
	startServer()
}

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func startServer() {
	app := fiber.New(fiber.Config{
		StructValidator: &structValidator{validate: validator.New()},
	})

	app.Use(func(c fiber.Ctx) error {
		fmt.Println(c.Path(), c.Method())
		if c.Path() == "/check" && c.Method() == "GET" {
			t1 := time.Now()
			nextErr := c.Next()
			allowed := c.Response().StatusCode() != http.StatusTooManyRequests
			latency := time.Since(t1)
			go services.UpdateMetrics(allowed, latency)
			return nextErr
		}
		return c.Next()
	})

	app.Get("/check", func(c fiber.Ctx) error {
		checkDto := new(services.CheckDTO)
		if err := c.Bind().Query(checkDto); err != nil {
			utils.SendValidationErrors(err, c)
			return err
		}
		allowed, headers, err := services.Check(checkDto)
		if err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				utils.SendValidationErrors(validationErrors, c)
				return nil
			}
		}
		if err != nil {
			if err.Error() == "rate limiter not found" {
				return c.Status(http.StatusNotFound).SendString("rate limiter not found")
			}
			return c.Status(http.StatusInternalServerError).SendString("Internal server error")
		}
		for key, value := range headers {
			c.Response().Header.Add(key, value)
		}
		if allowed {
			c.Status(http.StatusOK)
			return nil
		}
		c.Status(http.StatusTooManyRequests)
		return nil
	})
	app.Post("/configure", func(c fiber.Ctx) error {
		configDto := new(services.ConfigureDTO)
		if err := c.Bind().Body(configDto); err != nil {
			utils.SendValidationErrors(err, c)
			return err
		}
		err := services.Configure(configDto)
		if err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				utils.SendValidationErrors(validationErrors, c)
				return nil
			}
		}
		if err != nil {
			fmt.Println("/configure error:", err)
			return c.Status(http.StatusInternalServerError).SendString("Internal server error")
		}
		return c.SendString("configured")
	})
	app.Get("/metrics", func(c fiber.Ctx) error {
		reset := c.Query("reset", "")
		if reset == "true" {
			services.ResetMetrics()
			return c.SendStatus(http.StatusNoContent)
		}
		response := services.GetMetrics()
		return c.JSON(response)
	})

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + config.PORT); err != nil {
			log.Fatal("Fiber stopped:", err)
		}
	}()

	<-quit
	fmt.Println("Shutting down...")
	limiter.GetManager().StopAll()
	app.Shutdown()
	fmt.Println("Shutdown complete.")
}

func startSyncJob() {
	go func() {
		ticker := time.NewTicker(10 * time.Millisecond)
		for range ticker.C {
			limiter.GetManager().SyncLimiters()
		}
	}()
}

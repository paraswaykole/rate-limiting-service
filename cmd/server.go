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
	app.Get("/check", func(c fiber.Ctx) error {
		checkDto := new(services.CheckDTO)
		if err := c.Bind().Query(checkDto); err != nil {
			utils.SendValidationErrors(err, c)
			return err
		}
		allowed, err := services.Check(checkDto)
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
		return nil
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

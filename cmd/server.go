package main

import (
	"fmt"
	"log"
	"net/http"
	"rate-limiting-service/internal/config"
	"rate-limiting-service/internal/services"
	"rate-limiting-service/internal/storage"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type structValidator struct {
	validate *validator.Validate
}

func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func main() {
	storage.GetManager()
	app := fiber.New(fiber.Config{
		StructValidator: &structValidator{validate: validator.New()},
	})
	app.Get("/check", func(c fiber.Ctx) error {
		checkDto := new(services.CheckDTO)
		if err := c.Bind().Query(checkDto); err != nil {
			sendValidationErrors(err, c)
			return err
		}
		allowed, err := services.Check(checkDto)
		if err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				sendValidationErrors(validationErrors, c)
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
			sendValidationErrors(err, c)
			return err
		}
		err := services.Configure(configDto)
		if err != nil {
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				sendValidationErrors(validationErrors, c)
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
	log.Fatal(app.Listen(":" + config.PORT))
}

func sendValidationErrors(err error, c fiber.Ctx) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"field": e.Field(),
				"error": e.Error(),
			})
		}
	}
}

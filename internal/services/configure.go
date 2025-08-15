package services

import (
	"encoding/json"
	"rate-limiting-service/internal/limiter"
)

type ConfigureDTO struct {
	Key           string              `json:"key" validate:"required" message:"Valid key is required"`
	LimiterType   limiter.LimiterType `json:"limiterType" validate:"required"`
	Configuration json.RawMessage     `json:"configuration" validate:"required" message:"configuration key is required"`
}

func Configure(configDTO *ConfigureDTO) error {

	rateLimiter := limiter.NewLimiter(configDTO.Key, []string{}, configDTO.LimiterType)

	err := rateLimiter.Configure(configDTO.Configuration)
	if err != nil {
		return err
	}
	return nil
}

package services

import (
	"errors"
	"rate-limiting-service/internal/limiter"
)

type CheckDTO struct {
	Key  string   `query:"key" validate:"required" message:"Valid key is required"`
	Args []string `query:"args"`
}

func Check(checkDTO *CheckDTO) (bool, error) {
	rateLimiter := limiter.GetManager().AccessLimiter(checkDTO.Key, checkDTO.Args)
	if rateLimiter == nil {
		return false, errors.New("rate limiter not found")
	}
	return (*rateLimiter).Check(), nil
}

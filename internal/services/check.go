package services

import (
	"errors"
	"rate-limiting-service/internal/limiter"
)

type CheckDTO struct {
	Key  string   `query:"key" validate:"required" message:"Valid key is required"`
	Args []string `query:"args"`
}

func Check(checkDTO *CheckDTO) (bool, map[string]string, error) {
	rateLimiter := limiter.GetManager().AccessLimiter(checkDTO.Key, checkDTO.Args)
	if rateLimiter == nil {
		return false, nil, errors.New("rate limiter not found")
	}
	allowed, headers := (*rateLimiter).Check()
	return allowed, headers, nil
}

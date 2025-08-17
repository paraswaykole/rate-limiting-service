package config

import "rate-limiting-service/internal/utils"

var (
	PORT           = GetConfig("PORT", "3123")
	REDIS_ADDRESS  = GetConfig("REDIS_ADDRESS", "localhost:6379")
	REDIS_PASSWORD = GetConfig("REDIS_PASSWORD", "")
)

var (
	RATE_LIMITING_INSTANCE_ID = utils.RandomString(15)
)

const (
	SYNC_LIMITER_FREQUENCY_TIME_IN_MS = 15
)

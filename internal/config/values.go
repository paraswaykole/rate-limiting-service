package config

var (
	PORT           = GetConfig("PORT", "3000")
	REDIS_ADDRESS  = GetConfig("REDIS_ADDRESS", "localhost:6379")
	REDIS_PASSWORD = GetConfig("REDIS_PASSWORD", "")
)

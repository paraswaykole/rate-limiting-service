package config

import (
	"os"
	"strconv"
)

func GetConfig(name string, defaultValue string) string {
	val := os.Getenv(string(name))
	if len(val) == 0 {
		return defaultValue
	}
	return val
}

func GetIntConfig(name string, defaultValue int) int {
	value := os.Getenv(string(name))
	if intVal, err := strconv.Atoi(value); err != nil {
		return intVal
	}
	return defaultValue
}

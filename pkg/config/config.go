package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	RedisURL       string
	ServerPort     string
	RateLimitRate  int
	RateLimitBurst int
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/promoengine?sslmode=disable"),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379/0"),
		ServerPort:     getEnv("SERVER_PORT", ":8080"),
		RateLimitRate:  getEnvInt("RATE_LIMIT_RATE", 5),
		RateLimitBurst: getEnvInt("RATE_LIMIT_BURST", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

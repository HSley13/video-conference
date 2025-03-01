package config

import (
	"os"
	"strconv"
)

type Config struct {
	Dsn           string
	RedisURL      string
	JWTSecret     string
	Port          string
	JWTAccessExp  int
	JWTRefreshExp int
}

func Load() *Config {
	return &Config{
		Dsn:           getEnv("DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:     getEnv("JWT_SECRET", "supersecretkey"),
		Port:          getEnv("PORT", "3000"),
		JWTAccessExp:  getEnvAsInt("JWT_ACCESS_EXP", 15),
		JWTRefreshExp: getEnvAsInt("JWT_REFRESH_EXP", 1440),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

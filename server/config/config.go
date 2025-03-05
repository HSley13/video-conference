package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port             string
	JWTSecret        string
	AllowedOrigins   []string
	RedisURL         string
	PostgresDSN      string
	MaxConnections   int
	WebRTCIceServers []string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "3002"),
		JWTSecret:      getEnv("JWT_SECRET", "secret"),
		AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}, ","),
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		PostgresDSN:    getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		MaxConnections: getEnvAsInt("MAX_CONNECTIONS", 10),
		WebRTCIceServers: getEnvAsSlice("WEBRTC_ICE_SERVERS", []string{
			"stun:stun.l.google.com:19302",
			"stun:stun1.l.google.com:19302",
			"stun:stun2.l.google.com:19302",
			"stun:stun3.l.google.com:19302",
			"stun:stun4.l.google.com:19302",
		}, ","),
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

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseBool(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, sep string) []string {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	return strings.Split(strValue, sep)
}

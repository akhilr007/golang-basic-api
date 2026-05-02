package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Logger LoggerConfig
	Auth   AuthConfig
}

type ServerConfig struct {
	Port string
}

type DBConfig struct {
	URL string
}

type LoggerConfig struct {
	Level string
}

type AuthConfig struct {
	JWTSecret    string
	CookieSecure bool
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		DB: DBConfig{
			URL: mustGetEnv("DB_URL"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		Auth: AuthConfig{
			JWTSecret:    mustGetMinEnv("JWT_SECRET", 32),
			CookieSecure: getBoolEnv("COOKIE_SECURE", true),
		},
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func mustGetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		log.Fatalf("%s is required but not set", key)
	}
	return val
}

func mustGetMinEnv(key string, minLen int) string {
	val := mustGetEnv(key)
	if len(val) < minLen {
		log.Fatalf("%s must be at least %d characters", key, minLen)
	}
	return val
}

func getBoolEnv(key string, fallback bool) bool {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(val)
	if err != nil {
		log.Fatalf("%s must be a boolean", key)
	}

	return parsed
}

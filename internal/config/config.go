package config

import (
	"log"
	"os"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Logger LoggerConfig
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

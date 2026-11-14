package config

import (
	"os"
	"time"
)

type Config struct {
	HTTPPort          string
	DatabaseURL       string
	AppEnv            string
	DefaultUserHandle string
	CompileTimeout    time.Duration
	RunTimeout        time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		AppEnv:            getEnv("APP_ENV", "development"),
		DefaultUserHandle: getEnv("DEFAULT_USER_HANDLE", "demo"),
		CompileTimeout:    getEnvDuration("COMPILE_TIMEOUT", 20*time.Second),
		RunTimeout:        getEnvDuration("RUN_TIMEOUT", 3*time.Second),
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "file:bitforge.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

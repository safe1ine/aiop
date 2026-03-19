package config

import (
	"os"
)

type Config struct {
	Addr        string
	JWTSecret   string
	DBPath      string
	AdminUser   string
	AdminPass   string
	EnrollToken string // agents use this to self-register
}

func Load() *Config {
	return &Config{
		Addr:        getEnv("AIPO_ADDR", ":8080"),
		JWTSecret:   getEnv("AIPO_JWT_SECRET", "change-me-in-production"),
		DBPath:      getEnv("AIPO_DB_PATH", "aipo.db"),
		AdminUser:   getEnv("AIPO_ADMIN_USER", "admin"),
		AdminPass:   getEnv("AIPO_ADMIN_PASS", "admin"),
		EnrollToken: getEnv("AIPO_ENROLL_TOKEN", "enroll-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

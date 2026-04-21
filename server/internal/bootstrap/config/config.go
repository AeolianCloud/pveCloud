package config

import (
	"errors"
	"os"
)

type Config struct {
	AppEnv         string
	PublicAPIAddr  string
	AdminAPIAddr   string
	WorkerAddr     string
	MySQLDSN       string
	RedisAddr      string
	JWTWebSecret   string
	JWTAdminSecret string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:         getenvDefault("APP_ENV", "local"),
		PublicAPIAddr:  getenvDefault("PUBLIC_API_ADDR", ":8080"),
		AdminAPIAddr:   getenvDefault("ADMIN_API_ADDR", ":8081"),
		WorkerAddr:     getenvDefault("WORKER_ADDR", ":8082"),
		MySQLDSN:       os.Getenv("MYSQL_DSN"),
		RedisAddr:      os.Getenv("REDIS_ADDR"),
		JWTWebSecret:   os.Getenv("JWT_WEB_SECRET"),
		JWTAdminSecret: os.Getenv("JWT_ADMIN_SECRET"),
	}

	switch {
	case cfg.MySQLDSN == "":
		return Config{}, errors.New("MYSQL_DSN is required")
	case cfg.RedisAddr == "":
		return Config{}, errors.New("REDIS_ADDR is required")
	case cfg.JWTWebSecret == "":
		return Config{}, errors.New("JWT_WEB_SECRET is required")
	case cfg.JWTAdminSecret == "":
		return Config{}, errors.New("JWT_ADMIN_SECRET is required")
	}

	return cfg, nil
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

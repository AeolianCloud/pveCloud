package config

import (
	"errors"
	"os"
)

type Config struct {
	AppEnv    string
	MySQLDSN  string
	RedisAddr string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:    os.Getenv("APP_ENV"),
		MySQLDSN:  os.Getenv("MYSQL_DSN"),
		RedisAddr: os.Getenv("REDIS_ADDR"),
	}
	if cfg.MySQLDSN == "" {
		return Config{}, errors.New("MYSQL_DSN is required")
	}
	return cfg, nil
}

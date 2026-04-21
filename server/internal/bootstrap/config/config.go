package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config/config.yaml"

type Config struct {
	AppEnv         string `yaml:"app_env"`
	PublicAPIAddr  string `yaml:"public_api_addr"`
	AdminAPIAddr   string `yaml:"admin_api_addr"`
	WorkerAddr     string `yaml:"worker_addr"`
	MySQLDSN       string `yaml:"mysql_dsn"`
	RedisAddr      string `yaml:"redis_addr"`
	JWTWebSecret   string `yaml:"jwt_web_secret"`
	JWTAdminSecret string `yaml:"jwt_admin_secret"`
}

func Load() (Config, error) {
	return LoadFrom(defaultConfigPath)
}

func LoadFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	switch {
	case cfg.AppEnv == "":
		return Config{}, errors.New("app_env is required")
	case cfg.PublicAPIAddr == "":
		return Config{}, errors.New("public_api_addr is required")
	case cfg.AdminAPIAddr == "":
		return Config{}, errors.New("admin_api_addr is required")
	case cfg.WorkerAddr == "":
		return Config{}, errors.New("worker_addr is required")
	case cfg.MySQLDSN == "":
		return Config{}, errors.New("mysql_dsn is required")
	case cfg.RedisAddr == "":
		return Config{}, errors.New("redis_addr is required")
	case cfg.JWTWebSecret == "":
		return Config{}, errors.New("jwt_web_secret is required")
	case cfg.JWTAdminSecret == "":
		return Config{}, errors.New("jwt_admin_secret is required")
	}

	return cfg, nil
}

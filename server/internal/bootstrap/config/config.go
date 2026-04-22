package config

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "config/config.yaml"

type PaymentConfig struct {
	Provider        string `yaml:"provider"`
	CallbackBaseURL string `yaml:"callback_base_url"`
	NotifyPath      string `yaml:"notify_path"`
	MerchantID      string `yaml:"merchant_id"`
	MerchantSecret  string `yaml:"merchant_secret"`
}

type ResourceConfig struct {
	Provider    string `yaml:"provider"`
	APIEndpoint string `yaml:"api_endpoint"`
	APIToken    string `yaml:"api_token"`
}

type WorkerConfig struct {
	PollInterval string `yaml:"poll_interval"`
	BatchSize    int    `yaml:"batch_size"`
}

type Config struct {
	AppEnv         string         `yaml:"app_env"`
	PublicAPIAddr  string         `yaml:"public_api_addr"`
	AdminAPIAddr   string         `yaml:"admin_api_addr"`
	WorkerAddr     string         `yaml:"worker_addr"`
	MySQLDSN       string         `yaml:"mysql_dsn"`
	RedisAddr      string         `yaml:"redis_addr"`
	JWTWebSecret   string         `yaml:"jwt_web_secret"`
	JWTAdminSecret string         `yaml:"jwt_admin_secret"`
	Payment        PaymentConfig  `yaml:"payment"`
	Resource       ResourceConfig `yaml:"resource"`
	Worker         WorkerConfig   `yaml:"worker"`
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

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	switch {
	case strings.TrimSpace(cfg.AppEnv) == "":
		return errors.New("app_env is required")
	case strings.TrimSpace(cfg.PublicAPIAddr) == "":
		return errors.New("public_api_addr is required")
	case strings.TrimSpace(cfg.AdminAPIAddr) == "":
		return errors.New("admin_api_addr is required")
	case strings.TrimSpace(cfg.WorkerAddr) == "":
		return errors.New("worker_addr is required")
	case strings.TrimSpace(cfg.MySQLDSN) == "":
		return errors.New("mysql_dsn is required")
	case strings.TrimSpace(cfg.RedisAddr) == "":
		return errors.New("redis_addr is required")
	case strings.TrimSpace(cfg.JWTWebSecret) == "":
		return errors.New("jwt_web_secret is required")
	case strings.TrimSpace(cfg.JWTAdminSecret) == "":
		return errors.New("jwt_admin_secret is required")
	}

	return nil
}

func (cfg Config) ValidateForService(serviceName string) error {
	switch serviceName {
	case "public-api":
		return cfg.Payment.Validate()
	case "admin-api":
		return nil
	case "worker":
		if err := cfg.Resource.Validate(); err != nil {
			return err
		}
		return cfg.Worker.Validate()
	default:
		return errors.New("unsupported service name")
	}
}

func (cfg PaymentConfig) Validate() error {
	switch {
	case strings.TrimSpace(cfg.Provider) == "":
		return errors.New("payment.provider is required")
	case strings.TrimSpace(cfg.CallbackBaseURL) == "":
		return errors.New("payment.callback_base_url is required")
	case strings.TrimSpace(cfg.NotifyPath) == "":
		return errors.New("payment.notify_path is required")
	case strings.TrimSpace(cfg.MerchantID) == "":
		return errors.New("payment.merchant_id is required")
	case strings.TrimSpace(cfg.MerchantSecret) == "":
		return errors.New("payment.merchant_secret is required")
	}

	return nil
}

func (cfg ResourceConfig) Validate() error {
	switch {
	case strings.TrimSpace(cfg.Provider) == "":
		return errors.New("resource.provider is required")
	case strings.TrimSpace(cfg.APIEndpoint) == "":
		return errors.New("resource.api_endpoint is required")
	case strings.TrimSpace(cfg.APIToken) == "":
		return errors.New("resource.api_token is required")
	}

	return nil
}

func (cfg WorkerConfig) Validate() error {
	switch {
	case strings.TrimSpace(cfg.PollInterval) == "":
		return errors.New("worker.poll_interval is required")
	case cfg.BatchSize <= 0:
		return errors.New("worker.batch_size must be greater than zero")
	}

	if _, err := time.ParseDuration(strings.TrimSpace(cfg.PollInterval)); err != nil {
		return errors.New("worker.poll_interval must be a valid duration")
	}

	return nil
}

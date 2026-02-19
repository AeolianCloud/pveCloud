package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 定义服务运行所需的全部配置项，避免在业务代码中散落硬编码配置。
type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	JWT struct {
		Secret               string `yaml:"secret"`
		AccessTokenExpireH   int    `yaml:"access_token_expire_hours"`
		RefreshTokenExpireDH int    `yaml:"refresh_token_expire_days"`
	} `yaml:"jwt"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	PVEClientMode string `yaml:"pve_client_mode"`
	PVE           struct {
		BaseURL                 string `yaml:"base_url"`
		APIKey                  string `yaml:"api_key"`
		APISecret               string `yaml:"api_secret"`
		TimeoutSeconds          int    `yaml:"timeout_seconds"`
		MaxRetries              int    `yaml:"max_retries"`
		RetryBackoffMS          int    `yaml:"retry_backoff_ms"`
		CircuitFailureThreshold int    `yaml:"circuit_failure_threshold"`
		CircuitOpenSeconds      int    `yaml:"circuit_open_seconds"`
	} `yaml:"pve"`
}

// Load 读取 YAML 配置文件并完成基础校验，保证主流程启动前即可发现配置缺失。
func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("parse config yaml: %w", err)
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.JWT.AccessTokenExpireH == 0 {
		cfg.JWT.AccessTokenExpireH = 2
	}
	if cfg.JWT.RefreshTokenExpireDH == 0 {
		cfg.JWT.RefreshTokenExpireDH = 7
	}
	if cfg.PVEClientMode == "" {
		cfg.PVEClientMode = "mock"
	}

	if cfg.PVE.TimeoutSeconds <= 0 {
		cfg.PVE.TimeoutSeconds = 8
	}
	if cfg.PVE.MaxRetries < 0 {
		cfg.PVE.MaxRetries = 0
	}
	if cfg.PVE.RetryBackoffMS <= 0 {
		cfg.PVE.RetryBackoffMS = 200
	}
	if cfg.PVE.CircuitFailureThreshold <= 0 {
		cfg.PVE.CircuitFailureThreshold = 5
	}
	if cfg.PVE.CircuitOpenSeconds <= 0 {
		cfg.PVE.CircuitOpenSeconds = 30
	}

	return &cfg, nil
}

package bootstrap

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Worker   WorkerConfig   `yaml:"worker"`
	Log      LogConfig      `yaml:"log"`
}

type AppConfig struct {
	Name                   string `yaml:"name"`
	Env                    string `yaml:"env"`
	Addr                   string `yaml:"addr"`
	ShutdownTimeoutSeconds int    `yaml:"shutdown_timeout_seconds"`
}

type DatabaseConfig struct {
	Host                   string `yaml:"host"`
	Port                   int    `yaml:"port"`
	Name                   string `yaml:"name"`
	User                   string `yaml:"user"`
	Password               string `yaml:"password"`
	Charset                string `yaml:"charset"`
	ParseTime              bool   `yaml:"parse_time"`
	Loc                    string `yaml:"loc"`
	MaxOpenConns           int    `yaml:"max_open_conns"`
	MaxIdleConns           int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `yaml:"conn_max_lifetime_minutes"`
}

type JWTConfig struct {
	UserSecret         string `yaml:"user_secret"`
	AdminSecret        string `yaml:"admin_secret"`
	UserIssuer         string `yaml:"user_issuer"`
	AdminIssuer        string `yaml:"admin_issuer"`
	UserExpireMinutes  int    `yaml:"user_expire_minutes"`
	AdminExpireMinutes int    `yaml:"admin_expire_minutes"`
}

type WorkerConfig struct {
	ID                  string `yaml:"id"`
	PollIntervalSeconds int    `yaml:"poll_interval_seconds"`
	LockTTLSeconds      int    `yaml:"lock_ttl_seconds"`
	BatchSize           int    `yaml:"batch_size"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config %s: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:                   "pveCloud",
			Env:                    "local",
			Addr:                   ":8080",
			ShutdownTimeoutSeconds: 10,
		},
		Database: DatabaseConfig{
			Host:                   "127.0.0.1",
			Port:                   3306,
			Name:                   "pvecloud",
			User:                   "pvecloud",
			Charset:                "utf8mb4",
			ParseTime:              true,
			Loc:                    "Local",
			MaxOpenConns:           25,
			MaxIdleConns:           10,
			ConnMaxLifetimeMinutes: 60,
		},
		JWT: JWTConfig{
			UserIssuer:         "pvecloud-user",
			AdminIssuer:        "pvecloud-admin",
			UserExpireMinutes:  1440,
			AdminExpireMinutes: 480,
		},
		Worker: WorkerConfig{
			ID:                  "local-worker-1",
			PollIntervalSeconds: 5,
			LockTTLSeconds:      60,
			BatchSize:           10,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}

func (cfg *Config) Validate() error {
	if cfg.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if cfg.JWT.UserSecret == "" || cfg.JWT.AdminSecret == "" {
		return fmt.Errorf("jwt.user_secret and jwt.admin_secret are required")
	}

	return nil
}

func (cfg DatabaseConfig) DSN() string {
	loc := url.QueryEscape(cfg.Loc)
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Charset,
		cfg.ParseTime,
		loc,
	)
}

func (cfg AppConfig) ShutdownTimeout() time.Duration {
	return time.Duration(cfg.ShutdownTimeoutSeconds) * time.Second
}

func (cfg WorkerConfig) PollInterval() time.Duration {
	return time.Duration(cfg.PollIntervalSeconds) * time.Second
}

func (cfg WorkerConfig) LockTTL() time.Duration {
	return time.Duration(cfg.LockTTLSeconds) * time.Second
}

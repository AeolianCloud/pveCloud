package bootstrap

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

/**
 * Config 表示后端 YAML 配置的根结构。
 */
type Config struct {
	App      AppConfig      `yaml:"app"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Worker   WorkerConfig   `yaml:"worker"`
	OpenAPI  OpenAPIConfig  `yaml:"openapi"`
	Log      LogConfig      `yaml:"log"`
}

/**
 * AppConfig 表示应用基础运行配置。
 */
type AppConfig struct {
	Name                   string `yaml:"name"`
	Env                    string `yaml:"env"`
	Addr                   string `yaml:"addr"`
	ShutdownTimeoutSeconds int    `yaml:"shutdown_timeout_seconds"`
}

/**
 * DatabaseConfig 表示 MariaDB 连接和连接池配置。
 */
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

/**
 * RedisConfig 表示 Redis 连接和项目 key 前缀配置。
 */
type RedisConfig struct {
	Addr      string `yaml:"addr"`
	Password  string `yaml:"password"`
	DB        int    `yaml:"db"`
	KeyPrefix string `yaml:"key_prefix"`
}

/**
 * JWTConfig 表示管理端 JWT 配置。
 * User* 字段仅为兼容历史本地配置保留，当前运行时不再使用。
 */
type JWTConfig struct {
	UserSecret         string `yaml:"user_secret"`
	AdminSecret        string `yaml:"admin_secret"`
	UserIssuer         string `yaml:"user_issuer"`
	AdminIssuer        string `yaml:"admin_issuer"`
	UserExpireMinutes  int    `yaml:"user_expire_minutes"`
	AdminExpireMinutes int    `yaml:"admin_expire_minutes"`
}

/**
 * WorkerConfig 仅为兼容历史本地配置保留，当前运行时不再使用。
 */
type WorkerConfig struct {
	ID                  string `yaml:"id"`
	PollIntervalSeconds int    `yaml:"poll_interval_seconds"`
	LockTTLSeconds      int    `yaml:"lock_ttl_seconds"`
	BatchSize           int    `yaml:"batch_size"`
}

/**
 * OpenAPIConfig 兼容历史配置项。
 * 项目不再维护 OpenAPI 生成文件，运行时不会读取该配置。
 */
type OpenAPIConfig struct {
	Enabled  bool   `yaml:"enabled"`
	SpecPath string `yaml:"spec_path"`
}

/**
 * LogConfig 表示系统日志配置。
 */
type LogConfig struct {
	Level string `yaml:"level"`
}

/**
 * LoadConfig 读取并校验 YAML 配置文件。
 *
 * @param path YAML 配置文件路径；为空时使用 config.yaml
 * @return *Config 已合并默认值并通过校验的配置
 * @return error 读取、解析或校验失败原因
 */
func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败：%s：%w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%s：%w", path, err)
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
		Redis: RedisConfig{
			Addr:      "127.0.0.1:6379",
			DB:        0,
			KeyPrefix: "pvecloud:",
		},
		JWT: JWTConfig{
			AdminIssuer:        "pvecloud-admin",
			AdminExpireMinutes: 480,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}

/**
 * Validate 校验配置是否具备启动所需的关键字段。
 *
 * @return error 配置不合法时返回具体原因
 */
func (cfg *Config) Validate() error {
	if cfg.App.Addr == "" {
		return fmt.Errorf("app.addr 不能为空")
	}
	if cfg.Database.Name == "" {
		return fmt.Errorf("database.name 不能为空")
	}
	if cfg.Database.User == "" {
		return fmt.Errorf("database.user 不能为空")
	}
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis.addr 不能为空")
	}
	if cfg.Redis.KeyPrefix == "" {
		return fmt.Errorf("redis.key_prefix 不能为空")
	}
	if cfg.JWT.AdminSecret == "" {
		return fmt.Errorf("jwt.admin_secret 不能为空")
	}
	return nil
}

/**
 * DSN 生成 GORM MySQL 驱动使用的 MariaDB 连接字符串。
 *
 * @return string MariaDB 连接字符串
 */
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

/**
 * ShutdownTimeout 返回 API 服务优雅退出的超时时间。
 *
 * @return time.Duration 优雅退出超时时间
 */
func (cfg AppConfig) ShutdownTimeout() time.Duration {
	return time.Duration(cfg.ShutdownTimeoutSeconds) * time.Second
}

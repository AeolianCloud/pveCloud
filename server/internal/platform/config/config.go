package config

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"strings"
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
	Mail     MailConfig     `yaml:"mail"`
	Log      LogConfig      `yaml:"log"`
	Storage  StorageConfig  `yaml:"storage"`
}

/**
 * AppConfig 表示应用基础运行配置。
 */
type AppConfig struct {
	Name                   string `yaml:"name"`
	Env                    string `yaml:"env"`
	Addr                   string `yaml:"addr"`
	ShutdownTimeoutSeconds int    `yaml:"shutdown_timeout_seconds"`
	Timezone               string `yaml:"timezone"`
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
 * MailConfig 表示用户端密码找回所需的 SMTP 邮件配置。
 */
type MailConfig struct {
	Enabled              bool   `yaml:"enabled"`
	Host                 string `yaml:"host"`
	Port                 int    `yaml:"port"`
	Username             string `yaml:"username"`
	Password             string `yaml:"password"`
	FromAddress          string `yaml:"from_address"`
	FromName             string `yaml:"from_name"`
	UseTLS               bool   `yaml:"use_tls"`
	PasswordResetURLBase string `yaml:"password_reset_url_base"`
}

/**
 * LogConfig 表示系统日志配置。
 */
type LogConfig struct {
	Level string `yaml:"level"`
}

/**
 * StorageConfig 表示文件上传与存储配置。
 */
type StorageConfig struct {
	Driver       string   `yaml:"driver"`
	LocalPath    string   `yaml:"local_path"`
	MaxSize      int64    `yaml:"max_size"`
	AllowedTypes []string `yaml:"allowed_types"`
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
	if err := cfg.ApplyTimezone(); err != nil {
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
			Timezone:               "Asia/Shanghai",
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
			UserIssuer:         "pvecloud-user",
			UserExpireMinutes:  480,
			AdminIssuer:        "pvecloud-admin",
			AdminExpireMinutes: 480,
		},
		Mail: MailConfig{
			Enabled:              false,
			Host:                 "smtp.example.com",
			Port:                 587,
			FromAddress:          "no-reply@example.com",
			FromName:             "pveCloud",
			UseTLS:               true,
			PasswordResetURLBase: "http://localhost:5174/reset-password",
		},
		Log: LogConfig{
			Level: "info",
		},
		Storage: StorageConfig{
			Driver:    "local",
			LocalPath: "./uploads",
			MaxSize:   10485760,
			AllowedTypes: []string{
				"image/jpeg",
				"image/png",
				"image/gif",
				"image/webp",
				"application/pdf",
			},
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
	if strings.TrimSpace(cfg.App.Timezone) == "" {
		return fmt.Errorf("app.timezone 不能为空")
	}
	if _, err := time.LoadLocation(cfg.App.Timezone); err != nil {
		return fmt.Errorf("app.timezone 必须是有效 IANA 时区名：%w", err)
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
	if err := validateJWTSecret("jwt.admin_secret", cfg.JWT.AdminSecret); err != nil {
		return err
	}
	if err := validateJWTSecret("jwt.user_secret", cfg.JWT.UserSecret); err != nil {
		return err
	}
	if cfg.JWT.AdminIssuer == "" {
		return fmt.Errorf("jwt.admin_issuer 不能为空")
	}
	if cfg.JWT.AdminExpireMinutes <= 0 {
		return fmt.Errorf("jwt.admin_expire_minutes 必须大于 0")
	}
	if cfg.JWT.UserIssuer == "" {
		return fmt.Errorf("jwt.user_issuer 不能为空")
	}
	if cfg.JWT.UserExpireMinutes <= 0 {
		return fmt.Errorf("jwt.user_expire_minutes 必须大于 0")
	}
	if cfg.Mail.Enabled {
		if cfg.Mail.Host == "" {
			return fmt.Errorf("mail.host 不能为空")
		}
		if cfg.Mail.Port <= 0 {
			return fmt.Errorf("mail.port 必须大于 0")
		}
		if cfg.Mail.FromAddress == "" {
			return fmt.Errorf("mail.from_address 不能为空")
		}
		if cfg.Mail.PasswordResetURLBase == "" {
			return fmt.Errorf("mail.password_reset_url_base 不能为空")
		}
	}
	if cfg.Storage.Driver == "" {
		return fmt.Errorf("storage.driver 不能为空")
	}
	if cfg.Storage.Driver != "local" {
		return fmt.Errorf("storage.driver 当前仅支持 local")
	}
	if cfg.Storage.LocalPath == "" {
		return fmt.Errorf("storage.local_path 不能为空")
	}
	if cfg.Storage.MaxSize <= 0 {
		return fmt.Errorf("storage.max_size 必须大于 0")
	}
	if len(cfg.Storage.AllowedTypes) == 0 {
		return fmt.Errorf("storage.allowed_types 不能为空")
	}
	return nil
}

func validateJWTSecret(name string, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fmt.Errorf("%s 不能为空", name)
	}
	if len(trimmed) < 32 {
		return fmt.Errorf("%s 长度不能小于 32 字符", name)
	}
	normalized := strings.ToLower(trimmed)
	if strings.Contains(normalized, "change_me") || normalized == "default" || normalized == "secret" || normalized == "password" {
		return fmt.Errorf("%s 不能使用默认或弱密钥", name)
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
 * ApplyTimezone 将应用配置的时区设置为 Go 进程默认时区。
 *
 * @return error 时区加载失败原因
 */
func (cfg *Config) ApplyTimezone() error {
	loc, err := time.LoadLocation(cfg.App.Timezone)
	if err != nil {
		return fmt.Errorf("加载应用时区失败：%s：%w", cfg.App.Timezone, err)
	}
	time.Local = loc
	return nil
}

/**
 * ShutdownTimeout 返回 API 服务优雅退出的超时时间。
 *
 * @return time.Duration 优雅退出超时时间
 */
func (cfg AppConfig) ShutdownTimeout() time.Duration {
	return time.Duration(cfg.ShutdownTimeoutSeconds) * time.Second
}

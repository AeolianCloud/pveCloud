package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Security SecurityConfig `mapstructure:"security"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	Charset      string `mapstructure:"charset"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name, d.Charset)
}

// RedisConfig Redis 连接配置。
// 用途：存储管理后台登录会话（session），支持退出登录立即失效 + Refresh Token 刷新。
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	// KeyPrefix key 前缀，用于多环境/多实例隔离（可选）
	KeyPrefix string `mapstructure:"key_prefix"`
}

// Addr 返回 host:port。
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
	// RefreshExpireHours Refresh Token 有效期（小时）。
	// 说明：Refresh Token 用于刷新 Access Token，建议比 expire_hours 长。
	RefreshExpireHours int `mapstructure:"refresh_expire_hours"`
}

// SecurityConfig 安全相关配置。
type SecurityConfig struct {
	Login LoginSecurityConfig `mapstructure:"login"`
}

// LoginSecurityConfig 登录限流/防爆破配置（基于 Redis）。
type LoginSecurityConfig struct {
	// Enabled 是否开启登录安全守卫（限流+防爆破）
	Enabled bool `mapstructure:"enabled"`

	// PerIPPerMinute 单 IP 每分钟最大登录请求数（包含成功/失败）
	PerIPPerMinute int `mapstructure:"per_ip_per_minute"`
	// PerUserPerMinute 单用户名每分钟最大登录请求数（包含成功/失败）
	PerUserPerMinute int `mapstructure:"per_user_per_minute"`

	// FailWindowMinutes 失败计数窗口（分钟），在该窗口内累计失败次数
	FailWindowMinutes int `mapstructure:"fail_window_minutes"`
	// FailThreshold 失败次数阈值（达到后锁定）
	FailThreshold int `mapstructure:"fail_threshold"`
	// LockMinutes 锁定时长（分钟）
	LockMinutes int `mapstructure:"lock_minutes"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// Load 加载配置。
// 第一步：读取 config.yaml 获取 server.mode。
// 第二步：根据 mode 合并对应环境配置：
//
//	debug   → config.dev.yaml
//	release → config.prod.yaml
//
// 环境配置文件中的字段会覆盖 config.yaml 中的同名字段。
// 环境配置文件不存在时不报错，直接使用 config.yaml。
func Load(cfgFile string) (*Config, error) {
	// ── 第一步：读取基础配置 config.yaml ──────────────────────
	base := viper.New()
	if cfgFile != "" {
		base.SetConfigFile(cfgFile)
	} else {
		base.SetConfigName("config")
		base.SetConfigType("yaml")
		base.AddConfigPath(".")
		base.AddConfigPath("./backend")
	}

	if err := base.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// ── 第二步：根据 mode 合并环境配置 ───────────────────────
	mode := base.GetString("server.mode")
	envFile := map[string]string{
		"debug":   "config.dev",
		"release": "config.prod",
	}[mode]

	if envFile != "" {
		env := viper.New()
		env.SetConfigName(envFile)
		env.SetConfigType("yaml")
		env.AddConfigPath(".")
		env.AddConfigPath("./backend")

		if err := env.ReadInConfig(); err == nil {
			// 用环境配置文件的值覆盖基础配置
			for _, key := range env.AllKeys() {
				base.Set(key, env.Get(key))
			}
		}
		// 环境配置文件不存在时静默跳过，继续使用 config.yaml
	}

	var cfg Config
	if err := base.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	return &cfg, nil
}

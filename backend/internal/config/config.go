package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
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

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
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

package bootstrap

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/**
 * ConnectDatabase 根据配置创建 MariaDB 连接并完成启动前连通性检查。
 *
 * @param ctx 启动上下文
 * @param cfg MariaDB 连接和连接池配置
 * @return *gorm.DB GORM 数据库连接
 * @return error 连接或 Ping 失败原因
 */
func ConnectDatabase(ctx context.Context, cfg DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open MariaDB %s:%d/%s: %w", cfg.Host, cfg.Port, cfg.Name, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB for MariaDB %s:%d/%s: %w", cfg.Host, cfg.Port, cfg.Name, err)
	}

	// 连接池参数会影响 API 和 Worker 的并发数据库访问能力，必须在首次 Ping 前设置完成。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping MariaDB %s:%d/%s: %w", cfg.Host, cfg.Port, cfg.Name, err)
	}

	return db, nil
}

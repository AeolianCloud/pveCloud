package bootstrap

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/AeolianCloud/pveCloud/server/internal/platform/cache"
)

/**
 * ConnectRedis 根据配置创建 Redis 客户端并完成启动前连通性检查。
 *
 * @param ctx 启动上下文
 * @param cfg Redis 连接配置
 * @return *cache.Redis 带项目 key 前缀的 Redis 访问器
 * @return error 连接或 Ping 失败原因
 */
func ConnectRedis(ctx context.Context, cfg RedisConfig) (*cache.Redis, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis %s (db=%d): %w", cfg.Addr, cfg.DB, err)
	}

	return cache.NewRedis(client, cfg.KeyPrefix), nil
}

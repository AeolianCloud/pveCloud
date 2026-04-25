package cache

import (
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

/**
 * Redis 封装项目统一 Redis 客户端和 key 前缀规则。
 */
type Redis struct {
	client    *goredis.Client
	keyPrefix string
}

/**
 * NewRedis 创建带统一 key 前缀的 Redis 访问器。
 *
 * @param client go-redis 客户端
 * @param keyPrefix 项目 Redis key 前缀
 * @return *Redis Redis 访问器
 */
func NewRedis(client *goredis.Client, keyPrefix string) *Redis {
	return &Redis{
		client:    client,
		keyPrefix: normalizePrefix(keyPrefix),
	}
}

/**
 * Client 返回底层 go-redis 客户端。
 *
 * @return *redis.Client go-redis 客户端
 */
func (r *Redis) Client() *goredis.Client {
	return r.client
}

/**
 * Key 按项目统一前缀拼接 Redis key。
 *
 * @param parts key 业务片段
 * @return string 带统一前缀的 Redis key
 */
func (r *Redis) Key(parts ...string) string {
	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, ": ")
		if part != "" {
			cleanParts = append(cleanParts, part)
		}
	}

	return r.keyPrefix + strings.Join(cleanParts, ":")
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "pvecloud"
	}
	prefix = strings.TrimRight(prefix, ":")
	return prefix + ":"
}

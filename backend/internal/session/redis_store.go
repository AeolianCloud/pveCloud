// internal/session/redis_store.go
// Redis 会话存储实现。
//
// Key 设计：
// - {prefix}sess:seq        -> INCR 自增序列，用于生成 session_id（数值型，方便写入 JWT claims）
// - {prefix}sess:{id}       -> HASH，保存会话字段
//
// HASH 字段：
// - user_id        管理员 ID
// - refresh_jti    refresh token 的 jti（用于 refresh rotation）
// - expires_at     过期时间（unix 秒）
// - revoked_at     撤销时间（unix 秒，0 表示未撤销）
// - ip / ua        最近一次登录/刷新来源
package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore Redis 会话存储。
type RedisStore struct {
	rdb    *redis.Client
	prefix string
}

// NewRedisStore 创建 Redis 会话存储。
func NewRedisStore(rdb *redis.Client, keyPrefix string) *RedisStore {
	return &RedisStore{rdb: rdb, prefix: keyPrefix}
}

func (s *RedisStore) seqKey() string { return s.prefix + "sess:seq" }
func (s *RedisStore) sessKey(id uint) string {
	return s.prefix + "sess:" + strconv.FormatUint(uint64(id), 10)
}

// Create 创建会话并写入 Redis，同时设置过期时间（EXPIREAT）。
func (s *RedisStore) Create(userID uint, expiresAt time.Time, meta LoginMeta) (uint, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 生成数值型 session_id
	id64, err := s.rdb.Incr(ctx, s.seqKey()).Result()
	if err != nil {
		return 0, "", err
	}
	sessionID := uint(id64)

	refreshJTI := newJTI()
	key := s.sessKey(sessionID)

	nowUnix := time.Now().Unix()
	fields := map[string]any{
		"user_id":     strconv.FormatUint(uint64(userID), 10),
		"refresh_jti": refreshJTI,
		"expires_at":  strconv.FormatInt(expiresAt.Unix(), 10),
		"revoked_at":  "0",
		"last_used_at": strconv.FormatInt(nowUnix, 10),
		"ip":          meta.IP,
		"ua":          meta.UserAgent,
	}

	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, key, fields)
	pipe.ExpireAt(ctx, key, expiresAt)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, "", err
	}

	return sessionID, refreshJTI, nil
}

// Get 获取会话数据。
func (s *RedisStore) Get(sessionID uint) (*Session, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := s.sessKey(sessionID)
	m, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(m) == 0 {
		return nil, ErrNotFound
	}

	userID, _ := strconv.ParseUint(m["user_id"], 10, 64)
	expiresUnix, _ := strconv.ParseInt(m["expires_at"], 10, 64)
	revokedUnix, _ := strconv.ParseInt(m["revoked_at"], 10, 64)

	var revokedAt *time.Time
	if revokedUnix > 0 {
		t := time.Unix(revokedUnix, 0)
		revokedAt = &t
	}

	return &Session{
		ID:          sessionID,
		AdminUserID: uint(userID),
		RefreshJTI:  m["refresh_jti"],
		ExpiresAt:   time.Unix(expiresUnix, 0),
		RevokedAt:   revokedAt,
	}, nil
}

// RotateRefreshJTI 原子旋转 refresh_jti，避免并发 refresh 时旧 token 被重复使用。
func (s *RedisStore) RotateRefreshJTI(sessionID uint, expected, newJTI string, meta LoginMeta) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := s.sessKey(sessionID)
	nowUnix := time.Now().Unix()

	// Lua：比较 refresh_jti，若一致则更新；否则返回 0
	// 返回值：
	// - 1：成功
	// - -1：key 不存在
	// - -2：已撤销
	// - 0：refresh_jti 不匹配
	script := redis.NewScript(`
local key = KEYS[1]
if redis.call("EXISTS", key) == 0 then
  return -1
end
local revoked = redis.call("HGET", key, "revoked_at")
if revoked and tonumber(revoked) and tonumber(revoked) > 0 then
  return -2
end
local cur = redis.call("HGET", key, "refresh_jti")
if cur ~= ARGV[1] then
  return 0
end
redis.call("HSET", key,
  "refresh_jti", ARGV[2],
  "last_used_at", ARGV[5],
  "ip", ARGV[3],
  "ua", ARGV[4]
)
return 1
`)

	res, err := script.Run(ctx, s.rdb, []string{key}, expected, newJTI, meta.IP, meta.UserAgent, strconv.FormatInt(nowUnix, 10)).Int()
	if err != nil {
		return err
	}

	switch res {
	case 1:
		return nil
	case -1:
		return ErrNotFound
	case -2:
		return ErrRevoked
	default:
		return ErrRefreshJTINotMatch
	}
}

// Revoke 撤销会话：直接删除 key，保证 Access/Refresh 立即失效。
func (s *RedisStore) Revoke(sessionID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	n, err := s.rdb.Del(ctx, s.sessKey(sessionID)).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// newJTI 生成一个随机 jti。
// 说明：放在 session 包内，避免 auth/service 依赖外部 uuid 库。
func newJTI() string {
	// 使用 crypto/rand 生成 16 字节随机数，再编码为 32 位 hex
	// 不引入第三方 uuid 依赖，保持项目依赖简单。
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// 极端情况下 fallback，避免直接 panic
		return "jti_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b)
}

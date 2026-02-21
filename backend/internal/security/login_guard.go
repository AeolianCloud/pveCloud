// internal/security/login_guard.go
// 登录安全守卫：限流 + 防爆破（失败次数锁定），基于 Redis 实现。
//
// 目标：
// 1) 限流：防止对 /auth/login 的高频请求压垮后端或被用作探测/扫库
// 2) 防爆破：连续密码错误达到阈值后，对该“用户名+IP”短期锁定
//
// 说明：
// - 本守卫只针对登录接口，不影响已登录后的业务接口
// - 采用 Redis TTL 自动过期，避免人工清理
// - 所有拦截都返回业务错误码 TooManyReqs（10003），HTTP 200（符合项目规范）
package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"pvecloud/backend/internal/config"
)

var (
	// ErrLoginRateLimited 登录请求被限流。
	ErrLoginRateLimited = errors.New("登录请求过于频繁，请稍后再试")
	// ErrLoginLocked 登录失败次数过多被锁定。
	ErrLoginLocked = errors.New("登录失败次数过多，请稍后再试")
)

// LoginGuard 登录安全守卫。
type LoginGuard struct {
	rdb    *redis.Client
	prefix string
	cfg    config.LoginSecurityConfig
}

// NewLoginGuard 创建登录安全守卫。
func NewLoginGuard(rdb *redis.Client, keyPrefix string, cfg config.LoginSecurityConfig) *LoginGuard {
	// 兼容缺省：给一组“够用且保守”的默认值，避免配置遗漏导致守卫形同虚设
	if cfg.PerIPPerMinute <= 0 {
		cfg.PerIPPerMinute = 30
	}
	if cfg.PerUserPerMinute <= 0 {
		cfg.PerUserPerMinute = 10
	}
	if cfg.FailWindowMinutes <= 0 {
		cfg.FailWindowMinutes = 10
	}
	if cfg.FailThreshold <= 0 {
		cfg.FailThreshold = 5
	}
	if cfg.LockMinutes <= 0 {
		cfg.LockMinutes = 15
	}

	return &LoginGuard{
		rdb:    rdb,
		prefix: keyPrefix,
		cfg:    cfg,
	}
}

// PreCheck 登录前置检查：限流 + 锁定校验。
// username 用于按用户限流和锁定，ip 用于按 IP 限流和锁定维度。
func (g *LoginGuard) PreCheck(username, ip string) error {
	if g == nil || !g.cfg.Enabled {
		return nil
	}

	username = strings.TrimSpace(username)
	ip = strings.TrimSpace(ip)
	if username == "" || ip == "" {
		// 兜底：缺少关键维度时不做拦截（避免误伤），但仍允许后续业务返回参数错误
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1) 锁定校验（用户名+IP）
	lockKey := g.lockKey(username, ip)
	locked, err := g.rdb.Exists(ctx, lockKey).Result()
	if err != nil {
		return err
	}
	if locked > 0 {
		return ErrLoginLocked
	}

	// 2) 固定窗口限流（按分钟）
	window := time.Now().Format("200601021504") // YYYYMMDDHHmm

	// 单 IP 限流
	if g.cfg.PerIPPerMinute > 0 {
		n, err := g.incrWithTTL(ctx, g.ipRateKey(ip, window), 70*time.Second)
		if err != nil {
			return err
		}
		if int(n) > g.cfg.PerIPPerMinute {
			return ErrLoginRateLimited
		}
	}

	// 单用户名限流
	if g.cfg.PerUserPerMinute > 0 {
		n, err := g.incrWithTTL(ctx, g.userRateKey(username, window), 70*time.Second)
		if err != nil {
			return err
		}
		if int(n) > g.cfg.PerUserPerMinute {
			return ErrLoginRateLimited
		}
	}

	return nil
}

// RecordFailure 记录一次登录失败；当失败次数达到阈值时写入锁定 key（带 TTL）。
func (g *LoginGuard) RecordFailure(username, ip string) error {
	if g == nil || !g.cfg.Enabled {
		return nil
	}

	username = strings.TrimSpace(username)
	ip = strings.TrimSpace(ip)
	if username == "" || ip == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	failKey := g.failKey(username, ip)
	lockKey := g.lockKey(username, ip)

	failTTL := time.Duration(g.cfg.FailWindowMinutes) * time.Minute
	lockTTL := time.Duration(g.cfg.LockMinutes) * time.Minute

	// Lua：INCR + 首次设置 EXPIRE；达到阈值后 SETEX lockKey
	script := redis.NewScript(`
local failKey = KEYS[1]
local lockKey = KEYS[2]
local failTTL = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local lockTTL = tonumber(ARGV[3])

-- 如果已锁定就直接返回 -1
if redis.call("EXISTS", lockKey) == 1 then
  return -1
end

local n = redis.call("INCR", failKey)
if n == 1 then
  redis.call("EXPIRE", failKey, failTTL)
end

if n >= threshold then
  redis.call("SET", lockKey, "1", "EX", lockTTL)
  return -2
end
return n
`)

	res, err := script.Run(ctx, g.rdb, []string{failKey, lockKey}, int(failTTL.Seconds()), g.cfg.FailThreshold, int(lockTTL.Seconds())).Int()
	if err != nil {
		return err
	}
	_ = res // 目前不需要返回值，留作后续扩展（比如把剩余次数返回给前端）
	return nil
}

// RecordSuccess 登录成功：清理失败计数与锁定（避免用户输错一次后长期受影响）。
func (g *LoginGuard) RecordSuccess(username, ip string) error {
	if g == nil || !g.cfg.Enabled {
		return nil
	}

	username = strings.TrimSpace(username)
	ip = strings.TrimSpace(ip)
	if username == "" || ip == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := g.rdb.Del(ctx, g.failKey(username, ip), g.lockKey(username, ip)).Result()
	return err
}

// ---- key 规则 -------------------------------------------------

func (g *LoginGuard) ipRateKey(ip, window string) string {
	return fmt.Sprintf("%ssec:login:rl:ip:%s:%s", g.prefix, ip, window)
}

func (g *LoginGuard) userRateKey(username, window string) string {
	return fmt.Sprintf("%ssec:login:rl:user:%s:%s", g.prefix, username, window)
}

func (g *LoginGuard) failKey(username, ip string) string {
	return fmt.Sprintf("%ssec:login:fail:%s:%s", g.prefix, username, ip)
}

func (g *LoginGuard) lockKey(username, ip string) string {
	return fmt.Sprintf("%ssec:login:lock:%s:%s", g.prefix, username, ip)
}

// incrWithTTL 计数器自增并确保设置 TTL（首次写入时才设置 TTL）。
func (g *LoginGuard) incrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	// 用 Lua 保证 INCR 与 EXPIRE 的原子性，避免并发下漏设置 TTL
	script := redis.NewScript(`
local key = KEYS[1]
local ttl = tonumber(ARGV[1])
local n = redis.call("INCR", key)
if n == 1 then
  redis.call("EXPIRE", key, ttl)
end
return n
`)
	return script.Run(ctx, g.rdb, []string{key}, int(ttl.Seconds())).Int64()
}


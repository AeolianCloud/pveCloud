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
	"strconv"
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

	// 使用 WATCH + TxPipelined 实现无 Lua 的原子更新：
	// 1) 检查是否已锁定
	// 2) 失败计数 +1，首次失败设置窗口 TTL
	// 3) 达到阈值时写入锁定 key（TTL 自动解锁）
	//
	// 并发冲突时会触发 redis.TxFailedErr，按有限重试处理。
	for attempt := 0; attempt < 5; attempt++ {
		err := g.rdb.Watch(ctx, func(tx *redis.Tx) error {
			locked, err := tx.Exists(ctx, lockKey).Result()
			if err != nil {
				return err
			}
			if locked > 0 {
				// 已锁定无需重复操作
				return nil
			}

			current := int64(0)
			val, err := tx.Get(ctx, failKey).Result()
			switch {
			case err == redis.Nil:
				current = 0
			case err != nil:
				return err
			default:
				n, convErr := strconv.ParseInt(val, 10, 64)
				if convErr != nil {
					return convErr
				}
				current = n
			}

			next := current + 1
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Incr(ctx, failKey)
				if current == 0 {
					pipe.Expire(ctx, failKey, failTTL)
				}
				if next >= int64(g.cfg.FailThreshold) {
					pipe.Set(ctx, lockKey, "1", lockTTL)
				}
				return nil
			})
			return err
		}, failKey, lockKey)

		if err == nil {
			return nil
		}
		if err == redis.TxFailedErr {
			continue
		}
		return err
	}

	return redis.TxFailedErr
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
	// 使用事务保证 INCR 与 EXPIRE 一起提交（无 Lua）。
	// 这里每次都刷新 TTL，结合按分钟分桶 key，不影响限流语义。
	var incrCmd *redis.IntCmd
	_, err := g.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		incrCmd = pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, ttl)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return incrCmd.Val(), nil
}

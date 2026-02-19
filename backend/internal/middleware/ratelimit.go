package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/pkg/response"
)

// failRecord 记录某个 IP 的失败次数和锁定截止时间。
type failRecord struct {
	Count       int
	LockedUntil time.Time
}

// LoginRateLimiter 提供最小化登录失败防护，避免暴力猜解。
type LoginRateLimiter struct {
	mu      sync.Mutex
	records map[string]failRecord
	maxFail int
	lockDur time.Duration
}

// NewLoginRateLimiter 创建基于 IP 的登录失败限流器。
func NewLoginRateLimiter(maxFail int, lockDur time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{records: make(map[string]failRecord), maxFail: maxFail, lockDur: lockDur}
}

// Middleware 在登录接口前置校验 IP 是否已被临时锁定。
func (l *LoginRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c.ClientIP())
		if l.IsLocked(ip) {
			response.Error(c, http.StatusTooManyRequests, 42901, "too many failed attempts, please retry later")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RecordFailure 记录一次登录失败。
func (l *LoginRateLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	rec := l.records[ip]
	rec.Count++
	if rec.Count >= l.maxFail {
		rec.LockedUntil = time.Now().Add(l.lockDur)
		rec.Count = 0
	}
	l.records[ip] = rec
}

// RecordSuccess 登录成功后清空失败计数。
func (l *LoginRateLimiter) RecordSuccess(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.records, ip)
}

// IsLocked 判断 IP 是否在锁定期内。
func (l *LoginRateLimiter) IsLocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	rec, ok := l.records[ip]
	if !ok {
		return false
	}
	if rec.LockedUntil.After(time.Now()) {
		return true
	}
	if !rec.LockedUntil.IsZero() {
		delete(l.records, ip)
	}
	return false
}

func clientIP(raw string) string {
	ip := net.ParseIP(raw)
	if ip == nil {
		return raw
	}
	return ip.String()
}

package middleware

import (
	"errors"
	"time"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"pvecloud/backend/internal/session"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	// SessionID 登录会话 ID（用于服务端撤销会话与 Refresh Token）
	SessionID uint `json:"session_id"`
	jwt.RegisteredClaims
}

// JWTAuth JWT 鉴权中间件。
//
// 除了校验 JWT 的签名与过期时间外，还会校验 session 是否仍有效：
// - session 不存在 / 已撤销 / 已过期 → 401
//
// 这样可以实现“退出登录立即失效”，解决纯 JWT 无法服务端撤销的问题。
func JWTAuth(secret string, sessStore session.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, errcode.Unauthorized.Msg())
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, errcode.TokenInvalid.Msg())
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			// 尽量区分“过期”和“无效”，便于前端给更准确的提示
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.Unauthorized(c, errcode.TokenExpired.Msg())
			} else {
				response.Unauthorized(c, errcode.TokenInvalid.Msg())
			}
			c.Abort()
			return
		}

		// token 中必须携带 session_id（sid），用于服务端会话校验
		if claims.SessionID == 0 {
			response.Unauthorized(c, errcode.TokenInvalid.Msg())
			c.Abort()
			return
		}

		// 校验会话是否有效：未撤销、未过期、且属于当前用户
		sess, err := sessStore.Get(claims.SessionID)
		if err != nil {
			// 会话不存在（过期/撤销/从未创建）统一视为登录已过期
			response.Unauthorized(c, errcode.TokenExpired.Msg())
			c.Abort()
			return
		}
		if sess.AdminUserID != claims.UserID {
			response.Unauthorized(c, errcode.TokenInvalid.Msg())
			c.Abort()
			return
		}
		if sess.RevokedAt != nil || time.Now().After(sess.ExpiresAt) {
			response.Unauthorized(c, errcode.TokenExpired.Msg())
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("session_id", claims.SessionID)
		c.Next()
	}
}

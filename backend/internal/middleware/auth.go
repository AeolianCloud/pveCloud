package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pvecloud/backend/internal/security"
	"pvecloud/backend/pkg/response"
)

const (
	ctxUserIDKey  = "userID"
	ctxRoleKey    = "role"
	ctxTokenIDKey = "tokenID"
)

// AuthMiddleware 校验 Bearer Token 并将用户身份写入上下文，同时校验黑名单和强制下线状态。
func AuthMiddleware(jwtManager *JWTManager, tokenStore *security.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, 40101, "missing or invalid authorization header")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtManager.ParseToken(token)
		if err != nil || claims.Type != "access" {
			response.Error(c, http.StatusUnauthorized, 40102, "invalid or expired token")
			c.Abort()
			return
		}

		if tokenStore != nil && tokenStore.Enabled() {
			blocked, blockErr := tokenStore.IsAccessTokenBlacklisted(c.Request.Context(), claims.ID)
			if blockErr == nil && blocked {
				response.Error(c, http.StatusUnauthorized, 40103, "token has been revoked")
				c.Abort()
				return
			}

			forcedAt, forceErr := tokenStore.GetForceLogoutTime(c.Request.Context(), claims.UserID)
			if forceErr == nil && !forcedAt.IsZero() && claims.IssuedAt != nil && claims.IssuedAt.Time.Before(forcedAt) {
				response.Error(c, http.StatusUnauthorized, 40104, "user has been forced offline")
				c.Abort()
				return
			}
		}

		c.Set(ctxUserIDKey, claims.UserID)
		c.Set(ctxRoleKey, claims.Role)
		c.Set(ctxTokenIDKey, claims.ID)
		c.Next()
	}
}

// AdminOnlyMiddleware 限制仅管理员角色可访问。
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ctxRoleKey)
		if role != "admin" {
			response.Error(c, http.StatusForbidden, 40301, "admin role required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// UserIDFromContext 从 gin context 中安全读取 userID。
func UserIDFromContext(c *gin.Context) uint {
	if v, ok := c.Get(ctxUserIDKey); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// RoleFromContext 从 gin context 中安全读取 role。
func RoleFromContext(c *gin.Context) string {
	if v, ok := c.Get(ctxRoleKey); ok {
		if role, ok := v.(string); ok {
			return role
		}
	}
	return ""
}

// TokenIDFromContext 从 gin context 中安全读取 tokenID。
func TokenIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(ctxTokenIDKey); ok {
		if tokenID, ok := v.(string); ok {
			return tokenID
		}
	}
	return ""
}

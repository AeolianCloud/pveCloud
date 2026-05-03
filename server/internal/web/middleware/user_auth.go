package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/models"
	"github.com/AeolianCloud/pveCloud/server/internal/platform/bootstrap"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/shared/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/textutil"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	websupport "github.com/AeolianCloud/pveCloud/server/internal/web/support"
)

const (
	userContextKey        = "web_user"
	userIDContextKey      = "web_user_id"
	userSessionContextKey = "web_user_session"
)

/**
 * UserAuth 校验用户端 JWT 和会话状态。
 */
func UserAuth(cfg bootstrap.JWTConfig, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortUserUnauthorized(c)
			return
		}

		claims, err := jwtpkg.Parse(tokenString, cfg.UserSecret)
		if err != nil || claims.TokenType != "user" || claims.UserID == 0 || claims.Issuer != cfg.UserIssuer || strings.TrimSpace(claims.ID) == "" {
			abortUserUnauthorized(c)
			return
		}

		now := time.Now()
		var session models.UserSession
		err = db.WithContext(c.Request.Context()).
			Where("session_id = ? AND user_id = ?", claims.ID, claims.UserID).
			First(&session).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortUserUnauthorized(c)
			return
		}
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if session.Status != "active" || !session.ExpiresAt.After(now) {
			if session.Status == "active" && !session.ExpiresAt.After(now) {
				reason := "expired"
				_ = db.WithContext(c.Request.Context()).Model(&models.UserSession{}).
					Where("id = ?", session.ID).
					Updates(map[string]interface{}{
						"status":        "expired",
						"revoked_at":    now,
						"revoke_reason": reason,
					}).Error
			}
			abortUserUnauthorized(c)
			return
		}

		var user models.User
		err = db.WithContext(c.Request.Context()).Where("id = ?", claims.UserID).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortUserUnauthorized(c)
			return
		}
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if user.Status != "active" {
			response.Error(c, apperrors.ErrForbidden.WithMessage("用户账号已被禁用"))
			c.Abort()
			return
		}

		_ = db.WithContext(c.Request.Context()).Model(&models.UserSession{}).
			Where("id = ?", session.ID).
			Updates(map[string]interface{}{
				"last_seen_at": now,
				"last_seen_ip": c.ClientIP(),
				"user_agent":   textutil.TrimTo(c.Request.UserAgent(), 500),
			}).Error

		c.Set(userContextKey, user)
		c.Set(userIDContextKey, user.ID)
		c.Set(userSessionContextKey, websupport.SessionSummary(session))
		c.Next()
	}
}

/**
 * CurrentUser 从 Gin 上下文读取当前用户。
 */
func CurrentUser(c *gin.Context) (models.User, bool) {
	value, ok := c.Get(userContextKey)
	if !ok {
		return models.User{}, false
	}
	user, ok := value.(models.User)
	return user, ok && user.ID > 0
}

/**
 * CurrentUserID 从 Gin 上下文读取当前用户 ID。
 */
func CurrentUserID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}
	userID, ok := value.(uint64)
	return userID, ok && userID > 0
}

/**
 * CurrentUserSession 从 Gin 上下文读取当前用户会话摘要。
 */
func CurrentUserSession(c *gin.Context) (webdto.SessionSummary, bool) {
	value, ok := c.Get(userSessionContextKey)
	if !ok {
		return webdto.SessionSummary{}, false
	}
	session, ok := value.(webdto.SessionSummary)
	return session, ok && session.SessionID != ""
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func abortUserUnauthorized(c *gin.Context) {
	response.Error(c, apperrors.ErrUnauthorized)
	c.Abort()
}

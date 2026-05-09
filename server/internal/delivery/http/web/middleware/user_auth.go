package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	mysqluser "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/user"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	webauthusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/auth"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
)

const (
	userContextKey        = "web_user"
	userIDContextKey      = "web_user_id"
	userSessionContextKey = "web_user_session"
)

type UserAuthenticator interface {
	Authenticate(ctx context.Context, tokenString string, clientIP string, userAgent string) (webauthusecase.AuthenticatedUser, error)
}

/**
 * UserAuth 校验用户端 JWT 和会话状态。
 */
func UserAuth(authenticator UserAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortUserUnauthorized(c)
			return
		}

		auth, err := authenticator.Authenticate(c.Request.Context(), tokenString, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		c.Set(userContextKey, auth.User)
		c.Set(userIDContextKey, auth.User.ID)
		c.Set(userSessionContextKey, auth.Session)
		c.Next()
	}
}

/**
 * CurrentUser 从 Gin 上下文读取当前用户。
 */
func CurrentUser(c *gin.Context) (mysqluser.User, bool) {
	value, ok := c.Get(userContextKey)
	if !ok {
		return mysqluser.User{}, false
	}
	user, ok := value.(mysqluser.User)
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

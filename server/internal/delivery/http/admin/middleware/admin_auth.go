package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	mysqliam "github.com/AeolianCloud/pveCloud/server/internal/repository/mysql/iam"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/requestcontext"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	adminauthusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/auth"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

const (
	adminClaimsContextKey          = "admin_claims"
	adminUserContextKey            = "admin_user"
	adminSessionContextKey         = "admin_session"
	adminIDContextKey              = "admin_id"
	adminRoleIDsContextKey         = "admin_role_ids"
	adminPermissionCodesContextKey = "admin_permission_codes"
)

type AdminAuthenticator interface {
	Authenticate(ctx context.Context, tokenString string, clientIP string, userAgent string) (adminauthusecase.AuthenticatedAdmin, error)
}

/**
 * AdminAuth 校验管理端 JWT、会话和当前 RBAC，并把管理员身份写入请求上下文。
 */
func AdminAuth(authenticator AdminAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortAdminUnauthorized(c)
			return
		}

		auth, err := authenticator.Authenticate(c.Request.Context(), tokenString, c.ClientIP(), c.Request.UserAgent())
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		auditCtx := requestcontext.WithRequestContext(c.Request.Context(), auth.RequestContext)
		c.Request = c.Request.WithContext(auditCtx)

		c.Set(adminClaimsContextKey, auth.Claims)
		c.Set(adminUserContextKey, auth.Admin)
		c.Set(adminSessionContextKey, auth.Session)
		c.Set(adminIDContextKey, auth.Admin.ID)
		c.Set(adminRoleIDsContextKey, auth.RoleIDs)
		c.Set(adminPermissionCodesContextKey, auth.PermissionCodes)
		c.Next()
	}
}

/**
 * CurrentAdmin 从 Gin 上下文读取当前管理员。
 *
 * @param c Gin 请求上下文
 * @return mysqliam.AdminUser 管理员
 * @return bool 是否存在管理员身份
 */
func CurrentAdmin(c *gin.Context) (mysqliam.AdminUser, bool) {
	value, ok := c.Get(adminUserContextKey)
	if !ok {
		return mysqliam.AdminUser{}, false
	}
	admin, ok := value.(mysqliam.AdminUser)
	return admin, ok && admin.ID > 0
}

/**
 * CurrentAdminSession 从 Gin 上下文读取当前会话摘要。
 *
 * @param c Gin 请求上下文
 * @return admin.SessionSummary 会话摘要
 * @return bool 是否存在会话
 */
func CurrentAdminSession(c *gin.Context) (admindto.SessionSummary, bool) {
	value, ok := c.Get(adminSessionContextKey)
	if !ok {
		return admindto.SessionSummary{}, false
	}
	session, ok := value.(admindto.SessionSummary)
	return session, ok && session.SessionID != ""
}

/**
 * CurrentAdminID 从 Gin 上下文读取当前管理员 ID。
 *
 * @param c Gin 请求上下文
 * @return uint64 管理员 ID
 * @return bool 是否存在管理员身份
 */
func CurrentAdminID(c *gin.Context) (uint64, bool) {
	value, ok := c.Get(adminIDContextKey)
	if !ok {
		return 0, false
	}
	adminID, ok := value.(uint64)
	return adminID, ok && adminID > 0
}

/**
 * CurrentAdminRoleIDs 从 Gin 上下文读取当前管理员角色 ID。
 *
 * @param c Gin 请求上下文
 * @return []uint64 角色 ID 列表
 */
func CurrentAdminRoleIDs(c *gin.Context) []uint64 {
	value, ok := c.Get(adminRoleIDsContextKey)
	if !ok {
		return nil
	}
	roleIDs, _ := value.([]uint64)
	return roleIDs
}

/**
 * CurrentAdminPermissionCodes 从 Gin 上下文读取当前管理员权限码。
 *
 * @param c Gin 请求上下文
 * @return []string 权限码列表
 */
func CurrentAdminPermissionCodes(c *gin.Context) []string {
	value, ok := c.Get(adminPermissionCodesContextKey)
	if !ok {
		return nil
	}
	permissionCodes, _ := value.([]string)
	return permissionCodes
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}
	return parts[1], true
}

func abortAdminUnauthorized(c *gin.Context) {
	response.Error(c, apperrors.ErrUnauthorized)
	c.Abort()
}

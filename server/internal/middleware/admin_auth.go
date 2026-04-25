package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/pkg/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

const (
	adminClaimsContextKey          = "admin_claims"
	adminIDContextKey              = "admin_id"
	adminRoleIDsContextKey         = "admin_role_ids"
	adminPermissionCodesContextKey = "admin_permission_codes"
)

/**
 * AdminAuth 校验管理端 JWT 并把管理员身份写入请求上下文。
 *
 * @param cfg JWT 配置
 * @return gin.HandlerFunc Gin 中间件
 */
func AdminAuth(cfg bootstrap.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortAdminUnauthorized(c)
			return
		}

		claims, err := jwtpkg.Parse(tokenString, cfg.AdminSecret)
		if err != nil || claims.TokenType != "admin" || claims.AdminID == 0 || claims.Issuer != cfg.AdminIssuer {
			abortAdminUnauthorized(c)
			return
		}

		c.Set(adminClaimsContextKey, claims)
		c.Set(adminIDContextKey, claims.AdminID)
		c.Set(adminRoleIDsContextKey, claims.RoleIDs)
		c.Set(adminPermissionCodesContextKey, claims.PermissionCodes)
		c.Next()
	}
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

package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/AeolianCloud/pveCloud/server/internal/bootstrap"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/models"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	jwtpkg "github.com/AeolianCloud/pveCloud/server/internal/pkg/jwt"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

const (
	adminClaimsContextKey          = "admin_claims"
	adminUserContextKey            = "admin_user"
	adminSessionContextKey         = "admin_session"
	adminIDContextKey              = "admin_id"
	adminRoleIDsContextKey         = "admin_role_ids"
	adminPermissionCodesContextKey = "admin_permission_codes"
)

/**
 * AdminAuth 校验管理端 JWT、会话和当前 RBAC，并把管理员身份写入请求上下文。
 *
 * @param cfg JWT 配置
 * @param db 数据库连接
 * @return gin.HandlerFunc Gin 中间件
 */
func AdminAuth(cfg bootstrap.JWTConfig, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			abortAdminUnauthorized(c)
			return
		}

		claims, err := jwtpkg.Parse(tokenString, cfg.AdminSecret)
		if err != nil || claims.TokenType != "admin" || claims.AdminID == 0 || claims.Issuer != cfg.AdminIssuer || strings.TrimSpace(claims.ID) == "" {
			abortAdminUnauthorized(c)
			return
		}

		now := time.Now()
		var session models.AdminSession
		err = db.WithContext(c.Request.Context()).
			Where("session_id = ? AND admin_id = ?", claims.ID, claims.AdminID).
			First(&session).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortAdminUnauthorized(c)
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
				_ = db.WithContext(c.Request.Context()).Model(&models.AdminSession{}).
					Where("id = ?", session.ID).
					Updates(map[string]interface{}{
						"status":        "expired",
						"revoked_at":    now,
						"revoke_reason": reason,
					}).Error
			}
			abortAdminUnauthorized(c)
			return
		}

		var admin models.AdminUser
		err = db.WithContext(c.Request.Context()).
			Where("deleted_at IS NULL").
			Where("id = ?", claims.AdminID).
			First(&admin).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			abortAdminUnauthorized(c)
			return
		}
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		if admin.Status != "active" {
			response.Error(c, apperrors.ErrForbidden.WithMessage("管理员账号已被禁用"))
			c.Abort()
			return
		}

		roleIDs, err := currentRoleIDs(c, db, admin.ID)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}
		permissionCodes, err := currentPermissionCodes(c, db, admin.ID)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		_ = db.WithContext(c.Request.Context()).Model(&models.AdminSession{}).
			Where("id = ?", session.ID).
			Updates(map[string]interface{}{
				"last_seen_at": now,
				"last_seen_ip": c.ClientIP(),
				"user_agent":   trimTo(c.Request.UserAgent(), 500),
			}).Error

		c.Set(adminClaimsContextKey, claims)
		c.Set(adminUserContextKey, admin)
		c.Set(adminSessionContextKey, admindto.SessionSummary{
			SessionID: session.SessionID,
			IssuedAt:  session.IssuedAt,
			ExpiresAt: session.ExpiresAt,
		})
		c.Set(adminIDContextKey, admin.ID)
		c.Set(adminRoleIDsContextKey, roleIDs)
		c.Set(adminPermissionCodesContextKey, permissionCodes)
		c.Next()
	}
}

/**
 * CurrentAdmin 从 Gin 上下文读取当前管理员。
 *
 * @param c Gin 请求上下文
 * @return models.AdminUser 管理员
 * @return bool 是否存在管理员身份
 */
func CurrentAdmin(c *gin.Context) (models.AdminUser, bool) {
	value, ok := c.Get(adminUserContextKey)
	if !ok {
		return models.AdminUser{}, false
	}
	admin, ok := value.(models.AdminUser)
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

func currentRoleIDs(c *gin.Context, db *gorm.DB, adminID uint64) ([]uint64, error) {
	var roleIDs []uint64
	err := db.WithContext(c.Request.Context()).
		Table("admin_user_roles").
		Select("admin_roles.id").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", "active").
		Order("admin_roles.id ASC").
		Scan(&roleIDs).Error
	return roleIDs, err
}

func currentPermissionCodes(c *gin.Context, db *gorm.DB, adminID uint64) ([]string, error) {
	var codes []string
	err := db.WithContext(c.Request.Context()).
		Table("admin_user_roles").
		Distinct("admin_permissions.code").
		Joins("JOIN admin_roles ON admin_roles.id = admin_user_roles.role_id").
		Joins("JOIN admin_role_permissions ON admin_role_permissions.role_id = admin_roles.id").
		Joins("JOIN admin_permissions ON admin_permissions.id = admin_role_permissions.permission_id").
		Where("admin_user_roles.admin_id = ?", adminID).
		Where("admin_roles.status = ?", "active").
		Order("admin_permissions.code ASC").
		Scan(&codes).Error
	return codes, err
}

func trimTo(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

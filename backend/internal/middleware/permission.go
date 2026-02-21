package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// RequirePermission 校验当前用户是否拥有指定权限。
// 判断逻辑：
//  1. 若用户拥有 super_admin 角色（查库，不依赖 JWT role 字段），直接放行。
//  2. 否则通过 用户→角色→权限 三表联查，验证是否持有 permName 权限。
func RequirePermission(db *gorm.DB, permName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		// 先判断是否为 super_admin（查库，避免 JWT role 字段只取了第一个角色的问题）
		var superCount int64
		db.Table("admin_user_roles ur").
			Joins("JOIN admin_roles r ON r.id = ur.admin_role_id").
			Where("ur.admin_user_id = ? AND r.name = 'super_admin' AND r.deleted_at IS NULL", userID).
			Count(&superCount)
		if superCount > 0 {
			c.Next()
			return
		}

		// 查询：用户所属角色是否拥有 permName 权限
		// 关联路径：admin_user_roles → admin_role_permissions → admin_permissions
		var count int64
		db.Table("admin_permissions p").
			Joins("JOIN admin_role_permissions rp ON rp.admin_permission_id = p.id").
			Joins("JOIN admin_user_roles ur ON ur.admin_role_id = rp.admin_role_id").
			Where("ur.admin_user_id = ? AND p.name = ?", userID, permName).
			Count(&count)

		if count == 0 {
			response.Forbidden(c, errcode.Forbidden.Msg())
			c.Abort()
			return
		}

		c.Next()
	}
}

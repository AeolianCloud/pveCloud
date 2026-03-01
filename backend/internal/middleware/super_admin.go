// internal/middleware/super_admin.go
// 超级管理员校验中间件。
//
// 需求背景（为什么要单独做一个中间件）：
// - “菜单管理”属于平台级能力，不允许通过普通权限配置放开，策略固定为：仅 super_admin 可访问。
// - 如果仅用 RequirePermission，只要把权限分配给其他角色就会突破策略，因此这里单独做硬策略校验。
package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pvecloud/backend/pkg/response"
	"pvecloud/backend/pkg/response/errcode"
)

// RequireSuperAdmin 校验当前用户是否拥有 super_admin 角色。
//
// 说明：
// - 这里查库而不是用 JWT role 字段，避免 JWT role 只取第一个角色导致误判。
// - 只要拥有 super_admin 角色即放行（可兼容未来多角色）。
func RequireSuperAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var superCount int64
		db.Table("admin_user_roles ur").
			Joins("JOIN admin_roles r ON r.id = ur.admin_role_id").
			Where("ur.admin_user_id = ? AND r.name = 'super_admin' AND r.deleted_at IS NULL", userID).
			Count(&superCount)

		if superCount == 0 {
			response.Forbidden(c, errcode.Forbidden.Msg())
			c.Abort()
			return
		}

		c.Next()
	}
}


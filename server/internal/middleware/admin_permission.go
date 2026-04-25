package middleware

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

/**
 * AdminPermission 校验当前管理员是否拥有全部指定权限码。
 *
 * @param requiredCodes 必须具备的权限码
 * @return gin.HandlerFunc Gin 中间件
 */
func AdminPermission(requiredCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(requiredCodes) == 0 {
			c.Next()
			return
		}

		permissionSet := make(map[string]struct{}, len(CurrentAdminPermissionCodes(c)))
		for _, code := range CurrentAdminPermissionCodes(c) {
			permissionSet[code] = struct{}{}
		}

		for _, code := range requiredCodes {
			if _, ok := permissionSet[code]; !ok {
				response.Error(c, apperrors.ErrForbidden)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

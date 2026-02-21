// internal/middleware/oplog.go
// 操作日志中间件，挂在需要审计的写操作路由上，自动记录执行人、模块、动作和结果。
package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"pvecloud/backend/internal/model"
)

// WriteOpLog 返回一个写操作日志的中间件。
// module：操作模块（如 "admin"、"role"）
// action：操作动作（如 "create"、"update"、"delete"）
// getLabel：从 gin.Context 提取目标描述的函数，nil 时不记录 target_label
//
// 用法示例（挂在路由上）：
//
//	adminUsers.POST("", middleware.WriteOpLog(db, "admin", "create", nil), adminHandler.Create)
//	adminUsers.DELETE("/:id", middleware.WriteOpLog(db, "admin", "delete", nil), adminHandler.Delete)
func WriteOpLog(db *gorm.DB, module, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先执行后续 handler
		c.Next()

		// handler 执行完后再写日志（此时才能知道是否出错）
		// 只记录最终响应成功的操作（HTTP 2xx 且没有中途 Abort）
		if c.Writer.Status() >= 400 {
			return
		}

		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		uid, _ := userID.(uint)
		uname, _ := username.(string)

		// target_id：从 URL 参数 :id 取，取不到则为 0
		var targetID uint
		if idStr := c.Param("id"); idStr != "" && idStr != "/" {
			var n uint64
			if _, err := parseUint(idStr, &n); err == nil {
				targetID = uint(n)
			}
		}

		log := model.AdminOpLog{
			AdminUserID: uid,
			Username:    uname,
			Module:      module,
			Action:      action,
			TargetID:    targetID,
			Status:      1,
			IP:          c.ClientIP(),
		}
		// 静默写入，失败不影响业务
		db.Create(&log)
	}
}

// parseUint 简单解析无符号整数，避免引入 strconv 包外的依赖。
func parseUint(s string, out *uint64) (int, error) {
	var v uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, &parseError{s}
		}
		v = v*10 + uint64(c-'0')
	}
	*out = v
	return len(s), nil
}

type parseError struct{ s string }

func (e *parseError) Error() string { return "invalid number: " + e.s }

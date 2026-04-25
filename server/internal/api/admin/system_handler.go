package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

/**
 * SystemHandler 处理管理端系统检查接口。
 */
type SystemHandler struct{}

/**
 * NewSystemHandler 创建管理端系统接口处理器。
 *
 * @return *SystemHandler 管理端系统接口处理器
 */
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

/**
 * Ping 返回管理端 API 入口的基础连通性结果。
 *
 * @route GET /admin-api/ping
 * @response 200 {"code":0,"message":"成功","data":{"scope":"admin-api","pong":true}}
 */
func (h *SystemHandler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"scope": "admin-api",
		"pong":  true,
	})
}

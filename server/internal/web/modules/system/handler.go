package system

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
)

/**
 * SystemHandler 处理用户端系统检查接口。
 */
type SystemHandler struct{}

/**
 * NewSystemHandler 创建用户端系统接口处理器。
 *
 * @return *SystemHandler 用户端系统接口处理器
 */
func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

/**
 * Ping 返回用户端 API 入口的基础连通性结果。
 *
 * @route GET /api/ping
 * @response 200 {"code":0,"message":"成功","data":{"scope":"api","pong":true}}
 */
func (h *SystemHandler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"scope": "api",
		"pong":  true,
	})
}

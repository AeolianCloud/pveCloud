package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// parseUintParam 从路径参数读取 uint，失败时返回 0。
func parseUintParam(c *gin.Context, key string) uint {
	id64, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0
	}
	return uint(id64)
}

package web

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

func (h *SystemHandler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"scope": "api",
		"pong":  true,
	})
}

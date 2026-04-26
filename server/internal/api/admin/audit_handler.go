package admin

import (
	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/dto/admin"
	"github.com/AeolianCloud/pveCloud/server/internal/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/pkg/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/response"
	"github.com/AeolianCloud/pveCloud/server/internal/pkg/validator"
	"github.com/AeolianCloud/pveCloud/server/internal/services"
)

/**
 * AuditHandler 处理管理端审计和高危操作日志接口。
 */
type AuditHandler struct {
	auditService *services.AdminAuditService
}

/**
 * NewAuditHandler 创建管理端审计接口处理器。
 *
 * @param auditService 管理端审计服务
 * @return *AuditHandler 管理端审计接口处理器
 */
func NewAuditHandler(auditService *services.AdminAuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

/**
 * AuditLogs 查询后台普通审计日志。
 *
 * @route GET /admin-api/audit-logs
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission audit:view
 */
func (h *AuditHandler) AuditLogs(c *gin.Context) {
	var query admindto.AuditLogListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.auditService.AuditLogs(c.Request.Context(), query, hasPermission(c, "audit:sensitive_view"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * RiskLogs 查询后台高危操作日志。
 *
 * @route GET /admin-api/risk-logs
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission audit:view
 */
func (h *AuditHandler) RiskLogs(c *gin.Context) {
	var query admindto.RiskLogListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.auditService.RiskLogs(c.Request.Context(), query, hasPermission(c, "audit:sensitive_view"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func hasPermission(c *gin.Context, permissionCode string) bool {
	for _, code := range middleware.CurrentAdminPermissionCodes(c) {
		if code == permissionCode {
			return true
		}
	}
	return false
}

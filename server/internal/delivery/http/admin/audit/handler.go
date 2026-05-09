package audit

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/rbac"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	auditusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/audit"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

/**
 * AdminAuditHandler 处理普通后台操作日志接口。
 */
type AdminAuditHandler struct {
	adminAuditService *auditusecase.AdminAuditService
	permissionCodes   func(*gin.Context) []string
}

/**
 * NewAdminAuditHandler 创建普通后台操作日志接口处理器。
 *
 * @param adminAuditService 普通后台操作日志服务
 * @return *AdminAuditHandler 普通后台操作日志接口处理器
 */
func NewAdminAuditHandler(adminAuditService *auditusecase.AdminAuditService, permissionCodes func(*gin.Context) []string) *AdminAuditHandler {
	return &AdminAuditHandler{
		adminAuditService: adminAuditService,
		permissionCodes:   permissionCodes,
	}
}

/**
 * Logs 分页查询普通后台操作日志。
 *
 * @route GET /admin-api/audit-logs
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission audit-log:view
 */
func (h *AdminAuditHandler) Logs(c *gin.Context) {
	var query admindto.AuditLogListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	includeSensitive := rbac.HasPermissionCode(
		h.permissionCodes(c),
		"audit-log:sensitive-view",
	)
	result, err := h.adminAuditService.AuditLogs(c.Request.Context(), query, includeSensitive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

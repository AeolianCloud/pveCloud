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
 * AdminRoleHandler 处理管理端角色和权限码接口。
 */
type AdminRoleHandler struct {
	adminRoleService *services.AdminRoleService
}

/**
 * NewAdminRoleHandler 创建管理端角色接口处理器。
 *
 * @param adminRoleService 管理端角色服务
 * @return *AdminRoleHandler 管理端角色接口处理器
 */
func NewAdminRoleHandler(adminRoleService *services.AdminRoleService) *AdminRoleHandler {
	return &AdminRoleHandler{adminRoleService: adminRoleService}
}

/**
 * Roles 分页查询管理端角色。
 *
 * @route GET /admin-api/admin-roles
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminRoleHandler) Roles(c *gin.Context) {
	var query admindto.AdminRoleListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.adminRoleService.Roles(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * CreateRole 创建管理端角色。
 *
 * @route POST /admin-api/admin-roles
 * @request {"code":"ops","name":"运营","status":"active","permission_codes":["dashboard:view"]}
 * @response 200 {"code":0,"message":"成功","data":{"id":2,"code":"ops"}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminRoleHandler) CreateRole(c *gin.Context) {
	var req admindto.AdminRoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.adminRoleService.CreateRole(c.Request.Context(), operatorID, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * RoleDetail 查看管理端角色详情。
 *
 * @route GET /admin-api/admin-roles/{id}
 * @response 200 {"code":0,"message":"成功","data":{"id":2,"code":"ops"}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminRoleHandler) RoleDetail(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	result, err := h.adminRoleService.RoleDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * UpdateRole 更新管理端角色。
 *
 * @route PATCH /admin-api/admin-roles/{id}
 * @request {"name":"运营","status":"active","permission_codes":["dashboard:view"]}
 * @response 200 {"code":0,"message":"成功","data":{"id":2,"code":"ops"}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminRoleHandler) UpdateRole(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var req admindto.AdminRoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.adminRoleService.UpdateRole(c.Request.Context(), operatorID, id, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Permissions 查询系统权限码分组。
 *
 * @route GET /admin-api/admin-permissions
 * @response 200 {"code":0,"message":"成功","data":[]}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminRoleHandler) Permissions(c *gin.Context) {
	var query admindto.AdminPermissionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.adminRoleService.Permissions(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

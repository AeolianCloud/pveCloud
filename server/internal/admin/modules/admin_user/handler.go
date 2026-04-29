package adminuser

import (
	"strconv"

	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"

	"github.com/AeolianCloud/pveCloud/server/internal/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
)

/**
 * AdminUserHandler 处理管理员账号管理接口。
 */
type AdminUserHandler struct {
	adminUserService *AdminUserService
}

/**
 * NewAdminUserHandler 创建管理员账号接口处理器。
 *
 * @param adminUserService 管理员账号服务
 * @return *AdminUserHandler 管理员账号接口处理器
 */
func NewAdminUserHandler(adminUserService *AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{adminUserService: adminUserService}
}

/**
 * List 分页查询管理员账号。
 *
 * @route GET /admin-api/admin-users
 * @response 200 {"code":0,"message":"成功","data":{"list":[],"total":0,"page":1,"per_page":15,"last_page":0}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminUserHandler) List(c *gin.Context) {
	var query admindto.AdminUserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}

	result, err := h.adminUserService.List(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Create 创建管理员账号。
 *
 * @route POST /admin-api/admin-users
 * @request {"username":"ops","email":"ops@example.com","display_name":"运营管理员","password":"password","status":"active","role_ids":[1]}
 * @response 200 {"code":0,"message":"成功","data":{"id":2,"username":"ops"}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminUserHandler) Create(c *gin.Context) {
	var req admindto.AdminUserCreateRequest
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
	result, err := h.adminUserService.Create(c.Request.Context(), operatorID, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Detail 查看管理员账号详情。
 *
 * @route GET /admin-api/admin-users/{id}
 * @response 200 {"code":0,"message":"成功","data":{"id":2,"username":"ops"}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminUserHandler) Detail(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	result, err := h.adminUserService.Detail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * Update 更新管理员账号。
 *
 * @route PATCH /admin-api/admin-users/{id}
 * @request {"display_name":"运营管理员","status":"active","role_ids":[1]}
 * @response 200 {"code":0,"message":"成功","data":{"id":2,"username":"ops"}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminUserHandler) Update(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var req admindto.AdminUserUpdateRequest
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
	result, err := h.adminUserService.Update(c.Request.Context(), operatorID, id, req, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * ResetPassword 重置管理员密码。
 *
 * @route POST /admin-api/admin-users/{id}/password
 * @request {"password":"new-password"}
 * @response 200 {"code":0,"message":"成功","data":{}}
 * @auth admin jwt, permission admin:manage
 */
func (h *AdminUserHandler) ResetPassword(c *gin.Context) {
	id, ok := adminPathID(c)
	if !ok {
		return
	}
	var req admindto.AdminUserPasswordRequest
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
	if err := h.adminUserService.ResetPassword(c.Request.Context(), operatorID, id, req, c.ClientIP(), c.Request.UserAgent()); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

func adminPathID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, apperrors.ErrValidation.WithMessage("资源 ID 格式错误"))
		return 0, false
	}
	return id, true
}

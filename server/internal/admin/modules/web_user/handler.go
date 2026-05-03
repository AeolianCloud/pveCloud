package webuser

import (
	"github.com/gin-gonic/gin"

	admindto "github.com/AeolianCloud/pveCloud/server/internal/admin/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/middleware"
	"github.com/AeolianCloud/pveCloud/server/internal/admin/support"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
)

/**
 * WebUserHandler 处理后台 Web 用户管理接口。
 */
type WebUserHandler struct {
	service *WebUserService
}

/**
 * NewWebUserHandler 创建后台 Web 用户管理接口处理器。
 */
func NewWebUserHandler(service *WebUserService) *WebUserHandler {
	return &WebUserHandler{service: service}
}

func (h *WebUserHandler) Users(c *gin.Context) {
	var query admindto.WebUserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Users(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WebUserHandler) CreateUser(c *gin.Context) {
	var req admindto.WebUserCreateRequest
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
	result, err := h.service.CreateUser(c.Request.Context(), operatorID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WebUserHandler) UserDetail(c *gin.Context) {
	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}
	result, err := h.service.UserDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WebUserHandler) UpdateUser(c *gin.Context) {
	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.WebUserUpdateRequest
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
	result, err := h.service.UpdateUser(c.Request.Context(), operatorID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WebUserHandler) ResetPassword(c *gin.Context) {
	id, ok := support.AdminPathID(c)
	if !ok {
		return
	}
	var req admindto.WebUserPasswordRequest
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
	if err := h.service.ResetPassword(c.Request.Context(), operatorID, id, req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

func (h *WebUserHandler) Sessions(c *gin.Context) {
	var query admindto.WebUserSessionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(query); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	result, err := h.service.Sessions(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *WebUserHandler) RevokeSession(c *gin.Context) {
	var req admindto.WebUserSessionUpdateRequest
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
	if err := h.service.RevokeSession(c.Request.Context(), operatorID, c.Param("session_id")); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

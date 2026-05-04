package userprofile

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/web/dto"
	"github.com/AeolianCloud/pveCloud/server/internal/web/middleware"
)

/**
 * UserProfileHandler 处理当前用户资料接口。
 */
type UserProfileHandler struct {
	service *UserProfileService
}

/**
 * NewUserProfileHandler 创建当前用户资料接口处理器。
 */
func NewUserProfileHandler(service *UserProfileService) *UserProfileHandler {
	return &UserProfileHandler{service: service}
}

/**
 * UpdateProfile 更新当前用户基础资料。
 */
func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
	var req webdto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentUserSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	result, err := h.service.UpdateProfile(c.Request.Context(), userID, session.SessionID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

/**
 * ChangePassword 修改当前用户密码。
 */
func (h *UserProfileHandler) ChangePassword(c *gin.Context) {
	var req webdto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	session, ok := middleware.CurrentUserSession(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), userID, session.SessionID, req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{})
}

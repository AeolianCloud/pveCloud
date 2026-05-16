package logs

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	logsusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/logs"
)

type Handler struct {
	service *logsusecase.Service
}

func NewHandler(service *logsusecase.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UserSecurity(c *gin.Context) {
	var query admindto.UserSecurityLogQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.UserSecurityLogs(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) UserBusiness(c *gin.Context) {
	var query admindto.UserBusinessLogQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.UserBusinessLogs(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) FrontendErrors(c *gin.Context) {
	var query admindto.FrontendErrorLogQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.FrontendErrorLogs(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) BackendRuntime(c *gin.Context) {
	var query admindto.BackendRuntimeLogQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.BackendRuntimeLogs(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func bindQuery(c *gin.Context, target any) bool {
	if err := c.ShouldBindQuery(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return false
	}
	if err := validator.Struct(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return false
	}
	return true
}

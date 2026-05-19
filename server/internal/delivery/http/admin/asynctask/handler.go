package asynctask

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	asynctaskusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/asynctask"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
)

type Handler struct{ service *asynctaskusecase.Service }

func NewHandler(service *asynctaskusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	var query admindto.AsyncTaskListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Retry(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var req admindto.AsyncTaskRetryRequest
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if !bindJSON(c, &req) {
			return
		}
	}
	result, err := h.service.Retry(c.Request.Context(), operatorID, c.Param("task_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func currentAdminID(c *gin.Context) (uint64, bool) {
	operatorID, ok := middleware.CurrentAdminID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return operatorID, true
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

func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数格式错误"))
		return false
	}
	if err := validator.Struct(target); err != nil {
		response.Error(c, apperrors.ErrValidation.WithMessage("请求参数校验失败"))
		return false
	}
	return true
}

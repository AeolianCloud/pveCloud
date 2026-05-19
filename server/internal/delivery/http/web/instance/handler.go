package instance

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	instanceusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/instance"
)

type Handler struct{ service *instanceusecase.Service }

func NewHandler(service *instanceusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var query webdto.InstanceListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.List(c.Request.Context(), userID, query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Detail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Detail(c.Request.Context(), userID, c.Param("instance_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Start(c *gin.Context) {
	h.operate(c, h.service.Start)
}

func (h *Handler) Stop(c *gin.Context) {
	h.operate(c, h.service.Stop)
}

func (h *Handler) CreateRenewalOrder(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req webdto.RenewalOrderCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.CreateRenewalOrder(c.Request.Context(), userID, c.Param("instance_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) operate(c *gin.Context, fn func(context.Context, uint64, string) (webdto.InstanceDetail, error)) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := fn(c.Request.Context(), userID, c.Param("instance_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func currentUserID(c *gin.Context) (uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return userID, true
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

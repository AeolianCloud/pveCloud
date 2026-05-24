package payment

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/admin/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	paymentusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/payment"
)

type Handler struct{ service *paymentusecase.Service }

func NewHandler(service *paymentusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	var query admindto.PaymentListQuery
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

func (h *Handler) Detail(c *gin.Context) {
	result, err := h.service.Detail(c.Request.Context(), c.Param("payment_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Refunds(c *gin.Context) {
	var query admindto.RefundListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.Refunds(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Sync(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.Sync(c.Request.Context(), operatorID, c.Param("payment_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreateRefund(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	var req admindto.RefundCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.CreateRefund(c.Request.Context(), operatorID, c.Param("payment_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) RetryProvision(c *gin.Context) {
	operatorID, ok := currentAdminID(c)
	if !ok {
		return
	}
	result, err := h.service.RetryProvision(c.Request.Context(), operatorID, c.Param("payment_no"))
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

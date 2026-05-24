package payment

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	paymentusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/payment"
)

type Handler struct{ service *paymentusecase.Service }

func NewHandler(service *paymentusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req webdto.PaymentCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.Create(c.Request.Context(), userID, c.Param("order_no"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Show(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Get(c.Request.Context(), userID, c.Param("payment_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Callback(c *gin.Context) {
	provider := strings.TrimSpace(c.Param("provider"))
	if err := h.service.HandleCallback(c.Request.Context(), provider, c.Request); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"accepted": true})
}

func currentUserID(c *gin.Context) (uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return userID, true
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

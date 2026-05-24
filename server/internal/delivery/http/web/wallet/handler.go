package wallet

import (
	"github.com/gin-gonic/gin"

	"github.com/AeolianCloud/pveCloud/server/internal/delivery/http/web/middleware"
	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	webdto "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/dto"
	walletusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/web/wallet"
)

type Handler struct{ service *walletusecase.Service }

func NewHandler(service *walletusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) Show(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.Get(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Ledger(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var query webdto.WalletLedgerQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.Ledger(c.Request.Context(), userID, query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreateRecharge(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	var req webdto.WalletRechargeCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.service.CreateRecharge(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Recharge(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetRecharge(c.Request.Context(), userID, c.Param("recharge_no"))
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

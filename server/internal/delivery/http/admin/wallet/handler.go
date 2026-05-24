package wallet

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/AeolianCloud/pveCloud/server/internal/shared/errors"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/response"
	"github.com/AeolianCloud/pveCloud/server/internal/shared/validator"
	admindto "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/dto"
	walletusecase "github.com/AeolianCloud/pveCloud/server/internal/usecase/admin/wallet"
)

type Handler struct{ service *walletusecase.Service }

func NewHandler(service *walletusecase.Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(c *gin.Context) {
	var query admindto.WalletListQuery
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
	result, err := h.service.Detail(c.Request.Context(), c.Param("wallet_no"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Ledger(c *gin.Context) {
	var query admindto.WalletLedgerListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.Ledger(c.Request.Context(), query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Recharges(c *gin.Context) {
	var query admindto.WalletRechargeListQuery
	if !bindQuery(c, &query) {
		return
	}
	result, err := h.service.Recharges(c.Request.Context(), query)
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

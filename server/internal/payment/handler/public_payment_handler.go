package handler

import (
	"context"
	"encoding/json"
	"net/http"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

type PublicService interface {
	CreatePendingPayment(ctx context.Context, orderID uint64, payableAmount int64) (payment.PaymentOrder, error)
}

type PublicHandler struct {
	svc PublicService
}

type CreatePaymentRequest struct {
	OrderID       uint64 `json:"order_id"`
	PayableAmount int64  `json:"payable_amount"`
}

func NewPublicHandler(svc PublicService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	resp, err := h.svc.CreatePendingPayment(r.Context(), req.OrderID, req.PayableAmount)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

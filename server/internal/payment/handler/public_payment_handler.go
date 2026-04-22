package handler

import (
	"context"
	"net/http"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/payment"
)

// PaymentQueryService provides read-only access to payment orders.
type PaymentQueryService interface {
	GetPaymentOrder(ctx context.Context, paymentOrderNo string) (payment.PaymentOrder, error)
}

// PublicPaymentHandler exposes payment status query for authenticated users.
type PublicPaymentHandler struct {
	svc PaymentQueryService
}

// NewPublicPaymentHandler creates a new PublicPaymentHandler.
func NewPublicPaymentHandler(svc PaymentQueryService) *PublicPaymentHandler {
	return &PublicPaymentHandler{svc: svc}
}

// GetPaymentStatus handles GET /payments/{paymentOrderNo}
func (h *PublicPaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	paymentOrderNo := r.PathValue("paymentOrderNo")
	if paymentOrderNo == "" {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	order, err := h.svc.GetPaymentOrder(r.Context(), paymentOrderNo)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, order)
}

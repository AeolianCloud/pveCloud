package handler

import (
	"context"
	"io"
	"net/http"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
)

type CallbackService interface {
	MarkPaymentSuccess(ctx context.Context, paymentOrderNo string, rawPayload []byte) error
}

type CallbackHandler struct {
	svc CallbackService
}

func NewCallbackHandler(svc CallbackService) *CallbackHandler {
	return &CallbackHandler{svc: svc}
}

func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	paymentOrderNo := r.URL.Query().Get("payment_order_no")
	if paymentOrderNo == "" {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	rawPayload, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	if err := h.svc.MarkPaymentSuccess(r.Context(), paymentOrderNo, rawPayload); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

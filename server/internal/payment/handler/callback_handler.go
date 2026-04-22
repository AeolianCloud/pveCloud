package handler

import (
	"context"
	"io"
	"net/http"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
)

type CallbackService interface {
	MarkPaymentSuccess(ctx context.Context, paymentOrderNo string, rawPayload []byte) error
}

type ProviderVerifier interface {
	VerifyCallback(r *http.Request) (paymentOrderNo string, err error)
}

type CallbackHandler struct {
	svc      CallbackService
	verifier ProviderVerifier
}

func NewCallbackHandler(svc CallbackService, verifier ProviderVerifier) *CallbackHandler {
	return &CallbackHandler{svc: svc, verifier: verifier}
}

func (h *CallbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	paymentOrderNo, err := h.verifier.VerifyCallback(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	rawPayload, err := io.ReadAll(r.Body)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	if err := h.svc.MarkPaymentSuccess(r.Context(), paymentOrderNo, rawPayload); err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

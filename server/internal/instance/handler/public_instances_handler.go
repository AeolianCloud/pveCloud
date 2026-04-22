package handler

import (
	"context"
	"net/http"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
)

type PublicService interface {
	ListByUser(ctx context.Context, userID uint64) ([]instance.Instance, error)
}

type PublicHandler struct {
	svc PublicService
}

func NewPublicHandler(svc PublicService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	items, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

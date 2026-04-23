package handler

import (
	"context"
	"net/http"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
)

type AdminService interface {
	ListAll(ctx context.Context) ([]order.Order, error)
}

type AdminHandler struct {
	svc AdminService
}

func NewAdminHandler(svc AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

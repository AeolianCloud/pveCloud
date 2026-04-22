package handler

import (
	"context"
	"net/http"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
)

type PublicService interface {
	ListSaleableProducts(ctx context.Context) ([]catalog.SaleableProduct, error)
}

type PublicHandler struct {
	svc PublicService
}

func NewPublicHandler(svc PublicService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListSaleableProducts(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

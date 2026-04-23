package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
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

func (h *PublicHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID, err := strconv.ParseUint(r.PathValue("productID"), 10, 64)
	if err != nil || productID == 0 {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	items, err := h.svc.ListSaleableProducts(r.Context())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	for _, item := range items {
		if item.Product.ID == productID {
			httpx.WriteJSON(w, http.StatusOK, item)
			return
		}
	}

	httpx.WriteError(w, errorsx.ErrBadRequest)
}

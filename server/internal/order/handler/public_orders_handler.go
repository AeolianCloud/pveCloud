package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/order"
)

type PublicService interface {
	CreateOrder(ctx context.Context, in order.CreateInput) (order.CreateResult, error)
	ListByUser(ctx context.Context, userID uint64) ([]order.Order, error)
}

type PublicHandler struct {
	svc PublicService
}

type CreateOrderRequest struct {
	SKUID    uint64 `json:"sku_id"`
	RegionID uint64 `json:"region_id"`
	Cycle    string `json:"cycle"`
}

func NewPublicHandler(svc PublicService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	resp, err := h.svc.CreateOrder(r.Context(), order.CreateInput{
		UserID:   auth.MustUserID(r.Context()),
		SKUID:    req.SKUID,
		RegionID: req.RegionID,
		Cycle:    req.Cycle,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *PublicHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListByUser(r.Context(), auth.MustUserID(r.Context()))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

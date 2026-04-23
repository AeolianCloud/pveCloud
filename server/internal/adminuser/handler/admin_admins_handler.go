package handler

import (
	"context"
	"net/http"
	"strconv"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/adminuser"
)

type AdminAdminsService interface {
	ListAdmins(ctx context.Context, limit int) ([]adminuser.AdminRow, error)
}

type AdminAdminsHandler struct {
	svc AdminAdminsService
}

func NewAdminAdminsHandler(svc AdminAdminsService) *AdminAdminsHandler {
	return &AdminAdminsHandler{svc: svc}
}

func (h *AdminAdminsHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.svc.ListAdmins(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if items == nil {
		items = []adminuser.AdminRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

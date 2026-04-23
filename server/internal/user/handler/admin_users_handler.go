package handler

import (
	"context"
	"net/http"
	"strconv"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/user"
)

type AdminUserService interface {
	ListUsers(ctx context.Context, limit int) ([]user.UserRow, error)
}

type AdminUsersHandler struct {
	svc AdminUserService
}

func NewAdminUsersHandler(svc AdminUserService) *AdminUsersHandler {
	return &AdminUsersHandler{svc: svc}
}

func (h *AdminUsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.svc.ListUsers(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if items == nil {
		items = []user.UserRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

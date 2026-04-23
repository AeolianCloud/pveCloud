package handler

import (
	"context"
	"net/http"
	"strconv"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/task"
)

type AdminService interface {
	ListTasks(ctx context.Context, limit int) ([]task.Task, error)
}

type AdminHandler struct {
	svc AdminService
}

func NewAdminHandler(svc AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.svc.ListTasks(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

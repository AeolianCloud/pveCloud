package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/notification"
)

type NoticeService interface {
	ListByUser(ctx context.Context, userID uint64, limit int) ([]notification.Notification, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) error
	CountUnread(ctx context.Context, userID uint64) (int, error)
}

type PublicHandler struct {
	svc NoticeService
}

func NewPublicHandler(svc NoticeService) *PublicHandler {
	return &PublicHandler{svc: svc}
}

func (h *PublicHandler) ListNotices(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())

	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.svc.ListByUser(r.Context(), userID, limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if items == nil {
		items = []notification.Notification{}
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *PublicHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := auth.MustUserID(r.Context())
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}
	if err := h.svc.MarkRead(r.Context(), id, userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

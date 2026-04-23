package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/instance"
)

type DetailService interface {
	ListByUser(ctx context.Context, userID uint64) ([]instance.Instance, error)
}

type DetailHandler struct {
	svc DetailService
}

func NewDetailHandler(svc DetailService) *DetailHandler {
	return &DetailHandler{svc: svc}
}

func (h *DetailHandler) GetMineByID(w http.ResponseWriter, r *http.Request) {
	instanceID, err := strconv.ParseUint(r.PathValue("instanceID"), 10, 64)
	if err != nil || instanceID == 0 {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	items, err := h.svc.ListByUser(r.Context(), auth.MustUserID(r.Context()))
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	for _, item := range items {
		if item.ID == instanceID {
			httpx.WriteJSON(w, http.StatusOK, item)
			return
		}
	}

	httpx.WriteError(w, errorsx.ErrBadRequest)
}

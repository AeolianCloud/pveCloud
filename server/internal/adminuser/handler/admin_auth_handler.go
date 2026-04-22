package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AeolianCloud/pveCloud/server/internal/adminuser"
	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (adminuser.AuthResponse, error)
}

type Handler struct {
	svc AuthService
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAuthHandler(svc AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}
	resp, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

package handler

import (
	"context"
	"encoding/json"
	"net/http"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/user"
)

type AuthService interface {
	Login(ctx context.Context, phone, password string) (user.AuthResponse, error)
}

type Handler struct {
	svc AuthService
}

type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

func NewAuthHandler[T interface {
	Login(ctx context.Context, phone, password string) (user.AuthResponse, error)
}](svc T) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}
	resp, err := h.svc.Login(r.Context(), req.Phone, req.Password)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

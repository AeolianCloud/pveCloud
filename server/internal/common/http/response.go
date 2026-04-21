package httpx

import (
	"encoding/json"
	"net/http"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, err error) {
	appErr, ok := err.(*errorsx.Error)
	if !ok {
		appErr = errorsx.ErrInternal
	}
	WriteJSON(w, appErr.Status, map[string]any{
		"error": map[string]string{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
}

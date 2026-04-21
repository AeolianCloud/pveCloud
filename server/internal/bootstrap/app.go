package bootstrap

import (
	"net/http"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
)

func NewServer(addr string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

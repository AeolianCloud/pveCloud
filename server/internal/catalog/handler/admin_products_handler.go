package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	errorsx "github.com/AeolianCloud/pveCloud/server/internal/common/errors"
	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/catalog"
)

type AdminService interface {
	CreateSKU(ctx context.Context, productID uint64, in catalog.CreateSKUInput) (catalog.SKU, error)
}

type AdminHandler struct {
	svc AdminService
}

type CreateSKURequest struct {
	SKUName       string `json:"sku_name"`
	CPUCores      int    `json:"cpu_cores"`
	MemoryMB      int    `json:"memory_mb"`
	DiskGB        int    `json:"disk_gb"`
	BandwidthMbps int    `json:"bandwidth_mbps"`
	Status        string `json:"status"`
}

func NewAdminHandler(svc AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) CreateSKU(w http.ResponseWriter, r *http.Request) {
	productID, err := parseProductID(r.URL.Path)
	if err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	var req CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, errorsx.ErrBadRequest)
		return
	}

	resp, err := h.svc.CreateSKU(r.Context(), productID, catalog.CreateSKUInput{
		SKUName:       req.SKUName,
		CPUCores:      req.CPUCores,
		MemoryMB:      req.MemoryMB,
		DiskGB:        req.DiskGB,
		BandwidthMbps: req.BandwidthMbps,
		Status:        req.Status,
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func parseProductID(path string) (uint64, error) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 3 {
		return 0, errorsx.ErrBadRequest
	}
	return strconv.ParseUint(parts[1], 10, 64)
}

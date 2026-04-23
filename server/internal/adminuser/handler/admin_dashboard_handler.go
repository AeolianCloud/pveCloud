package handler

import (
	"database/sql"
	"net/http"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
)

type DashboardHandler struct {
	db *sql.DB
}

func NewDashboardHandler(db *sql.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

type DashboardStats struct {
	TotalOrders      int `json:"total_orders"`
	PendingOrders    int `json:"pending_orders"`
	TotalInstances   int `json:"total_instances"`
	RunningInstances int `json:"running_instances"`
	TotalUsers       int `json:"total_users"`
	TotalTasks       int `json:"total_tasks"`
	PendingTasks     int `json:"pending_tasks"`
}

func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	var stats DashboardStats

	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM orders`).Scan(&stats.TotalOrders)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM orders WHERE status = 'pending_payment'`).Scan(&stats.PendingOrders)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM instances`).Scan(&stats.TotalInstances)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM instances WHERE status = 'running'`).Scan(&stats.RunningInstances)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM async_tasks`).Scan(&stats.TotalTasks)
	_ = h.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM async_tasks WHERE status = 'pending'`).Scan(&stats.PendingTasks)

	httpx.WriteJSON(w, http.StatusOK, stats)
}

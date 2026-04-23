# PVE Cloud Missing Backend Gaps Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 notices、admin/dashboard、admin/users 三个后端缺口，让前端占位页全部替换为真实数据页面。

**Architecture:** 遵循现有模块化单体模式：新建 migration、model、repository、service、handler，在 bootstrap/app.go 中装配路由。前端遵循已有 API 模块 + Composition API 页面模式。

**Tech Stack:** Go, `database/sql`, MariaDB, `net/http` ServeMux, Bun, Vue 3, Composition API, Vitest

---

## Planned File Structure

### Backend

- Create: `server/migrations/0005_notifications.sql`
- Create: `server/internal/notification/model.go`
- Create: `server/internal/notification/repository.go`
- Create: `server/internal/notification/mysql_repository.go`
- Modify: `server/internal/notification/service.go`
- Create: `server/internal/notification/handler/public_notices_handler.go`
- Create: `server/internal/user/handler/admin_users_handler.go`
- Create: `server/internal/adminuser/handler/admin_admins_handler.go`
- Modify: `server/internal/bootstrap/app.go`

### Web Frontend

- Create: `web/src/api/notice.ts`
- Modify: `web/src/views/NoticePlaceholderPage.vue` → rename to real page content
- Modify: `web/src/router/index.ts`

### Admin Frontend

- Create: `admin/src/api/dashboard.ts`
- Create: `admin/src/api/user.ts`
- Modify: `admin/src/views/DashboardPlaceholderPage.vue` → rename to real page content
- Modify: `admin/src/views/UserManagePlaceholderPage.vue` → rename to real page content
- Modify: `admin/src/router/index.ts`

---

### Task 1: Add notifications migration and notification model

**Files:**
- Create: `server/migrations/0005_notifications.sql`
- Create: `server/internal/notification/model.go`

- [ ] **Step 1: Create the notifications migration**

```sql
-- server/migrations/0005_notifications.sql

CREATE TABLE notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  title VARCHAR(255) NOT NULL COMMENT '通知标题',
  body TEXT NOT NULL COMMENT '通知正文',
  type VARCHAR(32) NOT NULL COMMENT '通知类型：system-系统通知，provision-开通通知，billing-账单通知',
  is_read TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
  created_at DATETIME(3) NOT NULL COMMENT '创建时间',
  updated_at DATETIME(3) NOT NULL COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_user_id_created (user_id, created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户通知表';
```

- [ ] **Step 2: Create the notification model**

```go
// server/internal/notification/model.go

package notification

import "time"

type Notification struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
```

- [ ] **Step 3: Run build verification**

Run: `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`
Expected: PASS (no import changes yet)

- [ ] **Step 4: Commit**

```bash
git add server/migrations/0005_notifications.sql server/internal/notification/model.go
git commit -m "feat: add notifications table and model"
```

---

### Task 2: Implement notification repository

**Files:**
- Create: `server/internal/notification/repository.go`
- Create: `server/internal/notification/mysql_repository.go`

- [ ] **Step 1: Create the repository interface**

```go
// server/internal/notification/repository.go

package notification

import "context"

type Repository interface {
	Create(ctx context.Context, n Notification) (Notification, error)
	ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) error
	CountUnread(ctx context.Context, userID uint64) (int, error)
}
```

- [ ] **Step 2: Create the MySQL implementation**

```go
// server/internal/notification/mysql_repository.go

package notification

import (
	"context"
	"database/sql"
	"time"
)

type MySQLRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewMySQLRepository(db *sql.DB) *MySQLRepository {
	return &MySQLRepository{db: db, now: time.Now}
}

func (r *MySQLRepository) Create(ctx context.Context, n Notification) (Notification, error) {
	now := r.now().UTC()
	result, err := r.db.ExecContext(ctx, `
INSERT INTO notifications (user_id, title, body, type, is_read, created_at, updated_at)
VALUES (?, ?, ?, ?, 0, ?, ?)
`, n.UserID, n.Title, n.Body, n.Type, now, now)
	if err != nil {
		return Notification{}, err
	}
	id, _ := result.LastInsertId()
	n.ID = uint64(id)
	n.CreatedAt = now
	return n, nil
}

func (r *MySQLRepository) ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT id, user_id, title, body, type, is_read, created_at
FROM notifications
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ?
`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Notification
	for rows.Next() {
		var n Notification
		var isRead int
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &isRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		n.IsRead = isRead == 1
		items = append(items, n)
	}
	return items, nil
}

func (r *MySQLRepository) MarkRead(ctx context.Context, id uint64, userID uint64) error {
	_, err := r.db.ExecContext(ctx, `
UPDATE notifications SET is_read = 1, updated_at = ? WHERE id = ? AND user_id = ?
`, r.now().UTC(), id, userID)
	return err
}

func (r *MySQLRepository) CountUnread(ctx context.Context, userID uint64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0
`, userID).Scan(&count)
	return count, err
}
```

- [ ] **Step 3: Run build verification**

Run: `go -C server build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/notification/repository.go server/internal/notification/mysql_repository.go
git commit -m "feat: add notification repository with MySQL implementation"
```

---

### Task 3: Upgrade notification service to persist notifications

**Files:**
- Modify: `server/internal/notification/service.go`

- [ ] **Step 1: Rewrite the notification service**

Replace the entire content of `server/internal/notification/service.go` with:

```go
// server/internal/notification/service.go

package notification

import "context"

type Repo interface {
	Create(ctx context.Context, n Notification) (Notification, error)
	ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error)
	MarkRead(ctx context.Context, id uint64, userID uint64) error
	CountUnread(ctx context.Context, userID uint64) (int, error)
}

type Service struct {
	repo Repo
}

func NewService(repo Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) SendProvisionSuccess(ctx context.Context, userID uint64, instanceNo string) error {
	_, err := s.repo.Create(ctx, Notification{
		UserID: userID,
		Title:  "实例开通成功",
		Body:   "您的实例 " + instanceNo + " 已成功开通。",
		Type:   "provision",
	})
	return err
}

func (s *Service) SendProvisionFailure(ctx context.Context, userID uint64, orderID uint64) error {
	_, err := s.repo.Create(ctx, Notification{
		UserID: userID,
		Title:  "实例开通失败",
		Body:   "您的订单关联实例开通失败，请稍后重试或联系客服。",
		Type:   "provision",
	})
	return err
}

func (s *Service) ListByUser(ctx context.Context, userID uint64, limit int) ([]Notification, error) {
	return s.repo.ListByUser(ctx, userID, limit)
}

func (s *Service) MarkRead(ctx context.Context, id uint64, userID uint64) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *Service) CountUnread(ctx context.Context, userID uint64) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}
```

- [ ] **Step 2: Update the call site — the instance service expects `NewService()` with no args currently. Update bootstrap to pass the repo.**

In `server/internal/bootstrap/app.go`, every place that calls `notification.NewService()` must change to:

```go
notificationRepo := notification.NewMySQLRepository(db)
notificationSvc := notification.NewService(notificationRepo)
```

This happens in two places: the `public-api` case and the `admin-api` case.

- [ ] **Step 3: Run build verification**

Run: `go -C server build ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/notification/service.go server/internal/bootstrap/app.go
git commit -m "feat: upgrade notification service to persist via MariaDB"
```

---

### Task 4: Add notification handler and wire public route

**Files:**
- Create: `server/internal/notification/handler/public_notices_handler.go`
- Modify: `server/internal/bootstrap/app.go`

- [ ] **Step 1: Create the public notices handler**

```go
// server/internal/notification/handler/public_notices_handler.go

package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AeolianCloud/pveCloud/server/internal/auth"
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
	if userID == 0 {
		httpx.WriteError(w, ErrBadRequest("missing user"))
		return
	}

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
		httpx.WriteError(w, ErrBadRequest("invalid id"))
		return
	}
	if err := h.svc.MarkRead(r.Context(), id, userID); err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func ErrBadRequest(msg string) error {
	return &badRequestError{msg: msg}
}

type badRequestError struct {
	msg string
}

func (e *badRequestError) Error() string { return e.msg }
```

Wait — the codebase already has `errorsx.ErrBadRequest`. Use that instead of defining a custom error. Let me fix:

```go
// server/internal/notification/handler/public_notices_handler.go

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
```

- [ ] **Step 2: Wire the handler in bootstrap/app.go**

In the `public-api` case in `newHTTPApp`, after the existing handler instantiations, add:

```go
noticeHandler := notificationhandler.NewPublicHandler(notificationSvc)
```

And add the import alias at the top:

```go
notificationhandler "github.com/AeolianCloud/pveCloud/server/internal/notification/handler"
```

Then register the routes (authenticated):

```go
mux.Handle("GET /notices", userAuth(http.HandlerFunc(noticeHandler.ListNotices)))
mux.Handle("PUT /notices/{id}/read", userAuth(http.HandlerFunc(noticeHandler.MarkRead)))
```

- [ ] **Step 3: Run build verification**

Run: `go -C server build ./cmd/public-api`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/notification/handler/ server/internal/bootstrap/app.go
git commit -m "feat: add public notices API endpoint"
```

---

### Task 5: Add admin users handler (list users + list admins)

**Files:**
- Create: `server/internal/user/handler/admin_users_handler.go`
- Create: `server/internal/adminuser/handler/admin_admins_handler.go`
- Modify: `server/internal/bootstrap/app.go`

- [ ] **Step 1: Add ListUsers method to user service**

Append to `server/internal/user/service.go`:

```go
type UserRow struct {
	ID        uint64 `json:"id"`
	UserNo    string `json:"user_no"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s *Service) ListUsers(ctx context.Context, limit int) ([]UserRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, user_no, email, phone, status, created_at
FROM users
ORDER BY id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UserRow
	for rows.Next() {
		var u UserRow
		if err := rows.Scan(&u.ID, &u.UserNo, &u.Email, &u.Phone, &u.Status, &u.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, nil
}
```

- [ ] **Step 2: Create the admin users handler**

```go
// server/internal/user/handler/admin_users_handler.go

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
```

- [ ] **Step 3: Add ListAdmins method to adminuser service**

Append to `server/internal/adminuser/service.go`:

```go
type AdminRow struct {
	ID        uint64 `json:"id"`
	AdminNo   string `json:"admin_no"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (s *Service) ListAdmins(ctx context.Context, limit int) ([]AdminRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, admin_no, username, status, created_at
FROM admins
ORDER BY id DESC
LIMIT ?
`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []AdminRow
	for rows.Next() {
		var a AdminRow
		if err := rows.Scan(&a.ID, &a.AdminNo, &a.Username, &a.Status, &a.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, nil
}
```

- [ ] **Step 4: Create the admin admins handler**

```go
// server/internal/adminuser/handler/admin_admins_handler.go

package handler

import (
	"context"
	"net/http"
	"strconv"

	httpx "github.com/AeolianCloud/pveCloud/server/internal/common/http"
	"github.com/AeolianCloud/pveCloud/server/internal/adminuser"
)

type AdminAdminsService interface {
	ListAdmins(ctx context.Context, limit int) ([]adminuser.AdminRow, error)
}

type AdminAdminsHandler struct {
	svc AdminAdminsService
}

func NewAdminAdminsHandler(svc AdminAdminsService) *AdminAdminsHandler {
	return &AdminAdminsHandler{svc: svc}
}

func (h *AdminAdminsHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	items, err := h.svc.ListAdmins(r.Context(), limit)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if items == nil {
		items = []adminuser.AdminRow{}
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}
```

- [ ] **Step 5: Wire in bootstrap/app.go — admin-api case**

In the `admin-api` case of `newHTTPApp`, add handler instantiation and routes:

```go
adminUsersHandler := userhandler.NewAdminUsersHandler(userSvc)
adminAdminsHandler := adminhandler.NewAdminAdminsHandler(adminSvc)
```

And register:

```go
mux.Handle("GET /users", adminAuth(http.HandlerFunc(adminUsersHandler.ListUsers)))
mux.Handle("GET /admins", adminAuth(http.HandlerFunc(adminAdminsHandler.ListAdmins)))
```

Note: `userSvc` needs to be available in the `admin-api` case. Currently `user.NewService` is only in `public-api`. Add it to `admin-api` as well:

```go
webSigner := auth.NewJWTSigner(cfg.JWTWebSecret)
userSvc := user.NewService(db, webSigner)
```

This doesn't create a public route — it's only used for the admin handler to query users.

- [ ] **Step 6: Run build verification**

Run: `go -C server build ./cmd/admin-api`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add server/internal/user/service.go server/internal/user/handler/admin_users_handler.go server/internal/adminuser/service.go server/internal/adminuser/handler/admin_admins_handler.go server/internal/bootstrap/app.go
git commit -m "feat: add admin users and admins list endpoints"
```

---

### Task 6: Add admin dashboard stats endpoint

**Files:**
- Modify: `server/internal/bootstrap/app.go`

The dashboard aggregates counts from existing tables — no new migration needed. We add a simple handler that queries counts directly.

- [ ] **Step 1: Create a dashboard handler inline in bootstrap or as a small package**

Create `server/internal/adminuser/handler/admin_dashboard_handler.go`:

```go
// server/internal/adminuser/handler/admin_dashboard_handler.go

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
	TotalOrders     int `json:"total_orders"`
	PendingOrders   int `json:"pending_orders"`
	TotalInstances  int `json:"total_instances"`
	RunningInstances int `json:"running_instances"`
	TotalUsers      int `json:"total_users"`
	TotalTasks      int `json:"total_tasks"`
	PendingTasks    int `json:"pending_tasks"`
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
```

- [ ] **Step 2: Wire in bootstrap/app.go — admin-api case**

In the `admin-api` case, add:

```go
dashboardHandler := adminhandler.NewDashboardHandler(db)
```

And register:

```go
mux.Handle("GET /dashboard", adminAuth(http.HandlerFunc(dashboardHandler.Stats)))
```

- [ ] **Step 3: Run build verification**

Run: `go -C server build ./cmd/admin-api`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add server/internal/adminuser/handler/admin_dashboard_handler.go server/internal/bootstrap/app.go
git commit -m "feat: add admin dashboard stats endpoint"
```

---

### Task 7: Wire web frontend notices page to real API

**Files:**
- Create: `web/src/api/notice.ts`
- Modify: `web/src/views/NoticePlaceholderPage.vue`
- Modify: `web/src/router/index.ts`

- [ ] **Step 1: Create the notice API module**

```typescript
// web/src/api/notice.ts

import { request } from '../lib/http'

export interface Notice {
  id: number
  user_id: number
  title: string
  body: string
  type: string
  is_read: boolean
  created_at: string
}

export function listNotices(limit = 20): Promise<Notice[]> {
  return request<Notice[]>(`/notices?limit=${limit}`)
}

export function markNoticeRead(id: number): Promise<{ status: string }> {
  return request<{ status: string }>(`/notices/${id}/read`, { method: 'PUT' })
}
```

- [ ] **Step 2: Replace the placeholder page with a real page**

Replace the entire content of `web/src/views/NoticePlaceholderPage.vue` with:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listNotices, markNoticeRead, type Notice } from '../api/notice'

const notices = ref<Notice[]>([])
const loading = ref(true)
const error = ref('')

async function handleMarkRead(id: number) {
  try {
    await markNoticeRead(id)
    const n = notices.value.find(n => n.id === id)
    if (n) n.is_read = true
  } catch {
    // ignore
  }
}

onMounted(async () => {
  try {
    notices.value = await listNotices()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载通知失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="panel">
    <p class="tag">NOTICES</p>
    <h2>通知中心</h2>

    <p v-if="loading">加载中...</p>
    <p v-else-if="error" class="error">{{ error }}</p>
    <p v-else-if="notices.length === 0">暂无通知</p>

    <ul v-else class="notice-list">
      <li v-for="n in notices" :key="n.id" :class="{ unread: !n.is_read }">
        <div class="notice-header">
          <strong>{{ n.title }}</strong>
          <span class="time">{{ n.created_at }}</span>
        </div>
        <p>{{ n.body }}</p>
        <button v-if="!n.is_read" @click="handleMarkRead(n.id)">标记已读</button>
      </li>
    </ul>
  </section>
</template>

<style scoped>
.panel {
  padding: 28px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(29, 42, 51, 0.08);
}

.tag {
  margin: 0 0 8px;
  color: #9b5d32;
}

.notice-list {
  list-style: none;
  padding: 0;
}

.notice-list li {
  padding: 12px 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.notice-list li.unread {
  background: rgba(155, 93, 50, 0.06);
  padding-left: 8px;
  border-radius: 6px;
}

.notice-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 4px;
}

.time {
  color: #999;
  font-size: 0.85em;
}

button {
  margin-top: 6px;
  padding: 4px 12px;
  border: 1px solid #9b5d32;
  border-radius: 6px;
  background: transparent;
  color: #9b5d32;
  cursor: pointer;
}

.error { color: #c00; }
</style>
```

- [ ] **Step 3: Update web router to guard /notices**

In `web/src/router/index.ts`, update the `requiresAuth` check to include `/notices`:

```typescript
const requiresAuth = to.path === '/orders' || to.path.startsWith('/instances') || to.path === '/notices'
```

- [ ] **Step 4: Run tests and build**

Run: `bun --cwd web test`
Expected: PASS

Run: `bun --cwd web run build`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add web/src/api/notice.ts web/src/views/NoticePlaceholderPage.vue web/src/router/index.ts
git commit -m "feat: wire web notices page to real backend API"
```

---

### Task 8: Wire admin frontend dashboard and users pages to real APIs

**Files:**
- Create: `admin/src/api/dashboard.ts`
- Create: `admin/src/api/user.ts`
- Modify: `admin/src/views/DashboardPlaceholderPage.vue`
- Modify: `admin/src/views/UserManagePlaceholderPage.vue`

- [ ] **Step 1: Create the dashboard API module**

```typescript
// admin/src/api/dashboard.ts

import { request } from '../lib/http'

export interface DashboardStats {
  total_orders: number
  pending_orders: number
  total_instances: number
  running_instances: number
  total_users: number
  total_tasks: number
  pending_tasks: number
}

export function getDashboardStats(): Promise<DashboardStats> {
  return request<DashboardStats>('/dashboard')
}
```

- [ ] **Step 2: Create the user API module**

```typescript
// admin/src/api/user.ts

import { request } from '../lib/http'

export interface UserRow {
  id: number
  user_no: string
  email: string
  phone: string
  status: string
  created_at: string
}

export interface AdminRow {
  id: number
  admin_no: string
  username: string
  status: string
  created_at: string
}

export function listUsers(limit = 20): Promise<UserRow[]> {
  return request<UserRow[]>(`/users?limit=${limit}`)
}

export function listAdmins(limit = 20): Promise<AdminRow[]> {
  return request<AdminRow[]>(`/admins?limit=${limit}`)
}
```

- [ ] **Step 3: Replace the dashboard placeholder with a real page**

Replace the entire content of `admin/src/views/DashboardPlaceholderPage.vue` with:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDashboardStats, type DashboardStats } from '../api/dashboard'

const stats = ref<DashboardStats | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    stats.value = await getDashboardStats()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="panel">
    <p class="tag">DASHBOARD</p>
    <h2>管理后台</h2>

    <p v-if="loading">加载中...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <div v-else class="stats-grid">
      <div class="stat-card">
        <span class="stat-value">{{ stats?.total_orders ?? 0 }}</span>
        <span class="stat-label">总订单</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.pending_orders ?? 0 }}</span>
        <span class="stat-label">待支付</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.total_instances ?? 0 }}</span>
        <span class="stat-label">总实例</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.running_instances ?? 0 }}</span>
        <span class="stat-label">运行中</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.total_users ?? 0 }}</span>
        <span class="stat-label">总用户</span>
      </div>
      <div class="stat-card">
        <span class="stat-value">{{ stats?.pending_tasks ?? 0 }}</span>
        <span class="stat-label">待处理任务</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag {
  margin: 0 0 8px;
  color: #557257;
}

h2 { margin: 0 0 16px; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 16px;
}

.stat-card {
  display: flex;
  flex-direction: column;
  padding: 16px;
  border-radius: 12px;
  background: rgba(85, 114, 87, 0.08);
}

.stat-value {
  font-size: 1.8em;
  font-weight: 700;
}

.stat-label {
  font-size: 0.85em;
  color: #666;
}

.error { color: #c00; }
</style>
```

- [ ] **Step 4: Replace the user management placeholder with a real page**

Replace the entire content of `admin/src/views/UserManagePlaceholderPage.vue` with:

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listUsers, listAdmins, type UserRow, type AdminRow } from '../api/user'

const users = ref<UserRow[]>([])
const admins = ref<AdminRow[]>([])
const loading = ref(true)
const error = ref('')
const tab = ref<'users' | 'admins'>('users')

onMounted(async () => {
  try {
    const [u, a] = await Promise.all([listUsers(), listAdmins()])
    users.value = u
    admins.value = a
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="panel">
    <p class="tag">USERS</p>
    <h2>用户管理</h2>

    <div class="tabs">
      <button :class="{ active: tab === 'users' }" @click="tab = 'users'">注册用户</button>
      <button :class="{ active: tab === 'admins' }" @click="tab = 'admins'">管理员</button>
    </div>

    <p v-if="loading">加载中...</p>
    <p v-else-if="error" class="error">{{ error }}</p>

    <template v-else-if="tab === 'users'">
      <p v-if="users.length === 0">暂无用户</p>
      <table v-else>
        <thead><tr><th>ID</th><th>编号</th><th>手机</th><th>邮箱</th><th>状态</th><th>注册时间</th></tr></thead>
        <tbody>
          <tr v-for="u in users" :key="u.id">
            <td>{{ u.id }}</td><td>{{ u.user_no }}</td><td>{{ u.phone }}</td><td>{{ u.email || '-' }}</td><td>{{ u.status }}</td><td>{{ u.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </template>

    <template v-else>
      <p v-if="admins.length === 0">暂无管理员</p>
      <table v-else>
        <thead><tr><th>ID</th><th>编号</th><th>用户名</th><th>状态</th><th>创建时间</th></tr></thead>
        <tbody>
          <tr v-for="a in admins" :key="a.id">
            <td>{{ a.id }}</td><td>{{ a.admin_no }}</td><td>{{ a.username }}</td><td>{{ a.status }}</td><td>{{ a.created_at }}</td>
          </tr>
        </tbody>
      </table>
    </template>
  </section>
</template>

<style scoped>
.panel {
  padding: 24px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.9);
  color: #132224;
}

.tag { margin: 0 0 8px; color: #557257; }
h2 { margin: 0 0 16px; }

.tabs { margin-bottom: 16px; }
.tabs button {
  padding: 6px 16px;
  margin-right: 8px;
  border: 1px solid #557257;
  border-radius: 6px;
  background: transparent;
  color: #557257;
  cursor: pointer;
}
.tabs button.active {
  background: #557257;
  color: #fff;
}

table { width: 100%; border-collapse: collapse; }
th, td { padding: 8px 12px; text-align: left; border-bottom: 1px solid rgba(0,0,0,0.06); }
th { background: rgba(85, 114, 87, 0.08); }

.error { color: #c00; }
</style>
```

- [ ] **Step 5: Run tests and build**

Run: `bun --cwd admin test`
Expected: PASS

Run: `bun --cwd admin run build`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add admin/src/api/dashboard.ts admin/src/api/user.ts admin/src/views/DashboardPlaceholderPage.vue admin/src/views/UserManagePlaceholderPage.vue
git commit -m "feat: wire admin dashboard and user management to real APIs"
```

---

### Task 9: Full verification and build matrix

**Files:**
- None

- [ ] **Step 1: Run backend build**

Run: `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`
Expected: PASS

- [ ] **Step 2: Run web test and build**

Run: `bun --cwd web test && bun --cwd web run build`
Expected: PASS

- [ ] **Step 3: Run admin test and build**

Run: `bun --cwd admin test && bun --cwd admin run build`
Expected: PASS

---

## Self-Review Checklist

- Spec coverage:
  - notices table + backend CRUD: Tasks 1-4
  - notification persistence (was stub, now real): Task 3
  - admin users list endpoint: Task 5
  - admin admins list endpoint: Task 5
  - admin dashboard stats endpoint: Task 6
  - web notices page: Task 7
  - admin dashboard page: Task 8
  - admin users page: Task 8
  - full build matrix: Task 9

- Placeholder scan:
  - no TBD markers
  - no "implement later" or "handle appropriately"
  - all steps have complete code

- Type consistency:
  - `Notification` struct defined in Task 1, used in Tasks 2-4
  - `UserRow` / `AdminRow` defined in Task 5, used in Task 8
  - `DashboardStats` defined in Task 6, used in Task 8
  - `NoticeService` interface in handler matches service methods
  - Frontend API types match backend JSON tags

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-23-pvecloud-missing-backend-gaps.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**

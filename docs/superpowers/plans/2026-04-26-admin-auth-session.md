# Admin Auth Session Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete admin login session handling with server-side session validation, logout, token refresh, startup auth restore, audit logging, and basic login rate limiting.

**Architecture:** The Go API stores every admin JWT session in `admin_sessions`, uses the JWT `jti` as `session_id`, and reloads current admin roles and permissions from the database for protected requests. The admin frontend restores auth state through `/admin-api/auth/me`, logs out through `/admin-api/auth/logout`, and treats refresh failures as expired login.

**Tech Stack:** Go, Gin, GORM, MariaDB, JWT v5, Vue 3, Pinia, Axios, Vite, Bun.

---

### Task 1: Backend Session Foundation

**Files:**
- Modify: `server/internal/models/admin.go`
- Modify: `server/internal/pkg/jwt/jwt.go`
- Modify: `server/internal/pkg/errors/errors.go`
- Modify: `server/internal/dto/admin/auth_dto.go`
- Modify: `server/internal/dto/admin/dashboard_dto.go`

- [ ] Add `AdminSession` model mapped to `admin_sessions`.
- [ ] Add `AdminAuditLog` model mapped to `admin_audit_logs`.
- [ ] Add `ErrTooManyRequests` with HTTP 429 and code `42901`.
- [ ] Ensure JWT signing and parsing preserve `RegisteredClaims.ID`.
- [ ] Add `SessionSummary` and auth-state DTOs shared by login, me, refresh, and dashboard.

### Task 2: Backend Auth Service and Middleware

**Files:**
- Modify: `server/internal/services/admin_auth_service.go`
- Modify: `server/internal/services/admin_dashboard_service.go`
- Modify: `server/internal/middleware/admin_auth.go`
- Modify: `server/internal/api/admin/auth_handler.go`
- Modify: `server/internal/routes/admin_routes.go`

- [ ] Create sessions during successful login, using `jti=session_id`.
- [ ] Write login success/failure/logout/refresh audit records.
- [ ] Add basic login failure throttling by `IP + username/email`.
- [ ] Add `Me`, `Logout`, and `Refresh` service methods.
- [ ] Update admin auth middleware to validate token, session status, admin status, and current DB RBAC.
- [ ] Register `GET /auth/me`, `POST /auth/logout`, and `POST /auth/refresh`.

### Task 3: Admin Frontend Auth Flow

**Files:**
- Modify: `admin/src/types/auth.ts`
- Modify: `admin/src/types/dashboard.ts`
- Modify: `admin/src/api/auth.ts`
- Modify: `admin/src/api/http.ts`
- Modify: `admin/src/stores/auth.ts`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/src/components/AdminLayout.vue`

- [ ] Add session and auth-state TypeScript types.
- [ ] Add `getCurrentAdmin`, `logoutAdmin`, and `refreshAdminToken` API wrappers.
- [ ] Store session summary and restore auth state through `/auth/me`.
- [ ] Make route guard wait for auth restoration when a local token exists.
- [ ] Make logout call the backend and always clear local state.
- [ ] Keep login-page errors local and protected-route 401 handling centralized.

### Task 4: Verification and Project Log

**Files:**
- Update if present: project update log file.

- [ ] Run `gofmt -w` on changed Go files.
- [ ] Run `go test ./...` under `server`.
- [ ] Run `bun run build` under `admin`.
- [ ] If an update-log file exists, add a user-facing entry for safer admin login sessions.

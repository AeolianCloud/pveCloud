# PVE Cloud Frontend Real API Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace in-scope mock frontend business data in `web/` and `admin/` with the real backend APIs that already exist.

**Architecture:** Keep `web` and `admin` as isolated Vue 3 apps. Add a small per-app API layer, keep auth state in Pinia, enforce route guards for authenticated pages, and update views to render loading/error/empty states from live HTTP requests.

**Tech Stack:** Bun, Vue 3, Vue Router, Pinia, Vitest, native `fetch`

---

## Planned File Structure

- Create: `web/src/api/auth.ts`
- Create: `web/src/api/catalog.ts`
- Create: `web/src/api/order.ts`
- Create: `web/src/api/payment.ts`
- Create: `web/src/api/instance.ts`
- Modify: `web/src/lib/http.ts`
- Modify: `web/src/stores/auth.ts`
- Modify: `web/src/router/index.ts`
- Modify: `web/src/views/LoginView.vue`
- Modify: `web/src/views/RegisterView.vue`
- Modify: `web/src/views/ProductListView.vue`
- Modify: `web/src/views/ProductDetailView.vue`
- Modify: `web/src/views/OrderListView.vue`
- Modify: `web/src/views/PaymentStatusView.vue`
- Modify: `web/src/views/InstanceListView.vue`
- Modify: `web/src/views/InstanceDetailView.vue`
- Modify: `web/src/views/NoticeListView.vue`
- Modify: `web/src/views/LoginView.test.ts`
- Create: `admin/src/api/auth.ts`
- Create: `admin/src/api/catalog.ts`
- Create: `admin/src/api/order.ts`
- Create: `admin/src/api/instance.ts`
- Create: `admin/src/api/task.ts`
- Modify: `admin/src/lib/http.ts`
- Modify: `admin/src/stores/auth.ts`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/src/views/LoginView.vue`
- Modify: `admin/src/views/ProductManageView.vue`
- Modify: `admin/src/views/OrderManageView.vue`
- Modify: `admin/src/views/InstanceManageView.vue`
- Modify: `admin/src/views/TaskManageView.vue`
- Modify: `admin/src/views/DashboardView.vue`
- Modify: `admin/src/views/UserManageView.vue`
- Modify: `admin/src/views/DashboardView.test.ts`
- Modify: `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md`

### Task 1: Wire `web` HTTP, auth, and route guards

**Files:**
- Create: `web/src/api/auth.ts`
- Modify: `web/src/lib/http.ts`
- Modify: `web/src/stores/auth.ts`
- Modify: `web/src/router/index.ts`
- Modify: `web/src/views/LoginView.vue`
- Modify: `web/src/views/RegisterView.vue`
- Test: `web/src/views/LoginView.test.ts`

- [x] Add token-aware request helpers and normalized API response handling
- [x] Move login/register calls into `web/src/api/auth.ts`
- [x] Persist web auth token in store-backed state and support logout on `401`
- [x] Add route guard for authenticated user pages
- [x] Update login/register views to submit real forms and render request errors
- [x] Update login view test to cover the real submit flow contract
- [x] Run: `bun --cwd web test`

### Task 2: Replace `web` mock product/order/payment/instance flows

**Files:**
- Create: `web/src/api/catalog.ts`
- Create: `web/src/api/order.ts`
- Create: `web/src/api/payment.ts`
- Create: `web/src/api/instance.ts`
- Modify: `web/src/views/ProductListView.vue`
- Modify: `web/src/views/ProductDetailView.vue`
- Modify: `web/src/views/OrderListView.vue`
- Modify: `web/src/views/PaymentStatusView.vue`
- Modify: `web/src/views/InstanceListView.vue`
- Modify: `web/src/views/InstanceDetailView.vue`
- Modify: `web/src/views/NoticeListView.vue`

- [x] Add typed API wrappers for products, orders, payments, and instances
- [x] Replace static product list/detail data with live fetches
- [x] Add order creation from product detail and redirect to payment status
- [x] Replace static order list and payment status with live responses
- [x] Replace static instance list/detail with live responses
- [x] Convert notices page into an explicit out-of-scope placeholder
- [x] Run: `bun --cwd web run build`

### Task 3: Wire `admin` HTTP, auth, route guards, and real list pages

**Files:**
- Create: `admin/src/api/auth.ts`
- Create: `admin/src/api/catalog.ts`
- Create: `admin/src/api/order.ts`
- Create: `admin/src/api/instance.ts`
- Create: `admin/src/api/task.ts`
- Modify: `admin/src/lib/http.ts`
- Modify: `admin/src/stores/auth.ts`
- Modify: `admin/src/router/index.ts`
- Modify: `admin/src/views/LoginView.vue`
- Modify: `admin/src/views/ProductManageView.vue`
- Modify: `admin/src/views/OrderManageView.vue`
- Modify: `admin/src/views/InstanceManageView.vue`
- Modify: `admin/src/views/TaskManageView.vue`
- Modify: `admin/src/views/DashboardView.vue`
- Modify: `admin/src/views/UserManageView.vue`
- Test: `admin/src/views/DashboardView.test.ts`

- [x] Add token-aware request helpers and auth handling for admin requests
- [x] Move admin login into `admin/src/api/auth.ts`
- [x] Guard all admin routes except `/login`
- [x] Replace product/order/instance/task mock lists with live data pages
- [x] Convert dashboard and user management into explicit placeholders
- [x] Update the admin test to match the new dashboard placeholder behavior
- [x] Run: `bun --cwd admin test`

### Task 4: Verify frontend closure and sync docs

**Files:**
- Modify: `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md`

- [x] Run: `bun --cwd web run build`
- [x] Run: `bun --cwd admin run build`
- [x] Review `docs/superpowers` for frontend task wording that still implies mock-only flows
- [x] Update the MVP foundation plan to reflect real API integration progress

## Self-Review Checklist

- Spec coverage:
  - web/auth/products/orders/payments/instances: covered by Tasks 1-2
  - admin/auth/products/orders/instances/tasks: covered by Task 3
  - placeholder cleanup for out-of-scope pages: covered by Tasks 2-3
  - test/build verification and docs sync: covered by Task 4
- Placeholder scan:
  - no TBD markers
  - no "handle appropriately" style steps without explicit scope
- Type consistency:
  - each app owns its own API types and request helpers
  - bearer token handling stays inside each app's HTTP/auth layer

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-04-23-pvecloud-frontend-real-api-integration-implementation-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**

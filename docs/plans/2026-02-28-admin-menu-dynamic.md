# Admin Dynamic Menus Implementation Plan

> **For Codex/Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Admin 侧边栏菜单由后端动态下发；菜单结构落库并提供“菜单管理”CRUD，仅 `super_admin` 可访问完整菜单树与管理能力。

**Architecture:** 菜单以 `admin_menus` 单表存储树结构（`parent_id` + `sort`）；后端提供 `/menus/my`（按当前用户权限裁剪）与 `/menus` 管理接口（仅超管）；前端完全按下发渲染侧边栏，并提供菜单管理页维护数据库内容。

**Tech Stack:** Go 1.23+ / Gin / GORM / MySQL；Vue3 + TS + Vite；Naive UI；Pinia。

---

### Task 1: Database Schema + Seed

**Files:**
- Modify: `backend/sql/init.sql`
- Modify: `backend/sql/seed.sql`
- Create: `backend/sql/upgrade_20260228.sql`

**Steps:**
1. 新增表 `admin_menus`：字段包含 `parent_id/title/path/permission/icon/sort/visible` + 软删除（`deleted_at`）。
2. `seed.sql` 写入一套默认菜单：控制台、系统管理(目录)、管理员账号、角色管理、登录日志、操作日志、菜单管理(仅超管可见，permission 为空但 API 受超管保护)。
3. 升级脚本 `upgrade_20260228.sql`：为历史库补建 `admin_menus`（不破坏已有表）。

**Manual Test:**
- 执行 `init.sql + seed.sql` 后检查 `admin_menus` 是否有数据、层级是否正确、sort 是否生效。

---

### Task 2: Backend Menu Module

**Files:**
- Create: `backend/internal/model/admin_menu.go`
- Create: `backend/internal/service/menu/menu.go`
- Create: `backend/internal/handler/menu/menu.go`
- Modify: `backend/internal/router/router.go`
- Create: `backend/internal/middleware/super_admin.go`
- Modify: `backend/internal/middleware/oplog.go`

**Steps:**
1. Model：`AdminMenu` 嵌入 `model.Model`，定义 GORM 字段与 TableName。
2. Middleware：`RequireSuperAdmin(db)`，仅当用户拥有 `super_admin` 角色放行，否则 `403`。
3. Service：
   - `ListTreeAll()`：返回完整菜单树（排序）。
   - `ListTreeForUser(userID)`：读取用户权限集合 + 菜单表，按 `permission` 过滤并裁剪空目录。
   - `Create/Update/Delete`：基本 CRUD 校验（title/path 合法性、parent_id 存在性等）。
4. Handler：
   - `GET /api/v1/menus/my`：返回当前用户菜单树（登录即可）。
   - `GET /api/v1/menus`：完整菜单树（RequireSuperAdmin）。
   - `POST/PUT/DELETE /api/v1/menus`：管理接口（RequireSuperAdmin + WriteOpLog）。
5. OpLog：补 `module == "menu"` 的 target_label 预查询/提取（用 title）。

**Manual Test:**
- 超管：能访问 `/menus`、能增删改、操作日志能记录 title。
- 普通管理员：访问 `/menus` 返回 403；访问 `/menus/my` 返回裁剪后菜单。

---

### Task 3: Frontend Admin Dynamic Sidebar

**Files:**
- Create: `frontend-admin/src/api/menu.ts`
- Modify: `frontend-admin/src/layouts/DefaultLayout.vue`
- Modify: `frontend-admin/src/router/index.ts`

**Steps:**
1. 新增 `menu` API：`getMyMenus()`、`listMenus()`、`createMenu()`、`updateMenu()`、`deleteMenu()`。
2. DefaultLayout：
   - 登录后拉取 `/menus/my`，渲染 Naive UI `n-menu` 的 options。
   - 点击菜单项按 `path` 跳转；当前路由高亮按 `path` 匹配。
   - 目录节点无 path，不可跳转，仅展开。

**Manual Test:**
- 不同权限账号登录后，侧边栏菜单自动变化；刷新后仍正确。

---

### Task 4: Frontend Menu Management Page (Super Admin Only)

**Files:**
- Create: `frontend-admin/src/views/system/MenusView.vue`
- Modify: `frontend-admin/src/router/index.ts`

**Steps:**
1. 新增路由 `/system/menus`，meta 标记 `superAdmin: true`（前端守卫拦截非超管跳 403）。
2. 新增“菜单管理”页面：
   - Tree Table/层级展示（parent/children）。
   - 新增/编辑：title、path、permission、icon、sort、visible、parent_id。
   - 删除确认（软删）。
3. DefaultLayout：在系统管理菜单中，仅当 `authStore.isSuperAdmin` 为 true 才展示“菜单管理”入口（菜单项本身来自后端 `/menus/my`，但这里也保证“误配置/空库”时前端不会对非超管显示入口）。

**Manual Test:**
- 超管能打开菜单管理页；普通管理员路由被拦截 403。

---

## Acceptance Checklist (Manual)

1. 数据库
   - 执行 `backend/sql/init.sql`（新库）或 `backend/sql/upgrade_20260228.sql`（老库）后，存在 `admin_menus` 表。
   - 执行 `backend/sql/seed.sql` 后，`admin_menus` 至少包含：控制台、系统管理目录、管理员账号、角色管理、登录日志、操作日志、菜单管理（仅超管）。

2. 后端接口（用 Postman/curl 均可）
   - 任意已登录管理员：`GET /api/v1/menus/my` 返回 `code=0`，且结构为树形数组（含 `children`）。
   - 非超管：`GET /api/v1/menus` 返回 HTTP 403。
   - 超管：`GET /api/v1/menus` 返回 HTTP 200 且 `code=0`；`POST/PUT/DELETE /api/v1/menus` 可用。

3. 前端侧边栏
   - 登录后侧边栏不再由前端写死，完全由 `/menus/my` 渲染。
   - 普通管理员看不到“菜单管理”；超管能看到并可进入。

4. 菜单管理页
   - 超管在“菜单管理”可新建目录节点（path 留空）与页面节点（path 以 `/` 开头）。
   - 修改 `visible=0` 的菜单不会出现在 `/menus/my` 的下发结果中（侧边栏同步消失）。
   - 修改 `super_admin_only=1` 的菜单普通管理员不可见，超管可见。

5. 操作日志
   - 超管对菜单做 `create/update/delete` 后，在操作日志中能看到 module=`menu`，target_label 为菜单 title。

# Frontend Implementation Guardrails

This file is for AI implementation rules. Admin and web product/page design lives in `docs/admin/architecture.md` and `docs/web/architecture.md`; endpoint contracts live in `docs/server/api/` and matching business docs.

## Required Docs

Read these before frontend work:

- `docs/admin/architecture.md` for `admin/`
- `docs/web/architecture.md` for `web/`
- `docs/server/api/` when backend calls are involved
- `docs/server/api/conventions.md` for envelope, auth, and error semantics

## Shared Rules

- `admin/` and `web/` are fully independent.
- Do not create a shared frontend package.
- Do not import pages, components, requests, stores, types, constants, or utilities across `admin/` and `web/`.
- Each frontend owns its own `src/api/`, `src/stores/`, `src/types/`, `src/constants/`, `src/utils/`, and `src/router/`.
- Use the existing stack documented in the owning frontend architecture doc.

## Style Organization

- Keep global styles small: design tokens, CSS variables, reset/base elements, shell layout, and utilities reused by multiple pages.
- Put page-specific and component-specific styles in the owning Vue SFC with `<style scoped>`.
- Do not add page-only class blocks to `src/style.css`.
- If a style must be shared by multiple pages in the same frontend, extract it inside that frontend only; do not create a cross-app `shared/` style package.
- Use semantic CSS variables for theme-sensitive values. Scoped styles may define page-local variables on their root class.

## Third-Party UI Foundation

- Prefer mature third-party libraries for established frontend behavior that is easy to get subtly wrong: dialog focus trapping, select/listbox keyboard interaction, checkbox groups, popovers, tabs, date inputs, virtual scrolling, tables, charts, and accessibility primitives.
- For `admin/`, use PrimeVue as the preferred base component layer when introducing a UI library. Use PrimeVue components directly by default; wrap them under `admin/src/components/` only when pveCloud needs a stable business-facing API, a repeated component composition, or project-specific defaults.
- Use PrimeVue Styled Mode with an official preset theme as the primary visual base for frontend controls. Project CSS should provide pveCloud-specific layout, branding, page composition, and small overrides rather than rebuilding the same component visuals from scratch.
- PrimeFlex may be used as an optional admin layout utility layer for flex, grid, spacing, alignment, and responsive helpers. It must not replace PrimeVue Styled Mode or recreate PrimeVue component states.
- For protected `admin` pages, migrate UI to the PrimeVue suite progressively, excluding `LoginPage` unless the maintainer explicitly scopes login redesign. Dashboard, admin users, roles, sessions, system settings, audit logs, risk logs, and 403 should use PrimeVue components and admin-owned wrappers instead of spreading hand-written control styles.
- A third-party CSS utility or theme library may be introduced inside the owning frontend only. Do not create a cross-app shared package.
- Do not hand-roll complex component internals when a proven library already provides the behavior, unless the maintainer explicitly asks for a custom implementation.

## Admin Rules

- `admin` calls only `/admin-api/*`.
- Backend RBAC remains authoritative; frontend permission checks are only usability and navigation control.
- Route guards, request wrappers, auth stores, permission stores, menu constants, and layout behavior must match `docs/admin/architecture.md` and backend API docs.
- 前端页面、请求、类型和 store 文件必须按业务领域命名，避免用 `logs`、`common`、`utils`、`helper` 这类泛名承载具体业务；审计域统一使用 `audit` 命名，高危操作日志属于审计域。

## Web Rules

- `web` calls only `/api/*`.
- User auth, order, payment, instance, and ticket flows must match `docs/web/architecture.md` and backend API docs.

## Change Gate

Before implementing frontend pages, routes, API wrappers, types, constants, stores, or utility functions:

1. If the change is pure UI/UX polish, skip the document-first confirmation gate and implement directly in the owning frontend after reading this guardrail.
2. Update the owning frontend doc when page behavior, workflow, state semantics, routes, permissions, request wrappers, or durable product structure changes.
3. Do not write style-only layout, color, typography, spacing, icon, or density decisions into `docs/` unless the maintainer explicitly asks for a design spec.
4. Update `docs/server/api/` when backend calls change.
5. Stop for maintainer confirmation only after docs/contracts are changed.
6. Implement only in the owning frontend directory.

## Verification Baseline

```powershell
cd admin
bun run build
```

Use the equivalent command under `web/` once that frontend exists.

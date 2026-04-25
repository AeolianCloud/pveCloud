# pveCloud Frontend Rules

This reference consolidates `admin/` and `web/` frontend rules.

## Shared Frontend Decisions

- Runtime/package manager: Bun.
- Framework: Vue 3.
- Build tool: Vite.
- Language: TypeScript.
- `admin/` and `web/` are fully independent.
- Do not create a shared frontend package.
- Do not import pages, components, requests, stores, types, constants, or utilities across `admin/` and `web/`.
- Each frontend owns its own:
  - `src/api/`
  - `src/stores/`
  - `src/types/`
  - `src/constants/`
  - `src/utils/`
  - `src/router/`

## Admin Frontend

Scope: platform operation, support, and administrator workflows.

API boundary:

```text
admin -> /admin-api/*
```

Never call `/api/*` from `admin`.

Current stack:

| Area | Choice |
| --- | --- |
| Package/script runner | Bun |
| Build | Vite |
| UI framework | Vue 3 composition API |
| Language | TypeScript |
| Router | Vue Router |
| State | Pinia |
| HTTP | Axios |
| Icons | lucide-vue-next |

Expected pages:

- Login
- Dashboard
- Users
- Products/plans/regions/nodes/images/prices
- Orders
- Payments/wallet flows/manual credit
- Instances
- Tickets
- Admin users/roles/permissions
- System settings and audit logs

Admin state:

- `auth`: admin token and admin profile.
- `permission`: permission codes and menu visibility.
- `layout`: sidebar, theme, collapse state.
- `tabs`: optional admin multi-tabs.

Admin login:

```text
POST /admin-api/auth/login
```

Successful login returns admin JWT, admin summary, role IDs, and permission codes. Frontend stores `access_token` in auth store and `localStorage`, then sends:

```text
Authorization: Bearer <access_token>
```

Frontend may use `permission_codes` for page/button display, but backend RBAC remains authoritative.

Local admin dev server:

```powershell
cd admin
bun install
bun dev
```

The admin Vite server uses port `5174` and proxies `/admin-api/*` to `http://127.0.0.1:8080`.

## Web Frontend

Scope: public site and user center.

API boundary:

```text
web -> /api/*
```

Never call `/admin-api/*` from `web`.

Expected pages:

- Home
- Products
- Pricing
- Login/register
- Order creation
- Payment result
- User center
- My orders
- My instances
- Tickets

Web state:

- `auth`: user token and user profile.
- `cart`: optional order config draft.
- `order`: current order state.
- `instance`: optional instance-list cache.

## Frontend Change Gate

Before implementing frontend pages, routes, API wrappers, types, constants, stores, or utility functions:

1. Update this reference or the remaining lightweight docs.
2. Confirm API contract in `docs/server/api/openapi.yaml` when backend calls are involved.
3. Stop for maintainer confirmation.
4. Implement only in the owning frontend directory.

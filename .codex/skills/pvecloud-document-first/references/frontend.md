# Frontend Implementation Guardrails

This file is for AI implementation rules. Admin and web product/page design lives in `docs/admin/architecture.md` and `docs/web/architecture.md`; endpoint contracts live in `docs/server/api/openapi.yaml`.

## Required Docs

Read these before frontend work:

- `docs/admin/architecture.md` for `admin/`
- `docs/web/architecture.md` for `web/`
- `docs/server/api/openapi-src/` and generated `docs/server/api/openapi.yaml` when backend calls are involved
- `docs/server/api/conventions.md` for envelope, auth, and error semantics

## Shared Rules

- `admin/` and `web/` are fully independent.
- Do not create a shared frontend package.
- Do not import pages, components, requests, stores, types, constants, or utilities across `admin/` and `web/`.
- Each frontend owns its own `src/api/`, `src/stores/`, `src/types/`, `src/constants/`, `src/utils/`, and `src/router/`.
- Use the existing stack documented in the owning frontend architecture doc.

## Admin Rules

- `admin` calls only `/admin-api/*`.
- Backend RBAC remains authoritative; frontend permission checks are only usability and navigation control.
- Route guards, request wrappers, auth stores, permission stores, menu constants, and layout behavior must match `docs/admin/architecture.md` and OpenAPI.

## Web Rules

- `web` calls only `/api/*`.
- User auth, order, payment, instance, and ticket flows must match `docs/web/architecture.md` and OpenAPI.

## Change Gate

Before implementing frontend pages, routes, API wrappers, types, constants, stores, or utility functions:

1. Update the owning frontend doc when page/workflow/state changes.
2. Update `docs/server/api/openapi-src/` and run `node ./scripts/generate-openapi.mjs` when backend calls change.
3. Stop for maintainer confirmation.
4. Implement only in the owning frontend directory.

## Verification Baseline

```powershell
cd admin
bun run build
```

Use the equivalent command under `web/` once that frontend exists.

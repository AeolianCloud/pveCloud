# Backend Implementation Guardrails

This file is for AI implementation rules. Backend architecture, API conventions, jobs, business status lists, and integration design live in `docs/server/`; endpoint contracts live in `docs/server/api/openapi.yaml`.

## Required Docs

Read these before backend work:

- `docs/server/README.md`
- `docs/server/architecture.md`
- `docs/server/go-technical.md`
- `docs/server/api/conventions.md`
- `docs/server/api/openapi.yaml` when routes, handlers, DTOs, auth, response shapes, or status codes change
- `docs/server/jobs.md` when async task behavior changes
- `docs/server/integrations/README.md` when external systems are involved
- `docs/server/database/design.md` and `server/migrations/` when persistence changes

## Implementation Rules

- Keep the backend a Go monolith and follow the documented `server/` directory shape.
- Keep `/api/*` user routes and `/admin-api/*` admin routes separated.
- Keep business logic in services/jobs. External clients only adapt protocols.
- Use shared models under `server/internal/models/`; do not split models by frontend.
- Use `server/internal/pkg` only for stable infrastructure such as response, errors, JWT, password, pagination, validator, and logger.
- API handlers must use unified `pkg/errors` and `pkg/response`.
- Handler comments must stay aligned with OpenAPI.
- Startup OpenAPI validation remains the safety net when enabled.

## Auth and Permission

- Follow `docs/server/api/conventions.md` for auth and response semantics.
- Follow OpenAPI for protected endpoint declarations.
- Admin permission checks belong in admin permission middleware.
- Handlers declare required permission codes; they do not hand-roll permission logic.
- Backend RBAC is authoritative even if the frontend hides menu items.

## Jobs and Integrations

- API process creates jobs; worker executes jobs.
- Do not hold long database transactions around external PVE/payment/notify calls.
- Persist local recovery anchors before external resource operations.
- Idempotency and compensation rules must be documented before implementation.

## Verification Baseline

Use the narrowest meaningful set for the change, usually:

```powershell
cd server
gofmt -w .
go test ./...
```

For API changes, also ensure the OpenAPI loader test passes.

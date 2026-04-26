# Backend Implementation Guardrails

This file is for AI implementation rules. Backend architecture, API conventions, jobs, business status lists, and integration design live in `docs/server/`; endpoint contracts live in `docs/server/api/` and matching business docs.

## Required Docs

Read these before backend work:

- `docs/server/README.md`
- `docs/server/architecture.md`
- `docs/server/go-technical.md`
- `docs/server/api/conventions.md`
- `docs/server/api/` and matching backend docs when routes, handlers, DTOs, auth, response shapes, or status codes change
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
- Handler comments must stay aligned with `docs/server/api/`.
- 管理端普通操作写入 `admin_audit_logs`；高危操作必须同时写入普通审计日志和 `admin_risk_logs`，不要只写其中一边。
- 高危操作判断、风险等级和脱敏字段以 `docs/server/database/design.md`、`docs/server/api/` 和迁移 SQL 为准，技能只提醒双写规则。

## Go Naming and Code Rules

- Go code must follow idiomatic Go naming: exported identifiers use PascalCase, unexported identifiers use camelCase, and initialisms stay uppercase consistently, for example `ID`, `IP`, `URL`, `API`, `JWT`, not `Id`, `Ip`, `Url`.
- File names use lowercase snake_case and include business domain plus responsibility, for example `admin_audit_service.go`, `admin_user_handler.go`, `system_config_dto.go`.
- Backend files, types, services, handlers, DTOs, and helpers must use explicit business-domain names. Avoid vague buckets such as `log`, `manager`, `common`, `helper`, `utils`, `data`, or `base` for business code.
- If a capability is truly shared infrastructure, it belongs under `server/internal/pkg/` with a stable infrastructure name such as `response`, `validator`, `password`, `jwt`, `pagination`, or `logger`. Do not move business rules into `pkg`.
- Current project packages are broad (`services`, `api/admin`, `dto/admin`), so exported types must include domain and responsibility, for example `AdminAuditService`, `AdminUserHandler`, `SystemConfigRequest`. Do not use generic exported names such as `Service`, `Handler`, or `Manager` inside broad packages.
- Interfaces are introduced only when there is a real boundary or test seam. Do not create Java-style `IUserService` interfaces. Name interfaces by behavior when useful, for example `Reader`, `Writer`, `Validator`, or a precise domain behavior.
- Constants and variables use Go style, not screaming snake case. Use names like `adminStatusActive`, `adminSessionStatusRevoked`, or exported `ErrUnauthorized` where appropriate.
- Exported functions, types, interfaces, and package variables must keep the project block-comment style, with the first paragraph in Chinese explaining purpose. The first sentence should start with the identifier when practical.
- Keep handlers thin: bind and validate request, call service, return unified response. Business rules belong in services/jobs, not handlers or middleware.
- Return errors upward and convert them to unified API responses at the boundary. Do not use `panic` for business errors.
- Pass `context.Context` from request entry into database, Redis, jobs, and external calls.
- Keep database transactions short. Do not call external PVE/payment/notify systems inside long transactions.
- Logs, errors, API messages, CLI prompts, and operator-facing output use Chinese unless a third-party protocol requires fixed English.

## Auth and Permission

- Follow `docs/server/api/conventions.md` for auth and response semantics.
- Keep protected endpoint declarations aligned between route middleware, handler comments, and `docs/server/api/`.
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

For API changes, ensure route tests or focused handler/service tests cover the contract change where practical.

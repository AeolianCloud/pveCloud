# pveCloud Backend Rules

This reference consolidates backend architecture, Go technical rules, API conventions, jobs, and integration boundaries.

## Architecture

- Go monolith; do not introduce microservices or complex DDD.
- Keep user API and admin API separated:
  - `/api/*` for user/web.
  - `/admin-api/*` for admin.
- Core business service rules are shared to avoid duplicated order/payment/instance logic.
- Admin-specific actions use `admin_xxx_service.go`.
- Models are shared under `server/internal/models/`; do not split models by frontend.
- External protocol clients live under `server/internal/integrations/`.
- Long-running work lives under `server/internal/jobs/` and is executed by `cmd/worker`.

## Backend Directory Shape

```text
server/
├─ cmd/
│  ├─ api/
│  ├─ worker/
│  └─ setup-admin/
├─ internal/
│  ├─ bootstrap/
│  ├─ routes/
│  ├─ openapi/
│  ├─ api/
│  │  ├─ web/
│  │  └─ admin/
│  ├─ middleware/
│  ├─ services/
│  ├─ models/
│  ├─ dto/
│  │  ├─ web/
│  │  └─ admin/
│  ├─ jobs/
│  ├─ integrations/
│  └─ pkg/
├─ migrations/
├─ storage/logs/
├─ config.example.yaml
├─ go.mod
└─ go.sum
```

`internal/pkg` is only for stable backend infrastructure such as response, errors, JWT, password, pagination, validator, and logger. Do not move business rules into `pkg`.

## Go Technical Stack

| Area | Choice |
| --- | --- |
| Go | 1.26.2 |
| HTTP | Gin |
| ORM | GORM |
| DB driver | `gorm.io/driver/mysql` |
| Config | YAML |
| Dev reload | Air |
| Logging | standard `log/slog` JSON logs |
| OpenAPI | `github.com/getkin/kin-openapi` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Password | bcrypt from `golang.org/x/crypto/bcrypt` |
| Validation | `github.com/go-playground/validator/v10` |
| Tests | `github.com/stretchr/testify` |
| Jobs | `cmd/worker` + `async_tasks` |

Avoid runtime hot reload, global repository layer, event bus, microservice frameworks, codegen frameworks, and Kubernetes directory structure.

## Config

- Real config path defaults to `server/config.yaml`; keep it ignored.
- API and worker support `-config config.yaml`.
- Config groups: `app`, `database`, `jwt`, `worker`, `openapi`, `log`; later may add `pve`, `payment`, `mail`, and `sms`.
- Runtime config hot update is not supported; restart after config changes.

## API Rules

- API contract source: `docs/server/api/openapi.yaml`.
- API process exposes `GET /openapi.yaml` when enabled.
- Startup should validate OpenAPI when enabled.
- Initialization endpoints:
  - `GET /healthz`
  - `GET /api/ping`
  - `GET /admin-api/ping`
- `/healthz` is lightweight and returns non-2xx when database ping fails.
- All business responses use:

```json
{"code":0,"message":"成功","data":{}}
```

- Error response:

```json
{"code":40001,"message":"参数错误","data":null}
```

- Error code ranges:
  - `0`: success
  - `400xx`: parameter/validation
  - `401xx`: unauthenticated/token invalid/token expired
  - `403xx`: forbidden
  - `404xx`: not found
  - `409xx`: conflict/duplicate submission
  - `500xx`: internal
  - `600xx`: payment
  - `700xx`: PVE/instance
  - `800xx`: admin operation

Handlers must use unified `pkg/errors` and `pkg/response`.

## Auth and Permission

- User JWT claims: `user_id`, `token_type=user`, issued/expiry fields.
- Admin JWT claims: `admin_id`, `token_type=admin`, `role_ids`, `permission_codes`, issued/expiry fields.
- User and admin JWT secrets/issuers should be different.
- Admin RBAC:

```text
admin_users -> admin_user_roles -> admin_roles -> admin_role_permissions -> admin_permissions
```

- Permission checks live in admin permission middleware.
- Handlers declare required permission codes; they do not hand-roll permission logic.

## Core Business Statuses

Orders:

```text
pending, paid, provisioning, active, cancelled, expired, failed, refunded
```

Payments:

```text
created, pending, success, failed, closed, refunding, refunded
```

Instances:

```text
creating, running, stopped, suspended, expired, deleting, deleted, error
```

Async tasks:

```text
pending, running, success, failed, cancelled
```

Tickets:

```text
open, pending_admin, pending_user, closed
```

## Payment and Order Rules

- Order amount is calculated by backend service, never trusted from frontend.
- First phase is one order to one instance; do not add quantity to `orders`.
- Image selection must be filtered by region and configured PVE template mapping.
- Payment success must branch by `payment_scene` and `order_type`:
  - `payment_scene=order`, `order_type=new`: mark order `paid`, create one unique instance creation task.
  - `payment_scene=order`, `order_type=renew`: mark order `paid`, extend expiry or create renew sync task.
  - `payment_scene=topup`: write wallet transaction and increase balance; do not create instance task.
- Payment callback, manual credit, refund, instance provisioning, and instance deletion must be idempotent.

## Instance Provisioning Rules

Before calling PVE, persist local recovery anchors:

```text
instances.vmid
instances.provisioning_key
instances.pve_task_upid
```

Do not hold a DB transaction while making long external calls. Use local state first, then job execution, then compensation/retry for external failures.

## Worker and Jobs

- API process only creates tasks.
- Worker process pulls executable tasks from `async_tasks`.
- Task types:
  - `instance_create`
  - `instance_renew`
  - `order_expire`
  - `payment_check`
  - `instance_status_sync`
- Pull `pending` tasks with `run_at <= now()`.
- Lock with `locked_by` and `locked_until`.
- Re-check business state before execution.
- Failures update `last_error`, increment `retry_count`, and calculate next `run_at`.
- Mark `failed` after max retries and surface to admin.

## Integration Boundaries

External clients only adapt protocols; services/jobs own business rules.

PVE client lives in `server/internal/integrations/pve/` and can handle login, token/cookie, nodes, storage, templates, VM status, create/delete/reinstall, power actions, and task/UPID queries. It must not decide order eligibility, instance ownership, or delivery completion.

Payment client lives in `server/internal/integrations/payment/` and can create/query/close payment orders, validate callbacks, and request/query refunds. Payment orchestration belongs in `payment_service.go`; callbacks must be idempotent.

Notify adapters live in `server/internal/integrations/notify/`; first phase may reserve email/SMS without real providers.

## Local Backend Commands

```powershell
cd server
Copy-Item config.example.yaml config.yaml
go mod tidy
gofmt -w .
go test ./...
go run ./cmd/api -config config.yaml
go run ./cmd/worker -config config.yaml
air -c .air.toml
go run ./cmd/setup-admin -config config.yaml -username admin -email admin@example.com -password "change_me_password"
```

## Verification Baseline

Backend initialization is acceptable when:

- `go mod tidy` succeeds.
- `gofmt -w .` has no unexpected differences.
- `go test ./...` succeeds.
- API can start.
- `/healthz`, `/api/ping`, `/admin-api/ping`, and `/openapi.yaml` work.
- Worker can start and stop.

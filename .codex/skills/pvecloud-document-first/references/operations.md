# pveCloud Local Development and Operations

This reference consolidates local setup and deployment/operations boundaries.

## Local Dependencies

- Go 1.26.2
- Bun
- MariaDB 11.4.9
- Redis optional in the first phase

## Database

Target database:

```text
pvecloud
```

Initialization SQL:

```text
server/migrations/001_init.sql
```

Use local MariaDB credentials; do not commit real credentials.

## Backend Config

- Example config: `server/config.example.yaml`
- Real config: `server/config.yaml`
- `server/config.yaml` stays ignored.
- Config is YAML based; `.env` is not the main source.

Suggested groups:

```text
app
database
redis
jwt
worker
openapi
pve
payment
mail
sms
log
```

## Local Startup Order

1. MariaDB.
2. Redis if needed.
3. API process.
4. Worker process.
5. `admin` and later `web` Vite dev servers.

Health and diagnostics:

```text
GET http://localhost:8080/healthz
GET http://localhost:8080/openapi.yaml
GET http://localhost:8080/api/ping
GET http://localhost:8080/admin-api/ping
```

## Deployment Components

```text
api      user and admin HTTP API
worker   async task process
web      public site and user center
admin    management UI
MariaDB  business source of truth
Redis    cache/session/queue enhancement, optional early
```

## Deployment Boundaries

- Proxy `/api/*` to the Go API user routes.
- Proxy `/admin-api/*` to the Go API admin routes.
- `web` and `admin` can use separate domains or the same domain with different paths.
- `worker` exposes no public HTTP business endpoint.

## PVE Operations

- PVE is external; MariaDB remains the business source of truth.
- PVE operations must be orchestrated by backend services and worker jobs.
- PVE HTTP success is not instance delivery success; query task/UPID.
- Remote success plus local failure must be recoverable through `vmid`, `provisioning_key`, and `pve_task_upid`.

## Admin Operations and Audit

- High-risk admin operations must write `admin_audit_logs`.
- Audit should cover operator, object, before/after state, IP/user-agent when available, and remark.

## Backups

First phase backups should include:

- MariaDB full/incremental backups.
- Secure backup of `config.yaml` and deployment config.
- PVE node/template configuration records.

Recovery drills should cover orders, payments, wallet balances, instances, async tasks, and audit logs.

# pveCloud

PVE Cloud is a cloud sales platform for product catalog, order, payment, provisioning, and instance operations. This repository currently contains the backend baseline and runtime foundation for the first delivery phase.

## Architecture

- `server/`: Go modular monolith with `public-api`, `admin-api`, and `worker` entrypoints
- `MariaDB`: authoritative business storage
- `Redis`: cache, session, idempotency, and task-assist storage only
- `docs/`: architecture specs and execution plans

The frontend projects `web/` and `admin/` are intentionally not initialized yet. They will be added in later tasks after the backend baseline is stable.

## Repository Layout

```text
pveCloud/
  docs/
    superpowers/
      specs/
      plans/
  server/
    cmd/
      public-api/
      admin-api/
      worker/
    internal/
      bootstrap/
      common/
```

## Prerequisites

- Go 1.24+
- Docker Desktop with `docker compose`

## Quick Start

1. Copy environment variables:

```powershell
Copy-Item .env.example .env
```

2. Start MariaDB and Redis:

```powershell
docker compose up -d
```

3. Run backend tests:

```powershell
go -C server test ./...
```

4. Start the public API:

```powershell
$env:APP_ENV='local'
$env:PUBLIC_API_ADDR=':8080'
$env:ADMIN_API_ADDR=':8081'
$env:WORKER_ADDR=':8082'
$env:MYSQL_DSN='root:root@tcp(127.0.0.1:3306)/pvecloud?parseTime=true&loc=Local'
$env:REDIS_ADDR='127.0.0.1:6379'
$env:JWT_WEB_SECRET='change-me-web'
$env:JWT_ADMIN_SECRET='change-me-admin'
go -C server run ./cmd/public-api
```

Health check:

```powershell
Invoke-WebRequest http://127.0.0.1:8080/healthz
```

## Environment Variables

The backend baseline currently requires:

- `APP_ENV`
- `PUBLIC_API_ADDR`
- `ADMIN_API_ADDR`
- `WORKER_ADDR`
- `MYSQL_DSN`
- `REDIS_ADDR`
- `JWT_WEB_SECRET`
- `JWT_ADMIN_SECRET`

See `.env.example` for default local values.

## Development Commands

- Run all backend tests: `go -C server test ./...`
- Build all backend entrypoints: `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`
- Check compose file: `docker compose config`

## Current Scope

Completed in this baseline:

- repository bootstrap
- backend configuration model
- backend runtime factories
- shared logger, MariaDB, and Redis client factories
- common HTTP response and error helpers

Planned next:

- MariaDB migrations and schema baseline
- auth modules
- catalog, order, payment, and task center
- frontend projects `web/` and `admin`

## Documentation

- Design baseline: `docs/superpowers/specs/2026-04-21-pvecloud-sales-platform-design.md`
- Implementation plan: `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md`

# pveCloud

PVE Cloud is a cloud sales platform covering catalog, order, payment, provisioning, instance operations, and the minimum user/admin frontend flows for the MVP phase.

## Architecture

- `server/`: Go modular monolith with `public-api`, `admin-api`, and `worker` entrypoints
- `MariaDB`: authoritative business storage
- `Redis`: cache, session, idempotency, and task-assist storage only
- `docs/`: architecture specs, plans, and ADRs
- `web/`: Bun + Vue 3 user-side SPA
- `admin/`: Bun + Vue 3 admin-side SPA

## Repository Layout

```text
pveCloud/
  docs/
    adr/
    superpowers/
      specs/
      plans/
  admin/
  server/
    cmd/
      public-api/
      admin-api/
      worker/
    config/
      config.yaml
    internal/
      bootstrap/
      common/
  web/
```

## Prerequisites

- Go 1.24+
- Bun 1.3+
- Docker Desktop with `docker compose`

## Quick Start

1. Start MariaDB and Redis:

```powershell
docker compose up -d
```

2. Review backend config:

File: `server/config/config.yaml`

3. Run backend tests:

```powershell
go -C server test ./...
```

4. Run frontend verification when working on SPA slices:

```powershell
bun --cwd web run test
bun --cwd web run build
bun --cwd admin run test
bun --cwd admin run build
```

5. Start the public API:

```powershell
go -C server run ./cmd/public-api
```

Health check:

```powershell
Invoke-WebRequest http://127.0.0.1:8080/healthz
```

## Configuration

The backend baseline uses YAML only.

- config file path: `server/config/config.yaml`
- no `.env`
- no environment variable override

Current config covers:

- app environment
- public/admin/worker listen addresses
- MariaDB DSN
- Redis address
- JWT secrets

## Development Commands

- Run all backend tests: `go -C server test ./...`
- Build all backend entrypoints: `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`
- Run web tests: `bun --cwd web run test`
- Build web: `bun --cwd web run build`
- Run admin tests: `bun --cwd admin run test`
- Build admin: `bun --cwd admin run build`

## Current Scope

Completed in the current MVP slice:

- repository bootstrap
- YAML-based backend configuration
- backend runtime factories
- shared logger, MariaDB, and Redis client factories
- common HTTP response and error helpers
- schema baseline, auth, catalog, order, payment, task center, resource adapter, and instance flow skeletons
- minimum `web` and `admin` SPA shells with route-level views

## Documentation

- Design baseline: `docs/superpowers/specs/2026-04-21-pvecloud-sales-platform-design.md`
- Implementation plan: `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md`
- ADR 001: `docs/adr/001-task-source-of-truth.md`
- ADR 002: `docs/adr/002-capacity-reservation.md`

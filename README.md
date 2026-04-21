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
    config/
      config.yaml
    internal/
      bootstrap/
      common/
```

## Prerequisites

- Go 1.24+
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

4. Start the public API:

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

## Current Scope

Completed in this baseline:

- repository bootstrap
- YAML-based backend configuration
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

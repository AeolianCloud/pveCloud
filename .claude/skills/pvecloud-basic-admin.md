---
name: pvecloud-basic-admin
description: Use when working on the basic admin foundation. This skill helps AI stay inside the current backend/admin scope.
---

# pveCloud Basic Admin

## Purpose

This skill is a stage helper.
It keeps AI aligned with the current basic-admin scope without turning the skill into a second product spec.

Project truth still lives in:

- `docs/admin/architecture.md`
- `docs/server/api/`
- `docs/server/database/design.md`
- `server/migrations/`
- `docs/analysis/basic-admin-gap.md`
- `docs/plan/basic-admin-foundation.md`
- `docs/progress/`

## Current Scope

- Backend scope still covers: admin auth, dashboard, RBAC, admin sessions, system configs, audit logs, risk logs.
- Current admin frontend scope is narrower: only `Login`, `Dashboard`, and `403`.
- Removed admin pages must not be recreated unless the owning docs are updated first and the maintainer confirms reopening them.

## Working Rules

- Keep backend and frontend scope separate in your head.
- Do not confuse "backend capability still exists" with "frontend page should still exist".
- Audit is a single domain. Risk logs are part of audit.
- Backend RBAC is the final authority; frontend permission logic is usability-only.
- Keep the admin frontend aligned with the current narrowed surface unless docs explicitly reopen it.

## Verification Baseline

Backend work:

```powershell
cd server
go test ./...
```

Admin frontend work:

```powershell
cd admin
bun run build
```

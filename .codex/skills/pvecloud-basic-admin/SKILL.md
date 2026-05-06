---
name: pvecloud-basic-admin
description: "Use only when the task explicitly mentions the historical basic-admin foundation, basic-admin plan/progress docs, or drift between old stage notes and the current admin surface. This helper forces current page scope back to owner docs instead of historical phase notes."
---

# pveCloud Basic Admin

## Purpose

This skill is a stage helper.
It keeps AI aligned with the historical basic-admin scope without turning the skill into a second product spec.

Project truth still lives in:

- `docs/admin/architecture.md`
- `docs/admin/pages/README.md`
- `docs/admin/routing-permissions.md`
- `docs/server/api/`
- `docs/server/database/design.md`
- `server/migrations/`
- `docs/analysis/basic-admin-gap.md`
- `docs/plan/basic-admin-foundation.md`
- `docs/progress/`

## Historical Scope

- Backend scope during that stage covered: admin auth, dashboard, RBAC, admin sessions, system configs, audit logs, risk logs.
- Admin frontend scope changed during and after that stage. Do not treat any historical page list in this skill or progress notes as the current contract.
- Current admin page scope must come from `docs/admin/pages/README.md`, with route and permission semantics from `docs/admin/routing-permissions.md`.
- Removed or not-yet-opened admin pages must not be recreated unless the owning docs are updated first and the maintainer confirms reopening them.

## When This Skill Applies

Use this helper only when the task explicitly touches the historical basic-admin stage, including:

- basic-admin history around admin auth, dashboard, RBAC, admin sessions, system configs, audit logs, or the historical risk-log scope
- removed admin pages, reopened admin menus, or confusion between backend capability, historical phase notes, and current frontend page scope
- basic-admin gap, plan, or progress documents

Do not load this helper for unrelated current-stage work or ordinary admin implementation tasks. Historical progress can explain why a scope changed, but current contracts still live in the owner docs and machine contracts.

When this helper applies, read:

1. `AGENTS.md`
2. `.codex/skills/pvecloud-document-first/SKILL.md`
3. `docs/progress/MASTER.md`
4. `docs/analysis/basic-admin-gap.md`
5. `docs/plan/basic-admin-foundation.md`
6. `docs/admin/pages/README.md` when page scope matters
7. `docs/admin/routing-permissions.md` when route or permission scope matters
8. The relevant server or admin frontend architecture docs

## Working Rules

- Keep backend and frontend scope separate in your head.
- Do not confuse "backend capability still exists" with "frontend page should still exist".
- Audit is a single domain. Risk logs are part of audit.
- Backend RBAC is the final authority; frontend permission logic is usability-only.
- Keep the admin frontend aligned with the current documented surface unless docs explicitly reopen it.

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

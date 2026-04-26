---
name: pvecloud-basic-admin
description: Guide pveCloud basic admin foundation work. Use when Codex works on Dashboard, admin users, roles and permissions, admin sessions, system settings, audit logs, risk logs, admin menus, admin routes, admin RBAC, or remaining basic backend/admin frontend tasks before product, order, payment, instance, and ticket modules are opened.
---

# pveCloud Basic Admin

## Overview

Use this project-local skill to keep the current stage focused on a complete basic admin backend and admin frontend.

This skill does not replace project documents. Contracts still live in `docs/server/api/`, database facts live in `docs/server/database/` and `server/migrations/`, and admin frontend behavior lives in `docs/admin/architecture.md`.

## Start Every Session

- Read `AGENTS.md`.
- Read `.codex/skills/pvecloud-document-first/SKILL.md`.
- Read `docs/progress/MASTER.md`.
- Read `docs/analysis/basic-admin-gap.md`.
- Read `docs/plan/basic-admin-foundation.md`.
- Read the phase file linked from the active status in `docs/progress/MASTER.md`.

If the requested work changes API, database schema, permissions, admin routes, menus, page behavior, state semantics, configuration, or business process, update the matching document or migration first and stop for maintainer confirmation.

## Stage Boundary

In scope: Login, Dashboard, admin users, roles, permissions, admin sessions, system settings, audit logs, risk logs, and 403 access denied.

Out of scope: product plans, products, orders, payments, wallets, instances, PVE operations, tickets, and user-facing business flows.

Do not add unfinished business modules back to admin menus, protected routes, or fallback menus.

## Implementation Order

Finish the active phase in `docs/progress/MASTER.md` before starting a later phase unless the user explicitly changes priority.

Preferred order:

- Phase 1: centralized audit writing, masking, and high-risk action matrix.
- Phase 2: admin users, roles, and permissions.
- Phase 3: admin sessions and system settings.
- Phase 4: Dashboard metrics and admin UX closure.
- Phase 5: acceptance, tests, and release readiness.

## Audit Rules

Ordinary admin operations write `admin_audit_logs`.

High-risk admin operations write both `admin_audit_logs` and `admin_risk_logs`.

High-risk operations include resetting admin passwords, disabling admins, changing admin roles, creating or disabling roles, changing role permissions, revoking another admin session, changing sensitive config, login failure lockout, and captcha rate limiting.

Only `audit:view` can view main log information. `audit:sensitive_view` is required for masked `before_data`, `after_data`, and `user_agent`. Passwords, tokens, secrets, captcha values, and sensitive config plaintext must never be returned.

## Naming Rules

Use explicit business-domain names. Avoid `log`, `manager`, `common`, `helper`, `utils`, `data`, and `base` for concrete business code.

Use `audit` for the audit domain. Risk logs belong to the audit domain. Do not introduce vague names such as `admin_log_service`.

Go exported identifiers use PascalCase, unexported identifiers use camelCase, and initialisms stay uppercase, for example `ID`, `IP`, `URL`, `API`, `JWT`.

Go filenames use lowercase snake_case with domain plus responsibility, for example `admin_audit_service.go`, `admin_user_handler.go`, and `system_config_dto.go`.

## Progress Updates

After completing a task, update the matching `docs/progress/phase-*.md` checkbox, then update the count and Current Status in `docs/progress/MASTER.md`.

Keep progress notes factual. Record blockers, document confirmations, and verification commands that actually ran.

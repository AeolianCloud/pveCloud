---
name: pvecloud-backend-development
description: Execute backend development tasks for the pveCloud repository. Use when working under `server/`, implementing or modifying backend modules, continuing tasks from the project implementation plan, adding MariaDB migrations, wiring backend runtime code, or enforcing pveCloud-specific backend rules and boundaries.
---

# PVECloud Backend Development

## Overview

Use this skill for backend work in this repository only. Treat the implementation plan as the default task source, enforce the pveCloud module boundaries, and keep backend changes aligned with the architecture spec and database conventions.

## Required Workflow

1. Read `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md` before changing backend code.
2. Find the current unfinished backend task and continue from there unless the user explicitly redirects the task order.
3. Inspect the existing files under `server/` that are relevant to the task before editing.
4. Implement the smallest coherent backend slice that completes the task.
5. Run the task-specific verification commands. At minimum, run targeted backend tests or builds that prove the change.
6. Update the plan document checkbox state for the completed backend steps.
7. Report whether the current work forms a clean commit boundary. Commit only if the user asks or the workflow for that session requires it.

Do not skip verification.
Do not skip plan status updates.
Do not jump ahead to a later backend task unless the user explicitly asks for it.

## Task Source Of Truth

Default backend task source:

- `docs/superpowers/plans/2026-04-21-pvecloud-mvp-foundation.md`

Related architecture reference:

- `docs/superpowers/specs/2026-04-21-pvecloud-sales-platform-design.md`

When the plan and repository state diverge:

1. Trust the checked state in the plan as the first signal.
2. Verify the actual repository contents and recent commits.
3. If the plan is stale, update the plan status as part of the work.
4. If the requested work changes task boundaries or architecture, stop and clarify before implementing.

## Backend Module Boundaries

Keep backend code inside these module responsibilities.

- `bootstrap`: application assembly, config loading, runtime wiring, server startup
- `common`: shared infrastructure only, such as HTTP response helpers, error types, logger, database, cache, and narrowly shared constants
- `auth`: token/session support, identity parsing, auth middleware
- `user`: public-side user registration, login, profile, security settings
- `adminuser`: admin accounts, admin login, roles, permissions, admin operation records
- `catalog`: products, SKUs, regions, saleability, capacity reservation, node bindings
- `order`: order creation, querying, order status transitions, renew or change order flows
- `payment`: payment orders, callback verification, payment status transitions, refunds
- `billing`: pricing, billing cycle, discount rules, billing snapshots, service period calculations
- `instance`: business-side instance records, instance status facts, operation eligibility, service period facts
- `resource`: adapter for the underlying VM platform and resource-facing operations only
- `task`: async task creation, claiming, retry, idempotency, execution tracking
- `notification`: SMS, email, site messages, notification delivery records
- `audit`: user audit, admin operation logs, key business event tracking

Do not place domain rules in `cmd/` or thin HTTP handlers.
Do not put order, billing, payment, or instance state transition logic in `common`.
Do not let `resource` absorb business ordering or billing logic.

## Database Rules

Apply these rules to every new MariaDB migration and backend schema change.

- Every table must have a Chinese table comment.
- Every field must have a Chinese field comment.
- Every status field must document all enum values in Chinese comments.
- Monetary fields use integer cents.
- Time fields use `DATETIME(3)`.
- Status fields prefer readable strings such as `VARCHAR(32)`.
- Table names use lowercase snake_case plural nouns.
- Primary keys use `id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT`.
- Business numbers use `*_no`.
- Relationship fields use `*_id`.

If a migration lacks explicit Chinese comments, treat it as incomplete.

## Configuration And Runtime Rules

- Backend configuration is YAML-only.
- Read backend config from `server/config/config.yaml`.
- Do not introduce `.env` or environment-variable override logic unless the user explicitly requests a config model change.
- Keep `public-api`, `admin-api`, and `worker` as separate entrypoints.
- Keep MariaDB as the source of truth for business data and async task truth.
- Redis is support infrastructure only, never the source of truth for orders, payments, instances, or tasks.

## Verification Rules

For backend tasks, prefer the narrowest command that proves the task, then run broader checks if the task touches shared infrastructure.

Examples:

- config or bootstrap changes: `go -C server test ./internal/bootstrap/...`
- migration changes: `go -C server test ./internal/common/database -v`
- broad backend slice changes: `go -C server test ./...`
- entrypoint or compile-safety changes: `go -C server build ./cmd/public-api ./cmd/admin-api ./cmd/worker`

If a verification command cannot run, state exactly why.

## Completion Checklist

Before closing a backend task, confirm all of the following:

- relevant backend code under `server/` is implemented
- required verification commands were run
- plan checkbox state is updated
- any README or backend-facing docs touched by the task are updated if needed
- commit readiness is stated clearly

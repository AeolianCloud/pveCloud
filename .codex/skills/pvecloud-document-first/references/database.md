# Database Implementation Guardrails

This file is for AI implementation rules. Database design facts live in `docs/server/database/design.md`; executable schema lives in `server/migrations/`.

## Required Docs

Read these before database or transaction work:

- `docs/server/database/design.md`
- `server/migrations/`
- `docs/server/architecture.md` when persistence affects business flow
- `docs/server/jobs.md` when async task persistence changes

## Implementation Rules

- Update `docs/server/database/design.md` when table groups, business constraints, transaction boundaries, or data ownership rules change.
- Update `server/migrations/` when table, column, index, seed, or constraint SQL changes.
- Do not treat skill references as schema contracts.
- Money fields, status fields, indexes, constraints, and transaction boundaries must follow the database design doc and migration SQL.
- Do not call external systems inside a long database transaction.
- For idempotent business flows, document the unique key or locking strategy before implementation.

## Verification Baseline

- Review generated SQL manually.
- Run focused backend tests.
- When migrations are changed, apply them to a disposable local database before reporting completion when feasible.

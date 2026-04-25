# pveCloud Workflow Rules

This reference is the compact project workflow previously spread across `AGENTS.md`, `docs/README.md`, `docs/process/document-first.md`, and `docs/ai/context.md`.

## Required Reading Order

1. `AGENTS.md`
2. `.codex/skills/pvecloud-document-first/SKILL.md`
3. Relevant reference files under `.codex/skills/pvecloud-document-first/references/`
4. Machine contracts when needed:
   - `docs/server/api/openapi.yaml`
   - `server/migrations/`
   - `server/config.example.yaml`

## Document-First Gate

For implementation, migration, API, page, config, deployment, operations, or business-process changes:

1. Run `git status --short`.
2. Read the relevant skill reference.
3. Update the relevant design/API/reference document first.
4. Stop and ask the maintainer to confirm the design or contract.
5. Implement only after explicit confirmation.
6. Verify with focused tests/builds.

Read-only investigation and pure explanation do not need the confirmation gate unless edits become necessary.

## Code and Documentation Ownership

- `server/` is covered by `references/backend.md` and `references/database.md`.
- `admin/` and `web/` are covered by `references/frontend.md`.
- local setup and deployment are covered by `references/operations.md`.
- OpenAPI remains in `docs/server/api/openapi.yaml` because it is a machine-readable API contract.
- SQL migrations remain in `server/migrations/` because they are executable database contracts.

## Hard Project Decisions

- Backend: Go monolith, no microservices, no complex DDD.
- Go version: 1.26.2.
- Backend config: YAML file, not `.env` as the main source.
- Frontend: Bun + Vue 3 + Vite + TypeScript.
- Database: MariaDB 11.4.9 + InnoDB.
- Cache: Redis can be reserved early, not required for the first closure.
- User API: `/api/*`.
- Admin API: `/admin-api/*`.
- Worker: independent `cmd/worker`; API process only creates jobs.
- Admin-only tables use the `admin_` prefix.
- Money fields use integer cents and `_cents` suffix.
- Status fields use string constants, not database enum.

## Collaboration Rules

- Never overwrite user changes. If dirty files are unrelated, leave them alone.
- If dirty files overlap with the task, work with them rather than reverting.
- Final responses after code changes should include verification results and useful follow-up directions.

## Code Style

- Public functions, types, interfaces, and package variables use block comment style: `/** ... */`.
- The first comment paragraph is Chinese and explains purpose.
- API handler comments can use `@route`, `@request`, `@response`, and `@auth`; keep them aligned with OpenAPI.
- Internal comments should be Chinese and explain non-obvious business rules, idempotency, transactions, permissions, compensation, concurrency, and external-system boundaries.
- Do not write empty comments that only restate the code.
- Logs, errors, API `message`, CLI prompts, and user/admin/operator-facing output use Chinese.
- If a third-party protocol requires fixed English output, return English and explain nearby in Chinese comments.

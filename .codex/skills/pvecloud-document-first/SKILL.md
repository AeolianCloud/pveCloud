---
name: pvecloud-document-first
description: Enforce the pveCloud repository workflow. Use whenever Codex is working inside the pveCloud repo on implementation, migration, API, database, frontend page, admin/web request wrapper, configuration, deployment, operations, or business-process changes. This skill keeps AI workflow and engineering rules separate from docs/API/schema contracts.
---

# pveCloud Document First

## Purpose

Use this project-local skill to enforce the pveCloud document-first workflow.

This skill is an AI working method, not the project documentation source of truth. Interface contracts, architecture facts, database design, frontend page design, deployment notes, and config examples belong in `docs/`, `server/migrations/`, and `server/config.example.yaml`.

## Source Boundaries

- AI workflow and implementation guardrails: `.codex/skills/pvecloud-document-first/`.
- Human-readable project documentation: `docs/`.
- API machine contract: `docs/server/api/openapi.yaml`.
- Database machine contract: `server/migrations/`.
- Config example contract: `server/config.example.yaml`.

If a skill reference conflicts with `docs/` or a machine contract, the document or machine contract wins. Fix the stale skill reference before implementing.

## Hard Workflow

1. Run `git status --short` before changing anything.
2. Read `AGENTS.md`.
3. Read `references/workflow.md`.
4. Read the relevant skill guardrail:
   - backend work: `references/backend.md`
   - database or transaction work: `references/database.md`
   - admin/web frontend work: `references/frontend.md`
   - setup/deployment/operations work: `references/operations.md`
5. Read the matching project docs or machine contracts:
   - API: `docs/server/api/openapi.yaml` and `docs/server/api/`
   - backend design: `docs/server/`
   - admin frontend: `docs/admin/`
   - web frontend: `docs/web/`
   - database: `docs/server/database/` and `server/migrations/`
   - operations: `docs/development/` and `docs/operations/`
   - config: `server/config.example.yaml`
6. If the task changes behavior, API, schema, page, workflow, config, or operations, update the matching docs or machine contract first.
7. After documentation or contract changes, stop. Summarize the proposed design/API/acceptance points and ask the maintainer to confirm.
8. Do not implement code, migrations, frontend pages, or config changes until the maintainer explicitly confirms.
9. After confirmation, implement narrowly according to the confirmed docs.
10. Verify with the smallest meaningful tests/builds and report results.

## Non-Negotiable Gates

- If the user asks to "start", "initialize", "build", "implement", "connect end to end", or similar, still update docs/contracts first and stop for confirmation before code.
- If OpenAPI changes are needed, update `docs/server/api/openapi.yaml`; do not encode endpoint fields only in skill references.
- If database changes are needed, update `docs/server/database/` when design changes and `server/migrations/` when schema changes.
- If frontend request wrappers, types, constants, stores, or utilities are involved, update the owning frontend docs and OpenAPI when backend calls are involved.
- Keep `admin/` and `web/` independent. Do not create a shared frontend package.
- Do not overwrite user changes. If dirty files are unrelated, leave them alone. If dirty files overlap with the task, work with them.
- If a task is pure read-only investigation or explanation, no confirmation gate is needed unless edits become necessary.

## Confirmation Prompt

Use a concise stop message after docs/contracts are updated. Write it in the user's language:

```text
文档/契约已先更新。确认点：
- ...

请确认这些设计/API/验收点是否通过。你确认后我再进入实现。
```

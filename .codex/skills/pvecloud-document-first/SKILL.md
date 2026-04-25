---
name: pvecloud-document-first
description: Enforce the pveCloud repository workflow. Use whenever Codex is working inside the pveCloud repo on implementation, migration, API, database, frontend page, admin/web request wrapper, configuration, deployment, operations, or business-process changes. This skill requires reading AGENTS.md and docs first, updating the matching design/API/domain documents before code, then stopping for maintainer confirmation before any implementation.
---

# pveCloud Document First

## Purpose

Use this project-local skill to make `AGENTS.md` operationally strict. It does not replace repository rules; it reinforces them.

## Hard Workflow

1. Run `git status --short` before changing anything.
2. Read `AGENTS.md`.
3. Read `references/workflow.md`.
4. Find and read the domain reference for the task:
   - workflow and cross-cutting rules: `references/workflow.md`
   - backend architecture, Go, API, jobs, integrations: `references/backend.md`
   - database and transaction rules: `references/database.md`
   - admin/web frontend rules: `references/frontend.md`
   - local setup, deployment, operations: `references/operations.md`
   - API machine contract: `docs/server/api/openapi.yaml`
   - database machine contract: `server/migrations/`
5. If the task changes behavior, API, schema, page, workflow, config, or operations, update the matching docs or references first.
6. After documentation/reference changes, stop. Summarize the proposed contract/design and ask the maintainer to confirm.
7. Do not implement code, migrations, frontend pages, or config changes until the maintainer explicitly confirms the docs.
8. After confirmation, implement narrowly according to the confirmed docs.
9. Verify with the smallest meaningful tests/builds and report results.

## Non-Negotiable Gates

- If the user asks to "start", "initialize", "build", "implement", "connect end to end", or similar, still update docs first and stop for confirmation before code.
- If OpenAPI changes are needed, `docs/server/api/openapi.yaml` remains the machine-readable source of truth.
- If database changes are needed, update `references/database.md` and `server/migrations/` before code.
- If frontend request wrappers, types, constants, stores, or utilities are involved, update `references/frontend.md`, keep `admin/` and `web/` independent, and do not create a shared frontend package.
- Do not overwrite user changes. If `git status` shows unrelated dirty files, leave them alone.
- If a task is pure read-only investigation or explanation, no confirmation gate is needed unless edits become necessary.

## Confirmation Prompt

Use a concise stop message after docs are updated. Write it in the user's language:

```text
Docs have been updated first. Confirmation points:
- ...

Please confirm whether these design/API/acceptance points are approved. I will implement only after your confirmation.
```

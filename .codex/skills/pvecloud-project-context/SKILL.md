---
name: pvecloud-project-context
description: Use when the task asks about pveCloud's current technology stack, feature scope, module boundaries, stage status, or where to find the authoritative docs. This skill is navigation-only and must not duplicate project contracts.
---

# pveCloud Project Context

## Purpose

This skill helps AI find current pveCloud project facts without copying those facts into the AI workflow layer.
It is navigation-only.

Do not write API fields, response structures, table schemas, config values, page contracts, feature lists, or technology versions here.
Those facts live in `docs/`, `server/migrations/`, and `server/config.example.yaml`.

Use this skill when the user asks about:

- current technology stack
- current feature scope or opened/closed modules
- admin/web/server/database/operations boundaries
- project stage status or progress
- which owner docs should be read for a task
- drift between remembered project context and current docs

If the user asks to change code or contracts, also use `$pvecloud-document-first`.

## Source Map

Read the smallest relevant set:

- Project overview and current status: `docs/README.md`, `docs/ai/context.md`
- Server architecture and feature boundaries: `docs/server/README.md`, `docs/server/architecture.md`
- Server technology stack: `docs/server/go-technical.md`
- API contracts: `docs/server/api/README.md` when present, then relevant files in `docs/server/api/`
- Database design and executable schema: `docs/server/database/design.md`, `server/migrations/`
- Runtime config: `server/config.example.yaml`, then `server/internal/platform/config/config.go`
- Admin frontend architecture and stack: `docs/admin/README.md`, `docs/admin/architecture.md`
- Admin pages and permissions: `docs/admin/pages/README.md`, relevant `docs/admin/pages/*.md`, `docs/admin/routing-permissions.md`
- Web frontend architecture and stack: `docs/web/architecture.md`, relevant `docs/web/pages/*.md`
- Security baseline: `docs/security.md`
- Local development and operations: `docs/development/local-setup.md`, `docs/operations/deployment.md`
- Stage/progress status: `docs/progress/README.md`, `docs/progress/MASTER.md`, then relevant handoff/progress files

## Reading Rules

- Prefer owner docs over progress docs for current facts.
- Prefer migrations over prose for table shape.
- Prefer `server/config.example.yaml` over prose for config shape.
- If docs conflict, use the authority order in `AGENTS.md`, then point out the drift.
- Do not infer current feature availability from code alone when owner docs say the feature is not open.
- Do not infer admin capability from web docs, or web capability from admin docs.
- Do not read the whole `docs/` tree by default; start from indexes and load only task-relevant files.

## Answering Rules

- Cite the exact docs read by path.
- Distinguish current contract from historical progress.
- When giving a technology stack or feature list, say it is sourced from the owner docs rather than restating it as skill knowledge.
- If facts are missing or stale, say which owner doc should be updated.
- If the question leads to implementation or contract changes, switch to the document-first workflow before editing.

## Drift Checklist

When the user asks "what is current", "where are we", "what stack", "what is open", or similar:

1. Read the relevant index document first.
2. Check whether `docs/progress/MASTER.md` says a stage is historical, current, planned, or confirmed.
3. Cross-check at least one owner doc for the same fact.
4. If code appears ahead of docs, report it as drift rather than treating code as contract.

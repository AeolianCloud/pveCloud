---
name: pvecloud-document-first
description: Enforce the pveCloud document-first workflow. Use whenever work may change API, schema, frontend behavior, permissions, config, deployment, or business process. This skill defines AI workflow only and never replaces project contracts in docs/, migrations, or config.example.yaml.
---

# pveCloud Document First

## Purpose

This skill defines how AI should work in this repository.
It does not define API fields, response payloads, table schemas, or page contracts.

Use this skill whenever the task touches implementation, refactoring, migration, frontend behavior, request wrappers, routes, permissions, config, deployment, or cross-module business rules.

## Source Boundaries

- AI workflow, guardrails, and implementation habits: `.codex/skills/`
- Human-readable project facts and contracts: `docs/`
- Executable database contract: `server/migrations/`
- Executable config example contract: `server/config.example.yaml`

If a skill reference conflicts with a project document or machine contract, the project document or machine contract wins.

## Required Reading Order

1. `AGENTS.md`
2. This `SKILL.md`
3. `references/workflow.md`
4. The relevant domain guardrail:
   - backend: `references/backend.md`
   - database: `references/database.md`
   - frontend: `references/frontend.md`
   - operations: `references/operations.md`
   - basic admin stage helper when needed: `.codex/skills/pvecloud-basic-admin/SKILL.md`
5. The matching project docs or contracts:
   - server/API: `docs/server/`, `docs/server/api/`
   - admin frontend: `docs/admin/`
   - web frontend: `docs/web/`
   - database: `docs/server/database/`, `server/migrations/`
   - development/operations: `docs/development/`, `docs/operations/`
   - config: `server/config.example.yaml`
   - plans/progress: `docs/analysis/`, `docs/plan/`, `docs/progress/`

## Document-First Decision

Before making changes, classify the task:

- Contract or behavior change:
  API, schema, route meaning, permission logic, request wrapper semantics, page workflow, state semantics, config shape, deployment behavior, business process.
- Pure UI/UX polish:
  layout, spacing, colors, typography, iconography, visual density, responsive presentation, or non-contract copy.

## Mandatory Workflow

1. Run `git status --short`.
2. Read the required files in the order above.
3. Decide whether the task is contract/behavior work or pure UI/UX polish.
4. For contract/behavior work, update the owning docs or machine contracts first.
5. After updating those docs/contracts, stop and ask the maintainer to confirm.
6. Do not implement contract/behavior code until the maintainer explicitly confirms.
7. For pure UI/UX polish, implement directly after reading the frontend guardrail.
8. Verify with the smallest meaningful tests or builds.
9. Report what changed, what was verified, and any residual risk.

## Non-Negotiable Rules

- Skills may point to docs, but they must not become a second API spec or schema document.
- Do not hide durable product behavior inside skills.
- Do not overwrite user changes.
- Do not recreate removed frontend pages, routes, or menus unless the docs are updated first and the maintainer confirms reopening them.
- Keep `admin/` and `web/` independent. No shared frontend runtime package.
- If the repository does not contain a `web/` app yet, treat `docs/web/` as planning and contract guidance, not proof of an existing implementation.

## Stop Message

When docs/contracts were updated for a contract/behavior change, stop with a short confirmation message in the user's language:

```text
文档/契约已先更新。确认点：
- ...

请确认这些设计、契约或验收口径是否通过。你确认后我再进入实现。
```

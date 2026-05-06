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

When you are about to work on a new feature or enhancement, you must name the exact docs/contracts you read before coding. If you cannot name them, you have not satisfied the reading requirement.

For any non-admin-only feature, first classify the affected surfaces before coding:

- `web`: user-facing routes, state, request wrappers, login flow, display behavior
- `admin`: operational pages, menus, permissions, and management workflows
- `server/api`: `/api/*` and/or `/admin-api/*` contracts, auth, validation, response shape
- `database`: migrations, indexes, constraints, transaction boundaries, durable state
- `operations`: config examples, local startup, deployment proxy boundaries, runtime dependencies

Default rule: if a feature is not clearly a pure `admin` visual/UI change, treat it as a coupled change until the owning docs prove otherwise.

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

When a docs directory contains a `README.md`, read that index before opening deeper files. Then load only the task-relevant page, API, schema, config, or progress documents instead of reading the whole domain.

## Context Budget

Keep context focused as the project grows:

- Prefer directory indexes and architecture summaries before detailed docs.
- Read page-level docs only for the page or module being changed.
- Read API docs only for endpoints touched by the task.
- Read progress docs only to understand stage history, current status, or drift.
- Summarize older context in your own notes instead of repeatedly re-reading unrelated docs.
- If the needed docs no longer fit comfortably in context, stop and ask to narrow scope or split the task.

## Document-First Decision

Before making changes, classify the task:

- Contract or behavior change:
  API, schema, route meaning, permission logic, request wrapper semantics, page workflow, state semantics, config shape, deployment behavior, business process.
- Pure UI/UX polish:
  layout, spacing, colors, typography, iconography, visual density, responsive presentation, or non-contract copy.

Use this short classification note before editing files for non-trivial work:

```text
任务分类：
- 类型：契约/行为变更 | 纯 UI/UX | AI 工作流/协作规则
- 影响面：admin | web | server/api | database | operations | ai-workflow
- 已读 owner docs/contracts：
- 是否需要先停确认：
```

For coupled business features, do not implement one surface in isolation:

- user-visible feature: default `web + server/api + database`
- operational feature: default `admin + server/api + database`
- shared business fact: default `web + admin + server/api + database`
- anything that changes config, startup, proxying, or dependency order: add `operations`

Only these may proceed as single-surface work by default:

- pure visual/UI polish with no contract impact
- wording/layout adjustments that do not change flow or state
- already-confirmed scope within one existing surface

Do not start implementing after only drafting a plan from memory. For any contract or behavior change, first explicitly identify the owning contract files that were read and the owner docs or machine contracts that must change. If the required owner docs have not been read yet, stop reading and do not edit implementation files.

If the implementation plan changes while coding, or if you add a security hardening, validation rule, transaction rule, config default, permission check, route behavior, state transition, or storage behavior that was not in the already confirmed contract, treat it as a new contract/behavior change. Update the owning docs or machine contracts first, stop, and ask the maintainer to confirm before continuing implementation.

## Drift-First Check

Document-first also applies when code already moved ahead of docs.
Before continuing an unfinished task, after `git pull`, or when the working tree already contains user/code changes, do a drift check before writing more implementation:

- Inspect the changed or newly pulled files with `git status`, `git diff`, `git log`, or targeted file reads.
- Map the touched code area to the owning docs or machine contracts listed in the required reading order.
- Compare current code behavior against those docs, plus related plan/progress docs when the task affects stage scope.
- Treat stale, conflicting, or old-location docs as a contract/behavior problem, not as harmless cleanup.
- Do not continue feature implementation until the owning docs/contracts are brought back in line and the maintainer confirms the corrected scope.

If multiple docs disagree, do not silently pick the one that matches the code. Use the authority order from `AGENTS.md`: API docs, domain docs, frontend docs, migrations, and config examples are the contract sources; plan/progress docs must then be synchronized to that contract.

## Progress Docs Rule

`docs/progress/` is a stage ledger, not the final product contract.

- Use progress docs to understand what was done, why scope changed, and what was accepted.
- Do not use progress docs as the deciding source when they conflict with API, domain, frontend, migration, or config contracts.
- When a feature or stage is completed, make sure the durable facts are reflected in the owning contract docs first.
- If old progress docs no longer describe the current contract, update, archive, or mark them as historical before continuing implementation.
- When reading archived or historical progress docs, extract context only; re-check current contracts before coding.

## Commit Message Rule

When the maintainer asks AI to commit, the commit message must be written in Chinese by default and remain useful for review from another machine.

- Do not run `git add`, `git commit`, `git push`, or any other Git history/staging mutation unless the maintainer explicitly asks for that Git action in the current conversation.

- Use Chinese for the subject and body by default, unless the maintainer explicitly requests another language.
- Use a concise Chinese subject, but do not rely on the subject alone.
- Include a detailed Chinese body for non-trivial changes.
- Explain in Chinese why the change was needed, not only what files changed.
- Group the body by meaningful areas such as docs, workflow, frontend, backend, database, verification, and risk.
- Mention verification commands that were run, or explicitly say when no runtime verification was needed.
- Mention notable residual risks or follow-up constraints.
- Do not amend a commit that already matches `origin/*` unless the maintainer explicitly asks to rewrite published history.
- If the previous published commit message was too terse, create a new corrective commit that adds the missing workflow rule instead of rewriting remote history.

## Mandatory Workflow

1. Run `git status --short`.
2. Read the required files in the order above.
3. State the owner docs or machine contracts that were read for the task before editing implementation files.
4. If code was pulled, already changed, or the user says work is unfinished, run the drift-first check.
5. Decide whether the task is contract/behavior work or pure UI/UX polish.
6. For contract/behavior work, update the owning docs or machine contracts first.
7. After updating those docs/contracts, stop and ask the maintainer to confirm.
8. Do not implement contract/behavior code until the maintainer explicitly confirms.
9. While implementing, if new behavior or hardening is discovered or added, return to step 6 before continuing.
10. For pure UI/UX polish, implement directly after reading the frontend guardrail.
11. Before final response, compare implemented behavior against the updated owner docs/contracts and fix drift.
12. Verify with the smallest meaningful tests or builds.
    If `AGENTS.md` or `.codex/skills/**/SKILL.md` changed, run `python3 scripts/validate-skills.py`.
13. Report what changed, what was verified, and any residual risk.

## Non-Negotiable Rules

- Skills may point to docs, but they must not become a second API spec or schema document.
- Do not hide durable product behavior inside skills.
- Do not overwrite user changes.
- Do not mutate the Git staging area, local commits, or remote branches unless the maintainer explicitly asks for that Git action.
- Do not assume docs are current just because they exist; validate them against the current code path before continuing an unfinished feature.
- Do not implement a new feature, API, schema, permission, config, storage behavior, validation rule, security hardening, or transaction rule before reading the owning docs/contracts and updating them when needed.
- Do not treat security hardening as "implementation-only" if it changes accepted inputs, rejected inputs, storage semantics, audit behavior, transactional guarantees, permissions, or runtime configuration.
- Do not finish a task until implementation and owning docs/contracts have been compared for drift.
- Do not treat progress/plan docs as the only source of truth when they conflict with owning contract docs or implementation.
- Do not let completed phase notes keep driving new code after the durable contract docs have moved on.
- Do not add more code on top of a known code/docs mismatch unless the maintainer explicitly asks for an emergency implementation path.
- Do not recreate removed frontend pages, routes, or menus unless the docs are updated first and the maintainer confirms reopening them.
- Keep `admin/` and `web/` independent. No shared frontend runtime package.
- If the repository does not contain a `web/` app yet, treat `docs/web/` as planning and contract guidance, not proof of an existing implementation.
- Do not start a new business feature by implementing only one surface and planning to backfill the rest later.
- Do not treat `docs/web/` as evidence that the `web/` app already exists.
- Do not assume `node`, `bun`, or `go` are absent until checking an interactive shell with `bash -ic`; local version managers may only initialize there.

## Stop Message

When docs/contracts were updated for a contract/behavior change, stop with a short confirmation message in the user's language:

```text
文档/契约已先更新。确认点：
- ...

请确认这些设计、契约或验收口径是否通过。你确认后我再进入实现。
```

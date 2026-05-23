---
name: pvecloud-document-first
description: Enforce the pveCloud document-first workflow. Use whenever work may change API, schema, frontend behavior, permissions, config, deployment, business process, or AI collaboration workflow/skills. This skill defines AI workflow only and never replaces project contracts in docs/, migrations, or config.example.yaml.
---

# pveCloud Document First

## Purpose

This skill is the AI workflow entry for `pveCloud`.
It defines how AI should read, classify, pause, implement, and verify work through the document-first gates.
It must not define API fields, response payloads, table schemas, config values, page contracts, or durable product behavior.

Use this skill whenever the task touches implementation, refactoring, migration, frontend behavior, request wrappers, routes, permissions, config, deployment, cross-module business rules, or AI collaboration workflow files such as `CLAUDE.md` and `.claude/skills/`.

Use companion skills only when they match the user's actual task:

- `$pvecloud-project-context` for read-only questions about current stack, feature scope, module boundaries, stage status, or owner-doc locations.
- `$pvecloud-systematic-debugging` for concrete bugs, failing tests, broken builds, runtime errors, or unexpected behavior.
- `$pvecloud-skill-quality` for creating, editing, or reviewing pveCloud AI workflow skills.
- `$pvecloud-basic-admin` only for historical basic-admin scope or drift between old stage notes and the current admin surface.
- `$pvecloud-contract-quality` when writing or reviewing owner docs or machine contracts for state changes, external side effects, async jobs, permissions, config, schema, API, security, or cross-surface business behavior.

## Source Boundaries

- AI workflow and guardrails: `.claude/skills/`
- Human-readable project facts and contracts: `docs/`
- Executable database contract: `server/migrations/`
- Executable config example contract: `server/config.example.yaml`

If a skill reference conflicts with a project document or machine contract, the project document or machine contract wins.

For AI workflow-only work, read `.claude/skills/pvecloud-workflow.md`, the skill files being changed, and any project docs needed to check for stale or misleading project facts. Do not read unrelated product docs just to edit workflow guardrails.

## Classification

Before non-trivial edits, state:

```text
任务分类：
- 类型：契约/行为变更 | 纯 UI/UX | AI 工作流/协作规则
- 影响面：admin | web | server/api | database | operations | ai-workflow
- 已读 owner docs/contracts：
- 是否需要先停确认：
```

For any non-admin-only feature, first classify the affected surfaces before coding:

- `web`: user-facing routes, state, request wrappers, login flow, display behavior
- `admin`: operational pages, menus, permissions, and management workflows
- `server/api`: `/api/*` and/or `/admin-api/*` contracts, auth, validation, response shape
- `database`: migrations, indexes, constraints, transaction boundaries, durable state
- `operations`: config examples, local startup, deployment proxy boundaries, runtime dependencies

Default rule: if a feature is not clearly a pure `admin` visual/UI change, treat it as a coupled change until the owning docs prove otherwise.

Only these may proceed as single-surface work by default:

- pure visual/UI polish with no contract impact
- wording/layout adjustments that do not change flow or state
- already-confirmed scope within one existing surface

## Required Reading Order

1. `CLAUDE.md`
2. This skill
3. `.claude/skills/pvecloud-workflow.md`
4. `.claude/skills/pvecloud-contract-quality.md` when the task touches state changes, external side effects, async jobs, permissions, config, schema, API, security, or cross-surface business behavior
5. The task-relevant guardrail:
   - backend: `.claude/skills/pvecloud-backend.md`
   - database: `.claude/skills/pvecloud-database.md`
   - frontend: `.claude/skills/pvecloud-frontend.md`
   - operations: `.claude/skills/pvecloud-operations.md`
   - historical basic admin scope: `.claude/skills/pvecloud-basic-admin.md`
   - concrete failures: `.claude/skills/pvecloud-systematic-debugging.md`
   - AI workflow quality: `.claude/skills/pvecloud-skill-quality.md`
6. The matching project docs or contracts when implementation, behavior, or project facts are touched:
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
- AI workflow or collaboration rule change:
  `CLAUDE.md`, `.claude/skills/`, tool preferences, reading order, execution gates, final response habits, or skill quality rules.

For coupled business features, do not implement one surface in isolation:

- user-visible feature: default `web + server/api + database`
- operational feature: default `admin + server/api + database`
- shared business fact: default `web + admin + server/api + database`
- anything that changes config, startup, proxying, or dependency order: add `operations`

Only these may proceed as single-surface work by default:

- pure visual/UI polish with no contract impact
- wording/layout adjustments that do not change flow or state
- already-confirmed scope within one existing surface

## Drift-First Check

Document-first also applies when code already moved ahead of docs.
Before continuing an unfinished task, after `git pull`, or when the working tree already contains user/code changes, do a drift check before writing more implementation:

- Inspect the changed or newly pulled files with `git status`, `git diff`, `git log`, or targeted file reads.
- Map the touched code area to the owning docs or machine contracts listed in the required reading order.
- Compare current code behavior against those docs, plus related plan/progress docs when the task affects stage scope.
- Treat stale, conflicting, or old-location docs as a contract/behavior problem, not as harmless cleanup.
- Do not continue feature implementation until the owning docs/contracts are brought back in line and the maintainer confirms the corrected scope.

If multiple docs disagree, do not silently pick the one that matches the code. Use the authority order from `CLAUDE.md`: API docs, domain docs, frontend docs, migrations, and config examples are the contract sources; plan/progress docs must then be synchronized to that contract.

## AI Workflow Changes

- For AI workflow-only edits, update the AI workflow files directly after checking they do not introduce or override project contracts.
- For AI workflow changes, use `$pvecloud-skill-quality` before final verification.
- Keep the work inside `CLAUDE.md` and `.claude/skills/` unless a mismatch with project docs forces a contract update.
- Before finalizing AI workflow changes, do a text-level consistency pass against `CLAUDE.md`, the changed skill, and the related project docs.

## Progress Docs Rule

`docs/progress/` is a stage ledger, not the final product contract.

- Use progress docs to understand what was done, why scope changed, and what was accepted.
- Do not use progress docs as the deciding source when they conflict with API, domain, frontend, migration, or config contracts.
- When a feature or stage is completed, make sure the durable facts are reflected in the owning contract docs first.
- If old progress docs no longer describe the current contract, update, archive, or mark them as historical before continuing implementation.
- When reading archived or historical progress docs, extract context only; re-check current contracts before coding.

## Commit Message Rule

When the maintainer asks AI to commit, the commit message must follow Conventional Commits 1.0.0 (https://www.conventionalcommits.org/en/v1.0.0/) and remain useful for review from another machine. Explanatory text defaults to Chinese unless the maintainer requests another language.

- Do not run `git add`, `git commit`, `git push`, or any other Git history/staging mutation unless the maintainer explicitly asks for that Git action in the current conversation.

- Use the subject format `<type>[optional scope][!]: <description>`.
- Use common types such as `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`, `style`, and `revert`.
- Use `feat` for new capabilities, `fix` for bug fixes, and `!` or a `BREAKING CHANGE:` footer for breaking changes.
- The subject description may be Chinese, for example `docs(workflow): 统一 AI 提交规范`.
- Include a detailed Chinese body for non-trivial changes.
- Explain in Chinese why the change was needed, not only what files changed.
- Group the body by meaningful areas such as docs, workflow, frontend, backend, database, verification, and risk.
- Use footers for issue references, reviewers, or breaking changes; breaking changes must use `BREAKING CHANGE: <description>`.
- Mention verification commands that were run, or explicitly say when no runtime verification was needed.
- Mention notable residual risks or follow-up constraints.
- Do not amend a commit that already matches `origin/*` unless the maintainer explicitly asks to rewrite published history.
- If the previous published commit message was too terse, create a new corrective commit that adds the missing workflow rule instead of rewriting remote history.

## Mandatory Workflow

1. Run `git status --short`.
2. Read the required files in the order above.
3. If code was pulled, already changed, or the user says work is unfinished, run the drift-first check.
4. Decide whether the task is contract/behavior work, pure UI/UX polish, AI workflow, or still needs clarification.
5. For contract/behavior work, update the owning docs or machine contracts first.
6. For contract/behavior work covered by `$pvecloud-contract-quality`, run that quality gate before stopping for confirmation.
7. After updating those docs/contracts, stop and ask the maintainer to confirm.
8. Do not implement contract/behavior code until the maintainer explicitly confirms.
9. For AI workflow-only changes, update AI workflow files directly and run `$pvecloud-skill-quality` before final verification.
10. For pure UI/UX polish, implement directly after reading the frontend guardrail.
11. Verify with the smallest meaningful tests or builds.
12. Report what changed, what was verified, residual risk, and needed follow-up suggestions.

## Non-Negotiable Rules

- Skills may point to docs, but they must not become a second API spec or schema document.
- Do not hide durable product behavior inside skills.
- Do not overwrite user changes.
- Do not mutate the Git staging area, local commits, or remote branches unless the maintainer explicitly asks for that Git action.
- Do not let completed phase notes keep driving new code after the durable contract docs have moved on.
- Do not assume docs are current just because they exist; validate them against the current code path before continuing an unfinished feature.
- Do not treat progress/plan docs as the only source of truth when they conflict with owning contract docs or implementation.
- Do not add more code on top of a known code/docs mismatch unless the maintainer explicitly asks for an emergency implementation path.
- Do not recreate removed frontend pages, routes, or menus unless the docs are updated first and the maintainer confirms reopening them.
- Keep `admin/` and `web/` independent. No shared frontend runtime package.
- `admin/` and `web/` currently exist; their page scope, route meaning, and business availability still come from the owning docs.
- Do not start a new business feature by implementing only one surface and planning to backfill the rest later.
- Do not treat `docs/web/` as evidence that every planned user-facing feature is already implemented or open.

## Stop Message

When docs/contracts were updated for a contract/behavior change, stop with a short confirmation message in the user's language:

```text
文档/契约已先更新。确认点：
- ...

请确认这些设计、契约或验收口径是否通过。你确认后我再进入实现。
```

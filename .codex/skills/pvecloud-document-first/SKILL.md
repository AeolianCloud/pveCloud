---
name: pvecloud-document-first
description: Enforce the pveCloud document-first workflow. Use whenever work may change API, schema, frontend behavior, permissions, config, deployment, business process, or AI collaboration workflow/skills. This skill defines AI workflow only and never replaces project contracts in docs/, migrations, or config.example.yaml.
---

# pveCloud Document First

## Purpose

This skill is the AI workflow entry for `pveCloud`.
It defines how AI should read, classify, pause, implement, and verify work.
It must not define API fields, response payloads, table schemas, config values, page contracts, or durable product behavior.

Use it whenever the task touches implementation, refactoring, migration, frontend behavior, request wrappers, routes, permissions, config, deployment, cross-module business rules, or AI collaboration workflow files such as `AGENTS.md` and `.codex/skills/`.

## Source Boundaries

- AI workflow, guardrails, and implementation habits: `.codex/skills/`
- Human-readable project facts and contracts: `docs/`
- Executable database contract: `server/migrations/`
- Executable config example contract: `server/config.example.yaml`

If a skill conflicts with a project document or machine contract, the project document or machine contract wins.

## Required Reading

Always start with:

1. `AGENTS.md`
2. This `SKILL.md`
3. `references/workflow.md`
4. `references/contract-quality.md` when the task touches state changes, external side effects, async jobs, permissions, config, schema, API, security, or cross-surface business behavior
5. The task-relevant guardrail:
   - backend: `references/backend.md`
   - database: `references/database.md`
   - frontend: `references/frontend.md`
   - operations: `references/operations.md`
   - historical basic admin scope: `.codex/skills/pvecloud-basic-admin/SKILL.md`
6. The matching owner docs or machine contracts:
   - server/API: `docs/server/`, `docs/server/api/`
   - security: `docs/security.md`
   - admin frontend: `docs/admin/`
   - web frontend: `docs/web/`
   - database: `docs/server/database/`, `server/migrations/`
   - development/operations: `docs/development/`, `docs/operations/`
   - config: `server/config.example.yaml`
   - plans/progress: `docs/analysis/`, `docs/plan/`, `docs/progress/`

When a docs directory contains a `README.md`, read that index before opening deeper files. Then load only task-relevant page, API, schema, config, or progress documents.

For AI workflow-only work, read `references/workflow.md`, the skill files being changed, and any project docs needed to check for stale or misleading project facts. Do not read unrelated product docs just to edit workflow guardrails.

## Classification

Before non-trivial edits, state:

```text
任务分类：
- 类型：契约/行为变更 | 纯 UI/UX | AI 工作流/协作规则
- 影响面：admin | web | server/api | database | operations | ai-workflow
- 已读 owner docs/contracts：
- 是否需要先停确认：
```

Contract/behavior changes include API, schema, auth, permissions, route meaning, request wrapper semantics, page workflow, state semantics, config shape, deployment behavior, transaction/storage behavior, validation, security hardening, sensitive data handling, audit behavior, or business process.

Pure UI/UX polish includes layout, spacing, colors, typography, icons, responsive presentation, and non-contract copy.

AI workflow changes include `AGENTS.md`, `.codex/skills/`, skill metadata, and skill references.

## Hard Gates

- Run `git status --short` before editing.
- Name the exact owner docs/contracts read before editing implementation files.
- If the working tree already has changes, code was pulled, or the user is continuing unfinished work, do a drift-first check before adding implementation.
- For contract/behavior work, update owner docs or machine contracts first, then stop for maintainer confirmation.
- For contract/behavior work covered by `references/contract-quality.md`, run that quality gate before stopping for confirmation; do not treat vague statements such as "must be idempotent" or "must be recoverable" as sufficient contract.
- Do not implement contract/behavior code until the maintainer confirms.
- If implementation discovers new behavior, validation, hardening, storage, permission, route, config, or transaction changes, return to the document-first gate.
- For AI workflow-only changes, update AI workflow files directly after checking they do not introduce or override project contracts.
- Before final response, compare implementation and owner docs/contracts for drift.
- Verify with the smallest meaningful tests or builds.
- Do not mutate Git staging, commits, or remotes unless the maintainer explicitly asks.

Detailed workflow, drift, coupled-surface, progress-doc, commit-message, and document-sweep rules live in `references/workflow.md`.

## Stop Message

When docs/contracts were updated for a contract/behavior change, stop with:

```text
文档/契约已先更新。确认点：
- ...

请确认这些设计、契约或验收口径是否通过。你确认后我再进入实现。
```

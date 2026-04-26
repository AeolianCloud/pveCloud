# pveCloud Workflow Guardrails

This file defines how AI agents work in this repository. It does not define API fields, database schema, product behavior, or page content.

## Reading Order

1. `AGENTS.md`
2. `.codex/skills/pvecloud-document-first/SKILL.md`
3. The relevant guardrail file under `.codex/skills/pvecloud-document-first/references/`
4. The matching project docs or machine contracts under `docs/`, `server/migrations/`, or `server/config.example.yaml`

## Ownership Boundary

| Content | Source of truth |
| --- | --- |
| AI workflow and gates | `.codex/skills/pvecloud-document-first/` |
| API contracts and conventions | `docs/server/api/` |
| API conventions | `docs/server/api/conventions.md` |
| Backend architecture and business design | `docs/server/` |
| Database design | `docs/server/database/design.md` |
| Executable schema | `server/migrations/` |
| Admin frontend design | `docs/admin/architecture.md` |
| Web frontend design | `docs/web/architecture.md` |
| Local setup and operations | `docs/development/`, `docs/operations/` |
| Config example | `server/config.example.yaml` |

Skill references may point to these files and add implementation guardrails. They must not become a second API or schema document.

## Document-First Gate

For implementation, migration, API, page behavior, config, deployment, operations, or business-process changes:

1. Run `git status --short`.
2. Read the relevant skill guardrail.
3. Read the matching project docs and machine contracts.
4. Decide whether the work changes contracts/behavior or is pure UI/UX polish.
5. For contract/behavior work, update the matching docs/contracts first. For API changes, update `docs/server/api/` and the matching backend/frontend docs.
6. Stop and ask the maintainer to confirm the design or contract.
7. Implement contract/behavior work only after explicit confirmation.
8. For pure UI/UX polish, implement directly after reading frontend guardrails; do not write style-only preferences into docs.
9. Verify with focused tests/builds.
10. At wrap-up for branch-based work, merge the completed branch into the maintained base branch after approval, then delete local branches already merged into that base. Keep unmerged branches until the maintainer explicitly approves deletion.

Read-only investigation and pure explanation do not need the confirmation gate unless edits become necessary.

Pure UI/UX polish means visual-only changes such as spacing, layout, colors, typography, icons, component density, responsive presentation, or non-contractual wording. It does not include new routes, new actions, changed permissions, changed API calls, changed state semantics, or changed business workflow.

## Collaboration Rules

- Never overwrite user changes.
- If dirty files are unrelated, leave them alone.
- If dirty files overlap with the task, work with them instead of reverting.
- Keep final reports concise and include verification results after code changes.

## Git Commit Rules

- Commit messages must describe the real scope and outcome of the change. Do not use vague one-line messages such as `update`, `fix`, `change`, `wip`, or `调整`.
- Use a clear Conventional Commit style subject, for example `feat: complete basic admin foundation` or `fix: restore admin auth session validation`.
- For non-trivial changes, include a multi-line commit body that records the main modules changed, important behavior or contract effects, and verification commands that passed.
- Mention user-visible behavior, API/schema/menu/permission changes, data safety changes, or migration effects when they exist.
- Do not include unrelated dirty files in a commit. If files such as `.claude/` are unrelated or untracked, leave them out unless the maintainer explicitly asks to commit them.
- After a completed branch is merged, check stale local branches and clean up branches already merged into the maintained base branch. Report any remaining unmerged branches instead of silently deleting them.

## Code Style

- Public functions, types, interfaces, and package variables use block comment style: `/** ... */`.
- The first comment paragraph is Chinese and explains purpose.
- API handler comments can use `@route`, `@request`, `@response`, and `@auth`; keep them aligned with `docs/server/api/`.
- Internal comments should be Chinese and explain non-obvious business rules, idempotency, transactions, permissions, compensation, concurrency, and external-system boundaries.
- Do not write empty comments that only restate the code.
- Logs, errors, API `message`, CLI prompts, and user/admin/operator-facing output use Chinese.
- If a third-party protocol requires fixed English output, return English and explain nearby in Chinese comments.

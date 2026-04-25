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
| API source fragments | `docs/server/api/openapi-src/` |
| API generated contract | `docs/server/api/openapi.yaml` |
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
5. For contract/behavior work, update the matching docs/contracts first. For API changes, edit `docs/server/api/openapi-src/` and run `node ./scripts/generate-openapi.mjs`; do not hand-edit generated `docs/server/api/openapi.yaml`.
6. Stop and ask the maintainer to confirm the design or contract.
7. Implement contract/behavior work only after explicit confirmation.
8. For pure UI/UX polish, implement directly after reading frontend guardrails; do not write style-only preferences into docs.
9. Verify with focused tests/builds.

Read-only investigation and pure explanation do not need the confirmation gate unless edits become necessary.

Pure UI/UX polish means visual-only changes such as spacing, layout, colors, typography, icons, component density, responsive presentation, or non-contractual wording. It does not include new routes, new actions, changed permissions, changed API calls, changed state semantics, or changed business workflow.

## Collaboration Rules

- Never overwrite user changes.
- If dirty files are unrelated, leave them alone.
- If dirty files overlap with the task, work with them instead of reverting.
- Keep final reports concise and include verification results after code changes.

## Code Style

- Public functions, types, interfaces, and package variables use block comment style: `/** ... */`.
- The first comment paragraph is Chinese and explains purpose.
- API handler comments can use `@route`, `@request`, `@response`, and `@auth`; keep them aligned with OpenAPI.
- Internal comments should be Chinese and explain non-obvious business rules, idempotency, transactions, permissions, compensation, concurrency, and external-system boundaries.
- Do not write empty comments that only restate the code.
- Logs, errors, API `message`, CLI prompts, and user/admin/operator-facing output use Chinese.
- If a third-party protocol requires fixed English output, return English and explain nearby in Chinese comments.

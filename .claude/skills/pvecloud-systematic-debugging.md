---
name: pvecloud-systematic-debugging
description: Use when investigating pveCloud bugs, failing tests, regressions, broken builds, runtime errors, or unexpected behavior. Forces reproduce-first debugging, evidence-based root cause analysis, and fallback to pvecloud-document-first when a fix changes contracts or behavior.
---

# pveCloud Systematic Debugging

## Purpose

Use this skill to debug pveCloud failures without guessing.
It is an investigation workflow, not a place for product facts.

If a fix may change API shape, route semantics, permissions, state transitions, database behavior, config, deployment, security handling, frontend workflow, or cross-surface business behavior, switch to `$pvecloud-document-first` before editing contracts or implementation.

## When To Use

Use this skill for:

- failing tests, builds, type checks, migrations, or startup
- user-reported bugs, regressions, inconsistent UI behavior, or broken flows
- suspicious logs, panics, unhandled errors, or incorrect data
- "fix this", "why is this failing", "debug", "报错", "有 bug", "不对劲"

Do not use this skill for:

- new feature planning without a concrete failure
- pure visual polish with no broken behavior
- historical scope questions; use `$pvecloud-project-context` or `$pvecloud-basic-admin` as appropriate

## Required Setup

1. Run `git status --short`.
2. Read `CLAUDE.md`.
3. Read `$pvecloud-document-first` enough to classify whether the eventual fix could be contract/behavior work.
4. Read only the owner docs and code paths needed to understand the failing surface.

For an existing dirty worktree or continued unfinished work, perform the drift-first check from `$pvecloud-document-first` before changing code.

## Debugging Loop

1. Capture the exact symptom.
   - Record the command, route, page, request, log line, stack trace, or reproduction path.
   - Do not summarize away the first failing error.
2. Reproduce or narrow the failure.
   - Prefer the smallest command or UI path that demonstrates the problem.
   - If reproduction is impossible locally, state what evidence is available and what is missing.
3. Compare against a known-good pattern.
   - Search nearby tests, handlers, components, stores, migrations, wrappers, or docs for the same pattern.
   - Check whether the bug is a local inconsistency or a contract/document drift.
4. Form one concrete hypothesis.
   - Tie it to files, functions, state, data, or config.
   - Avoid stacking multiple speculative fixes.
5. Test the hypothesis with a targeted read, log, test, or minimal code inspection.
6. Apply the smallest fix that addresses the verified cause.
7. Re-run the reproducer and the smallest meaningful regression checks.

If a hypothesis fails, say what it ruled out and choose the next one from evidence.

## Root-Cause Discipline

- Do not apply a fix until the failure has a concrete root-cause hypothesis tied to code, data, config, docs, or environment.
- Do not stack speculative fixes. Change one causal area, then re-run the reproducer or targeted check.
- If the same symptom survives two attempted fixes, stop editing and rebuild the evidence trail from the first failing output.
- If a third attempt would be needed, pause and report what was tried, what each attempt ruled out, and which owner docs or architecture assumptions may be wrong.
- For multi-component failures, add or run diagnostics at the boundary first: request/response, state transition, database row, external adapter, route guard, or build step.
- Prefer creating a formal regression test when fixing security, permission, state, transaction, idempotency, request-wrapper, or data-loss bugs.
- For bug fixes and behavior changes, capture a failing command, failing test, reproducible UI path, or documented current gap before changing code when practical.
- If no formal test exists and adding one would exceed the task scope, use the strongest available reproducible check and state the remaining risk.

## pveCloud Gates

- API, permission, auth, route, request wrapper, error code, state, transaction, idempotency, schema, config, deployment, security, or page workflow changes require `$pvecloud-document-first` before implementation.
- Admin frontend must stay within `/admin-api/*`; web frontend must stay within `/api/*`.
- Do not treat code behavior as contract when owner docs say a feature or page is not open.
- Do not introduce new dependencies until checking existing wrappers, official SDKs, or maintained libraries fit the documented contract.
- Do not keep temporary repro scripts, logs, generated fixtures, or probe data after verification unless they become formal tests.

## Output Shape

When reporting progress or final results, prefer:

```text
Symptom:
Root cause:
Fix:
Verification:
Remaining risk:
```

For unresolved failures:

```text
Observed:
Ruled out:
Most likely next cause:
Blocked by:
```

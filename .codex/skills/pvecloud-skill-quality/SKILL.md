---
name: pvecloud-skill-quality
description: Use when creating, editing, reviewing, or finalizing pveCloud AI workflow skills. Checks metadata, trigger clarity, progressive disclosure, contract boundaries, validation rules, and consistency with AGENTS.md and owner docs.
---

# pveCloud Skill Quality

## Purpose

Use this skill to keep pveCloud AI workflow skills useful, concise, and within their lane.
It is a quality gate for `.codex/skills/` and `AGENTS.md`, not a product documentation source.

## When To Use

Use this skill when:

- adding or changing `.codex/skills/**/SKILL.md`
- adding or changing `.codex/skills/**/agents/openai.yaml`
- changing `AGENTS.md` workflow rules
- reviewing AI workflow drift, stale skill references, or skill bloat

## Quality Checklist

Check:

- front matter has `name` and `description`
- directory name, front matter `name`, and `$skill-name` references match
- `agents/openai.yaml` has a `default_prompt` that explicitly mentions `$skill-name`
- description clearly says when the skill should trigger
- body says when to use, when to skip, required inputs, and output or verification shape
- skill only defines AI workflow, guardrails, or implementation habits
- skill does not define API fields, response payloads, table schemas, config values, page contracts, feature availability, or durable product behavior
- long details are moved to `references/` when `SKILL.md` becomes hard to scan
- references are one level deep from `SKILL.md` and only loaded when relevant
- rules do not conflict with `AGENTS.md`, owner docs, migrations, or config examples
- temporary validation scripts or probes are not left behind unless intentionally part of the skill

## Review Method

Use targeted text checks:

```text
git status --short .codex/skills AGENTS.md
find .codex/skills -maxdepth 3 -type f | sort
rg -n "pattern" .codex/skills AGENTS.md
```

Choose search patterns based on the suspected issue, such as old skill names, stale paths, missing `$skill-name`, or accidental product facts.

## Output Shape

```text
技能质量自检：
- 元数据：
- 触发与退出：
- 渐进披露：
- 契约边界：
- 引用一致性：
- 验证：
- 需要修正：
```

If a skill conflicts with project docs or machine contracts, report the conflict and fix the skill, not the project contract.

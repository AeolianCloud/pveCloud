# Operations Implementation Guardrails

This file is for AI implementation rules. Local setup and deployment facts live in `docs/development/local-setup.md`, `docs/operations/deployment.md`, and `server/config.example.yaml`.

## Required Docs

Read these before setup, config, deployment, or operations work:

- `docs/development/local-setup.md`
- `docs/operations/deployment.md`
- `server/config.example.yaml`
- `docs/server/go-technical.md` when backend startup/config behavior changes

## Implementation Rules

- Config is YAML based; do not introduce `.env` as the main source.
- Keep real secrets out of git.
- Update `server/config.example.yaml` for supported config keys.
- Update operations docs when startup order, deployment components, proxy routes, backup expectations, or recovery procedures change.
- Worker exposes no public business HTTP endpoint unless explicitly documented and approved.
- PVE/payment/notify operational behavior must be recoverable and auditable where documented.

## Verification Baseline

- Validate example config still loads when config behavior changes.
- Run focused backend tests for config/bootstrap changes.
- For deployment docs, include concrete commands or paths that an operator can follow.

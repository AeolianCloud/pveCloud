# ADR 001: MariaDB Is The Async Task Source Of Truth

## Status

Accepted on 2026-04-22.

## Decision

Use `async_tasks` in MariaDB as the only task truth for backend async execution.
Redis may assist dispatch and short-lived coordination, but it is not authoritative.

## Consequences

- worker can recover task execution after Redis loss or worker restart
- payment callback and task creation can share one backend transaction boundary
- idempotency keys stay anchored on relational business records instead of cache state

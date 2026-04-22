# ADR 002: Capacity Reservation Before Payment Completion

## Status

Accepted on 2026-04-22.

## Decision

Create short-lived `resource_reservations` before payment completion to reduce oversell risk.
Pending orders hold a reservation briefly; payment success consumes that reservation into final allocation.

## Consequences

- unpaid expired orders can release reserved capacity automatically
- order creation can fail early when no saleable capacity exists
- provisioning can consume an existing reservation instead of rechecking from scratch

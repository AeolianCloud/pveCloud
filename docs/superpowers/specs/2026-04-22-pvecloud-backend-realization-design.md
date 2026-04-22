# PVE Cloud Backend Realization Design

**Date:** 2026-04-22  
**Status:** Draft approved for planning  
**Scope:** Subproject 1 of the full "complete implementation" effort

## 1. Goal

Turn the current MVP foundation slice from a skeleton implementation into a real backend business closure.

This design covers the first subproject only:

- real MariaDB-backed repositories
- real transaction boundaries
- real order, payment, async task, provisioning, and instance state flow
- real callback handling for payments
- real worker claim and execution flow
- real public/admin API routes required for the business closure
- mock resource adapter with production-style contract

This design does **not** include:

- real third-party payment merchant integration
- real PVE or other VM platform integration
- full frontend integration
- non-essential APIs outside the core closure

## 2. Decomposition

The complete "full implementation" request is split into three subprojects:

1. **Subproject 1: Backend real closure**
   - selected for immediate implementation
   - makes the backend truly runnable with MariaDB as source of truth
   - keeps payment and resource integrations configurable and replaceable

2. **Subproject 2: External adapter integration**
   - replace the mock payment/provider integration with real external systems
   - replace the mock resource adapter with the real VM or PVE API

3. **Subproject 3: Frontend full integration**
   - connect `web/` and `admin/` to the real backend APIs
   - add real auth flow, data refresh, operations, and error handling

Subproject 1 is intentionally the only implementation target for the next planning step.

## 3. Chosen Approach

Three implementation approaches were considered:

1. **Vertical business closure first**
   - recommended and selected
   - implement one real business chain end-to-end:
     user auth -> product browse -> create order -> payment callback -> task creation -> worker provisioning -> instance query
   - gives the earliest real value and the cleanest regression path

2. **Repository-first horizontal buildout**
   - implement repositories for all modules first, then services and APIs
   - structurally neat, but delays runnable business closure too long

3. **API-first surface completion**
   - make handlers and routes look complete first
   - rejected because it preserves fake internals and creates maximum rework

The selected approach is **vertical business closure first**, while still enforcing proper repository and transaction boundaries.

## 4. Architectural Boundaries

### 4.1 Included Modules

The following modules are in scope for Subproject 1:

- `user`: registration, login, identity lookup
- `adminuser`: admin login
- `catalog`: product listing, product detail, saleability lookup, capacity reservation
- `billing`: billing snapshot generation
- `order`: order creation, order query, order state transitions
- `payment`: payment order creation, callback verification entry, callback idempotency, payment status query
- `task`: task creation, claim, retry metadata, task query
- `instance`: provisioning facts, instance query, service period facts
- `resource`: mock provider behind a stable adapter contract
- `audit`: business event recording
- `notification`: business event notification recording or dispatch entry

### 4.2 Excluded Modules

The following are explicitly out of scope for this subproject:

- real payment vendor API integration
- real VM platform integration
- advanced instance actions such as reboot, reinstall, stop, start
- full admin operating workflows beyond essential queries
- frontend business flow completion

### 4.3 Runtime Shape

Three backend entrypoints remain:

- `public-api`
- `admin-api`
- `worker`

MariaDB remains the sole source of truth for:

- users
- catalog saleability facts
- orders
- billing snapshots
- payment orders and callback logs
- async tasks
- instances and service facts

Redis may be used only as support infrastructure and must not become the truth source for business state.

## 5. Core Business Closure

### 5.1 Closure Flow

Subproject 1 must make the following business closure real:

1. user registers or logs in
2. user queries saleable products
3. user creates an order
4. backend creates billing snapshot and payment order
5. payment callback arrives and is verified against configured payment settings
6. payment success moves order to `paid` and creates one `create_instance` async task
7. worker claims the task and executes provisioning
8. provisioning writes instance and service period facts
9. order moves to `active`
10. user can query their instances

### 5.2 Transaction Boundaries

Three transaction boundaries are mandatory:

1. **Order creation transaction**
   - create `orders`
   - create `billing_records`
   - create `payment_orders`
   - bind or persist reservation relation

2. **Payment success transaction**
   - record callback log
   - move `payment_orders.pending -> success`
   - move `orders.pending_payment -> paid`
   - create one unique async provisioning task

3. **Provision success transaction**
   - create `instances`
   - create `instance_services`
   - consume reservation
   - move `orders.paid -> active`
   - move `async_tasks.processing -> success`

No resource action may be triggered inside the payment callback transaction.

## 6. State Machines

### 6.1 Order State

Field: `orders.order_status`

Allowed values:

- `pending_payment`
- `paid`
- `provisioning`
- `active`
- `failed`
- `closed`

Allowed transitions:

- `pending_payment -> paid`
- `paid -> provisioning`
- `provisioning -> active`
- `pending_payment -> closed`
- `provisioning -> failed`
- `failed -> provisioning`

### 6.2 Payment State

Field: `payment_orders.pay_status`

Allowed values:

- `pending`
- `success`
- `failed`
- `refunded`

Allowed transitions:

- `pending -> success`
- `pending -> failed`
- `success -> refunded`

### 6.3 Reservation State

Field: `resource_reservations.status`

Allowed values:

- `reserved`
- `consumed`
- `released`
- `expired`

Allowed transitions:

- create order -> `reserved`
- provision success -> `consumed`
- cancel or timeout -> `released` or `expired`

### 6.4 Task State

Field: `async_tasks.status`

Allowed values:

- `pending`
- `processing`
- `success`
- `failed`
- `retrying`

Allowed transitions:

- `pending -> processing -> success`
- `processing -> retrying -> processing`
- `processing -> failed`

### 6.5 Instance State

Field: `instances.instance_status`

Allowed values retained:

- `creating`
- `running`
- `stopped`
- `reinstalling`
- `starting`
- `stopping`
- `failed`
- `expired`

Subproject 1 actively uses only:

- `running`
- implicit failure through task and order failure handling

Later subprojects may add real operational transitions.

## 7. Idempotency

Three idempotency guarantees are required:

1. payment callback idempotency key:
   - `payment_order_no`

2. provisioning task uniqueness key:
   - `task_type + business_type + business_id`

3. reservation consumption relation:
   - `order_id <-> reservation_id`

Expected behavior:

- duplicate payment callbacks do not create duplicate tasks
- duplicate task creation requests return the same logical task
- worker retry does not create duplicate instances

## 8. Repository and Data Access Design

### 8.1 Data Access Technology

Use `database/sql` for Subproject 1.

Reasoning:

- the codebase already uses it
- transaction handling and explicit locking are clearer here
- this phase needs control and correctness more than ORM abstraction

### 8.2 Shared Database Helpers

Add shared helpers in `internal/common/database`:

- `WithTx(ctx, db, fn)`
- small row scan helpers only where repetition is meaningful

These helpers must not absorb business rules.

### 8.3 Module Repositories

Suggested repository responsibilities:

- `user/repository.go`
  - create user
  - find by phone
  - find by id

- `adminuser/repository.go`
  - find admin by username

- `catalog/repository.go`
  - list saleable products
  - get product detail
  - find saleable node
  - create reservation
  - get reservation for update
  - consume reservation
  - release expired reservations

- `order/repository.go`
  - create order transaction inputs
  - get order by number
  - list orders by user
  - get paid order for provisioning
  - update order state

- `billing/repository.go`
  - create billing record
  - get billing snapshot by order id

- `payment/repository.go`
  - create payment order
  - get payment by number
  - insert callback log
  - mark success in transaction

- `task/repository.go`
  - create task in transaction
  - claim pending task
  - mark success
  - mark retry
  - mark failed
  - append task log
  - list tasks

- `instance/repository.go`
  - create instance in transaction
  - create instance service fact in transaction
  - get instance by id
  - list instances by user
  - list all instances

### 8.4 Service Responsibilities

Services orchestrate business flow only:

- validate business inputs
- call repositories
- control state transitions
- define transaction boundaries
- call external adapters

Services must not directly scatter SQL across modules.

## 9. Configuration Model

Configuration remains YAML-only in `server/config/config.yaml`.

Subproject 1 extends config with nested models:

```yaml
payment:
  provider: mock
  callback_base_url: http://127.0.0.1:8080
  notify_path: /payments/callback
  merchant_id: ""
  merchant_secret: ""

resource:
  provider: mock
  api_endpoint: ""
  api_token: ""

worker:
  poll_interval: 3s
  batch_size: 10
```

Config struct additions:

- `PaymentConfig`
- `ResourceConfig`
- `WorkerConfig`

Rules:

- payment callback processing must read provider configuration from YAML
- no `.env`
- no environment-variable overrides
- resource provider must be swappable without changing order, task, or instance business logic

## 10. HTTP API Surface

Subproject 1 implements only the APIs required for the core closure.

### 10.1 Public API

- `POST /auth/register`
- `POST /auth/login`
- `GET /products`
- `GET /products/:id`
- `POST /orders`
- `GET /orders`
- `GET /payments/:paymentOrderNo`
- `POST /payments/callback`
- `GET /instances`
- `GET /instances/:id`

JWT protection is required for:

- `POST /orders`
- `GET /orders`
- `GET /payments/:paymentOrderNo`
- `GET /instances`
- `GET /instances/:id`

### 10.2 Admin API

- `POST /auth/login`
- `GET /products`
- `GET /orders`
- `GET /instances`
- `GET /tasks`

All admin query APIs must be protected by admin JWT middleware.

### 10.3 Route Organization

`bootstrap/app.go` may continue using `http.ServeMux`, but route registration must be grouped clearly by module.

This subproject does not require switching routers.

## 11. Resource Adapter Strategy

Resource integration stays behind `resource.VMClient`.

Subproject 1 ships a mock provider implementation with production-style behavior:

- deterministic create VM response
- stable contract for request and response
- no business logic leakage into the adapter

Requirements:

- worker and instance services call the adapter exactly as they would call a real provider
- future real provider replacement must be isolated inside `internal/resource`

## 12. Payment Adapter Strategy

Payment handling is split into two layers:

1. **Business payment service**
   - creates payment orders
   - handles callback processing
   - manages status transitions
   - writes callback logs

2. **Configurable provider verifier**
   - reads configured provider type and secrets
   - validates callback payload according to selected provider

For Subproject 1:

- provider remains configurable
- callback route is real
- callback state transitions are real
- merchant credentials are expected to be user-provided later

This lets the later real payment setup change configuration without forcing redesign of business logic.

## 13. Error Handling

### 13.1 API Errors

Public-facing APIs must return structured application errors:

- invalid input -> `bad_request`
- failed auth -> `unauthorized`
- permission mismatch -> `forbidden`
- business conflict -> `conflict`
- unexpected failure -> `internal_error`

### 13.2 Callback and Worker Errors

Payment and worker paths must preserve more operational detail:

- payment callback failures must be logged with payment order context
- task execution failures must append task logs
- retryable resource failures must move tasks to `retrying`
- terminal failures must land in `failed`

### 13.3 Consistency Requirements

No external call result may partially update business facts outside a defined transaction boundary.

If a transaction fails:

- order must not advance
- payment must not falsely mark success
- task must not falsely mark success
- instance facts must not partially persist

## 14. Testing Strategy

Subproject 1 requires three layers of testing.

### 14.1 Unit Tests

Keep service-level tests for:

- state transitions
- idempotency rules
- error branches
- callback behavior

### 14.2 Repository Integration Tests

Add MariaDB-backed tests for:

- `CreateOrderTx`
- `MarkPaymentSuccessTx`
- `ClaimPendingTask`
- `CreateInstanceTx`

Focus:

- transaction consistency
- uniqueness enforcement
- row-lock correctness
- state transition correctness

### 14.3 End-to-End Backend Closure Tests

Replace the fake `internal/e2e` harness with a real DB-backed flow:

1. seed or create user
2. query saleable product
3. create order
4. simulate payment callback
5. run one worker cycle
6. verify order, task, and instance facts

The e2e result must prove:

- order becomes `active`
- task becomes `success`
- instance becomes visible and `running`
- no duplicate provisioning occurs

## 15. Implementation Order

Subproject 1 should be implemented in this order:

1. config extensions and shared transaction helpers
2. repositories for `catalog`, `billing`, `order`, and `payment`
3. real order creation and payment callback handling
4. repositories and services for `task`
5. repositories and services for `instance`
6. mock `resource` provider implementation behind final contract
7. public and admin API route completion
8. real e2e harness and test matrix

This order prioritizes making payment-to-provisioning real before adding broader surface area.

## 16. Success Criteria

Subproject 1 is complete when all of the following are true:

- all core business facts are stored in MariaDB
- payment callback creates exactly one provisioning task
- worker can claim and execute provisioning tasks
- provisioning writes instance and service facts
- public APIs support the full closure from auth to instance query
- admin APIs can query products, orders, instances, and tasks
- payment config is YAML-driven and ready for later real credentials
- resource adapter is swappable without business-layer redesign
- backend unit, repository, and e2e tests pass

## 17. Risks and Mitigations

### Risk 1: Scope explosion

Mitigation:

- enforce only the APIs required for the closure
- keep non-essential admin flows out of scope

### Risk 2: Service logic remains too fake

Mitigation:

- require repository-backed implementations
- require transaction-based integration tests

### Risk 3: Future adapter replacement becomes invasive

Mitigation:

- lock payment and resource contracts now
- keep provider-specific behavior inside adapter or verifier layers only

### Risk 4: Worker inconsistency under retries

Mitigation:

- persist retry metadata in MariaDB
- enforce unique task business key
- append task logs for debugging and replay analysis

## 18. Decision Summary

Subproject 1 will implement a **real backend closure** with:

- MariaDB-backed repositories
- real payment callback business processing
- real worker and task execution flow
- real instance fact persistence
- closure-only public and admin APIs
- mock resource provider behind a final contract

This design intentionally defers only the truly external integrations and the frontend real-business hookup.

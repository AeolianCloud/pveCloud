# pveCloud Database Rules

This reference consolidates MariaDB design rules. Executable schema remains in `server/migrations/`.

## Target

```text
database: pvecloud
engine: MariaDB 11.4.9 / InnoDB
charset: utf8mb4
collation: utf8mb4_unicode_ci
```

MariaDB is the business source of truth. PVE is an external resource system.

## Base Conventions

- Primary keys use `BIGINT UNSIGNED AUTO_INCREMENT`.
- Money fields use integer cents and `_cents` suffix.
- Status fields use `VARCHAR` and Go constants, not database enum.
- Tables and columns require MariaDB `COMMENT`.
- Time fields use `DATETIME(3)`.
- Common tables include `created_at` and `updated_at`; soft-deleted tables include `deleted_at`.
- External display uses business numbers such as `order_no`, `payment_no`, and `instance_no`, not auto-increment IDs.
- JSON fields store snapshots, third-party payloads, and config fragments; do not use them for high-frequency query conditions.

## Table Groups

Accounts and permissions:

```text
users
admin_users
admin_roles
admin_permissions
admin_user_roles
admin_role_permissions
```

Product catalog:

```text
products
product_plans
regions
pve_nodes
images
region_images
plan_prices
```

Orders, payments, wallets:

```text
orders
payment_orders
payment_notify_logs
wallet_accounts
wallet_transactions
```

Instances and jobs:

```text
instances
async_tasks
```

Tickets, config, audit:

```text
tickets
ticket_messages
system_configs
admin_audit_logs
```

## Key Business Rules

- Admin-only tables use `admin_` prefix.
- Admin permission codes use `domain:action`, for example `order:view` and `payment:manual_credit`.
- Product price uniqueness is `plan_id + region_id + billing_period`.
- Orders are one order to one instance in phase one. Do not add `quantity`.
- Store order product/price snapshots to protect historical orders from later price changes.
- `images` are logical OS images; real PVE templates are stored on `region_images.pve_template_id`.
- Users should only see images active in the selected region.
- Wallet transactions must write `balance_after_cents`.
- Wallet account balance changes use optimistic lock or row lock through `wallet_accounts.version`.
- `wallet_accounts.frozen_cents` is reserved only; add a freeze detail table before implementing real freeze flows.
- Refund status is currently reserved on `payment_orders`; add a refund table before supporting partial/multiple/refund-review flows.

## Key Constraints

- `orders.order_no` unique.
- `payment_orders.payment_no` unique.
- `payment_orders.channel + third_trade_no` unique when `third_trade_no` exists.
- `region_images.node_id + image_id` unique.
- `instances.order_id` unique.
- `instances.node_id + vmid` unique.
- `instances.provisioning_key` unique.
- `async_tasks.idempotency_key` unique.
- `wallet_accounts.user_id` unique.

## Important Indexes

- `users.email`, `users.phone`, `users.username`.
- `orders.user_id + created_at`, `orders.status + expired_at`.
- `payment_orders.status + created_at`.
- `region_images.region_id + status + sort_order`.
- `instances.user_id + status`, `instances.expire_at`.
- `async_tasks.status + run_at`, `async_tasks.locked_until`.
- `tickets.status + updated_at`.
- `admin_audit_logs.admin_id + created_at`, `admin_audit_logs.object_type + object_id`.

## Transactions

Order creation transaction:

- Validate plan, region, image, and price.
- Create `orders`.
- Create `payment_orders` or reserve balance-payment deduction logic.

Payment success transaction:

- Lock and update `payment_orders`.
- Branch by `payment_scene` and `order_type`.
- Update `orders` or `wallet_accounts`.
- Write `wallet_transactions`.
- Create unique `async_tasks`.

Instance provisioning should use at least two local transactions:

1. Claim task, lock order, create or reuse instance placeholder, persist VMID/idempotency anchors, and mark order provisioning.
2. After PVE task success, update instance, order, and task final state.

Do not call PVE HTTP inside a long DB transaction.

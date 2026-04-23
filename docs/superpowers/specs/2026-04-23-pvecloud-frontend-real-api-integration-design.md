# PVE Cloud Frontend Real API Integration Design

**Date:** 2026-04-23  
**Status:** Draft approved for implementation planning

## Goal

Replace the mock and hard-coded business data in `web/` and `admin/` with calls to the real backend APIs that already exist, while keeping the two frontend projects fully isolated.

## Scope

This design only covers pages backed by real backend routes that already exist.

### In Scope

#### `web/`

- User login
- User registration
- Product list
- Product detail
- Order creation
- Order list
- Payment status
- Instance list
- Instance detail

#### `admin/`

- Admin login
- Product list
- Order list
- Instance list
- Task list

#### Shared expectations

- Token-based authenticated requests
- Route guards for authenticated pages
- Consistent loading, empty, and error states
- Removal of fake business data from pages within scope

### Out of Scope

- `web` notices real API integration
- `admin` users real API integration
- `admin` dashboard real API integration
- Any backend changes to add new endpoints for the frontend
- Cross-project shared frontend package extraction

Out-of-scope pages should stop pretending to be wired to real business data and instead render a clear "not connected yet" style state.

## Existing Backend Contract

### `web` routes to consume

- `POST /auth/login`
- `POST /auth/register`
- `GET /products`
- `GET /products/{id}`
- `POST /orders`
- `GET /orders`
- `GET /payments/{paymentOrderNo}`
- `GET /instances`
- `GET /instances/{id}`

### `admin` routes to consume

- `POST /auth/login`
- `GET /products`
- `GET /orders`
- `GET /instances`
- `GET /tasks`

## Frontend Architecture

The two frontend apps remain independent. Each app gets its own internal API layer, auth store, route guard behavior, and page-level state handling.

### `web/` structure

- `src/lib/http.ts`: base request helper, auth header injection, response/error normalization
- `src/stores/auth.ts`: token state, login/register/logout actions
- `src/api/auth.ts`
- `src/api/catalog.ts`
- `src/api/order.ts`
- `src/api/payment.ts`
- `src/api/instance.ts`

### `admin/` structure

- `src/lib/http.ts`: base request helper, auth header injection, response/error normalization
- `src/stores/auth.ts`: token state, login/logout actions
- `src/api/auth.ts`
- `src/api/catalog.ts`
- `src/api/order.ts`
- `src/api/instance.ts`
- `src/api/task.ts`

### Why this structure

- It avoids pushing request logic into every page
- It keeps auth and error handling consistent
- It preserves the explicit isolation between `web` and `admin`
- It is small enough to fit the current MVP frontend scale

## Data Flow

### Authentication

- Login and registration responses store the issued token in the local auth store
- Authenticated requests include `Authorization: Bearer <token>`
- `web` protected routes:
  - `/orders`
  - `/instances`
  - `/instances/:id`
- `admin` protected routes:
  - all routes except `/login`
- On `401`, the app clears local auth state and redirects to login

### `web` user flow

1. User logs in or registers
2. Product list loads real products
3. Product detail loads a real product by id
4. Product detail allows creating an order
5. Successful order creation redirects to payment status using the real `paymentOrderNo`
6. Order list loads the authenticated user's orders
7. Payment status reads the real payment order status and allows refresh
8. Instance list and detail load the authenticated user's real instances

### `admin` management flow

1. Admin logs in
2. Product page loads real products
3. Order page loads real orders
4. Instance page loads real instances
5. Task page loads real async tasks

## View Behavior

### Pages that move from mock to real data

#### `web/`

- `LoginView.vue`
- `RegisterView.vue`
- `ProductListView.vue`
- `ProductDetailView.vue`
- `OrderListView.vue`
- `PaymentStatusView.vue`
- `InstanceListView.vue`
- `InstanceDetailView.vue`

#### `admin/`

- `LoginView.vue`
- `ProductManageView.vue`
- `OrderManageView.vue`
- `InstanceManageView.vue`
- `TaskManageView.vue`

### Pages left as explicit placeholders

#### `web/`

- `NoticeListView.vue`

#### `admin/`

- `DashboardView.vue`
- `UserManageView.vue`

These pages should display a simple honest status such as "backend capability not connected yet" instead of fake production data.

## Error Handling

- Each real-data page must visibly handle:
  - initial loading
  - request failure
  - empty result
- Error messaging stays simple and local to the page
- No global notification framework is introduced in this task

## Testing

Minimum required verification:

- Keep existing frontend tests passing
- Add or update at least one `web` test covering auth form submission or real-data rendering behavior
- Add or update at least one `admin` test covering a real-data page render path
- Run:
  - `bun --cwd web test`
  - `bun --cwd web run build`
  - `bun --cwd admin test`
  - `bun --cwd admin run build`

## Documentation Follow-up

After implementation, review `docs/superpowers` and update any task/spec status that still describes the frontend as mock-only where this task closes the gap.

## Constraints

- No backend API contract invention
- No frontend package sharing between `web` and `admin`
- No migration to a heavier data-fetching library
- No feature-flag or compatibility shim layer

## Success Criteria

- In-scope `web` pages use real backend data
- In-scope `admin` pages use real backend data
- Mock business lists and details are removed from in-scope pages
- Authenticated pages enforce login and carry bearer tokens
- Both frontend projects still build and test successfully

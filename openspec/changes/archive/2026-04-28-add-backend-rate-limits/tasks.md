## 1. Middleware foundation

- [x] 1.1 Add a reusable backend rate-limit middleware in `apps/backend-go/internal/transport/http/middleware/ratelimit.go` that tracks per-key request budgets and rejects over-limit requests with an explicit `429` response envelope.
- [x] 1.2 Add focused middleware-level tests covering within-budget requests, over-budget requests, and route-appropriate key selection behavior.

## 2. Login route protection

- [x] 2.1 Attach the rate-limit middleware to `POST /login/localLogin` and `POST /api/login/localLogin` in the auth route registration path without changing existing credential validation behavior for allowed traffic.
- [x] 2.2 Add handler or route-level tests verifying normal local login behavior remains intact within budget and throttled login attempts are rejected before credential validation.

## 3. Datasource validation protection

- [x] 3.1 Attach the rate-limit middleware only to canonical datasource validation routes (`POST /api/ds/validate` and `GET /api/ds/validate/:id`) without broadening throttling to unrelated datasource routes.
- [x] 3.2 Extend datasource handler or route tests to verify validate requests still succeed within budget and return explicit throttling failures when the validation budget is exceeded.

## 4. Audit export protection

- [x] 4.1 Attach the rate-limit middleware only to `POST /api/audit/export` and `GET /api/audit/download` while preserving the existing export/download path and format validation behavior.
- [x] 4.2 Extend audit handler tests to verify export and download requests remain functional within budget and return explicit throttling failures when the audit export budget is exceeded.

## 5. Verification and rollout safety

- [x] 5.1 Run targeted backend test suites for auth, datasource, and audit handler/middleware coverage and fix any regressions introduced by route-level throttling.
- [x] 5.2 Run repository-level backend validation (`make test` and any affected drift/security checks) and document the initial throttling thresholds plus rollback steps in the change record or implementation summary.

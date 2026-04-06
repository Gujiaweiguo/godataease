## 1. Governed route scope and middleware contract

- [x] 1.1 Audit the governed dataset/chart runtime entry points (`/dataset/previewWithPerm`, chart permission-aware handlers, and compatibility-bridge permission flows) and lock the exact first-wave routes that will adopt real `RowPermissionMiddleware()` behavior without changing non-governed endpoints.
- [x] 1.2 Define the middleware context contract for the first-wave routes, including how dataset or dataset-group identity, authenticated user prerequisites, and fail-closed error conditions are resolved and stored in Gin context without moving SQL rule compilation out of the services.
- [x] 1.3 Add or update backend middleware-focused tests that prove the chosen route payloads can establish governed row-permission context, and that missing or invalid context terminates the request instead of behaving like a warning-only pass-through.

## 2. Middleware enforcement rollout

- [x] 2.1 Replace the warning-only `RowPermissionMiddleware()` stub in `internal/transport/http/middleware/permission.go` with real context-validation and fail-closed behavior for the first-wave governed routes.
- [x] 2.2 Wire `RowPermissionMiddleware()` into the approved governed route registrations while preserving the existing dataset/dashboard permission middleware stack and avoiding accidental enforcement on non-governed routes.
- [x] 2.3 Update the affected handlers, services, and backend tests so middleware-established context composes cleanly with `PreviewWithPermission`, `QueryDataWithPermission`, and `ListByDQWithPermission`, while the service layer remains authoritative for row/column rule application.

## 3. Verification and scope control

- [x] 3.1 Run backend verification for the middleware rollout (`go test` for affected transport/service packages or `make test`) and confirm success, denial, unauthenticated, and fail-closed error scenarios for the governed routes in scope.
- [x] 3.2 Run `TEST_DB_HOST=127.0.0.1 make test-integration` if the rollout changes permission-sensitive runtime behavior beyond isolated middleware unit coverage, and record any environment blockers or pre-existing failures separately from change-induced failures.
- [x] 3.3 Perform a final scope check confirming this change only activates row-permission middleware enforcement for explicitly governed runtime routes and does not expand into whitelist persistence, system-variable support, or unrelated permission-center redesign.

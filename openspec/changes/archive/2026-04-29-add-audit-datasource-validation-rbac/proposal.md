## Why

The backend now rate-limits audit export/download and datasource validation, but both surfaces are still effectively available to any authenticated user. That leaves a gap where broad JWT access can still be used to probe datasource targets or retrieve audit exports without the finer-grained authorization boundaries expected for sensitive operations.

## What Changes

- Add finer-grained authorization requirements for audit export and audit download so they are not available to every authenticated user by default.
- Add finer-grained authorization requirements for datasource validation routes so validation cannot be used as a general-purpose probe surface by users without datasource management authority.
- Align canonical and compatibility datasource validation routes with the same authorization semantics.
- Add route-level backend verification covering allowed and denied authorization outcomes for the affected routes.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `audit-logs`: add authorization requirements for audit export and download flows.
- `datasource-validation-checking-canonical`: add authorization requirements for datasource validation routes and aliases.

## Impact

- Affected backend code:
  - `apps/backend-go/internal/transport/http/router.go`
  - `apps/backend-go/internal/transport/http/handler/audit_handler.go`
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_datasource_routes.go`
  - `apps/backend-go/internal/transport/http/middleware/menu_auth.go`
  - `apps/backend-go/internal/transport/http/middleware/permission_middleware.go`
- Affected runtime behavior:
  - Audit export/download may return `403` for authenticated users lacking audit authorization.
  - Datasource validation routes may return `403` for authenticated users lacking datasource validation authority.
- Affected verification:
  - Backend handler/router tests need extension for authorization allow/deny coverage on audit and datasource validation routes.
- Breaking changes:
  - Authenticated users who previously relied on broad JWT-only access to these routes may lose access unless they already hold the relevant audit or datasource permissions.

## Why

The Go backend currently hardens sensitive endpoints with authentication, SSRF checks, and path validation, but it does not enforce request-rate controls on expensive or abuse-prone entrypoints. Login, datasource validation, and audit export/download remain vulnerable to brute-force, probing, and resource-consumption patterns that should be constrained before broader security hardening continues.

## What Changes

- Add a minimal backend rate-limiting middleware for targeted HTTP routes in the Go backend.
- Apply rate limits to local login endpoints to reduce brute-force and repeated credential probing risk.
- Apply rate limits to datasource validation endpoints to reduce TCP probe abuse against user-supplied datasource targets.
- Apply rate limits to audit export and download endpoints to reduce repeated export generation and download abuse.
- Add targeted verification for the new rate-limit behavior at the handler/middleware level without changing existing success semantics for normal request volume.

## Capabilities

### New Capabilities
- `api-rate-limiting`: Provide route-scoped request throttling for security-sensitive and resource-intensive backend HTTP endpoints.

### Modified Capabilities
- `login-management`: Add request-throttling requirements for local login endpoints so repeated failed or burst login attempts are constrained.
- `datasource-validation-checking-canonical`: Add request-throttling requirements for datasource validation routes so validation cannot be abused as a high-frequency probe surface.
- `audit-logs`: Add request-throttling requirements for audit export and download flows so authenticated users cannot repeatedly generate or fetch exports without control.

## Impact

- Affected backend code:
  - `apps/backend-go/internal/transport/http/middleware/`
  - `apps/backend-go/internal/transport/http/router.go`
  - `apps/backend-go/internal/transport/http/handler/auth_handler.go`
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/audit_handler.go`
- Affected runtime behavior:
  - Login, datasource validate, and audit export/download requests may return explicit throttling responses when request volume exceeds configured limits.
- Affected verification:
  - Backend handler/middleware tests need extension for throttling behavior.
- Breaking changes:
  - None intended for normal request patterns; only abusive or bursty clients should observe new rejection behavior.

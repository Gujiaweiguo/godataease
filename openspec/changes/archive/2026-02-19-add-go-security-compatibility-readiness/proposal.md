# Change: Go backend security and compatibility readiness hardening

## Plan Version
Plan v1 (Atlas/Hephaestus unique execution baseline)

## Why
The Java-to-Go migration has broad route coverage, but production cutover risk is still concentrated in security semantics parity and Java-client compatibility edge cases. Without a focused hardening change, frontend cutover can pass smoke tests but fail in permission correctness, async export behavior, and legacy route contracts.

## What Changes
- Harden Java-compatible route behavior for high-risk modules (template and legacy compatibility paths).
- Enforce migration-safe contract semantics for unimplemented compatibility endpoints (no silent success).
- Close row-level and column-level permission parity gaps between Java and Go.
- Add compatibility gate tests (contract diff + negative security paths) and staging shadow validation.
- Add export-task behavior parity checks tied to authorization and async status transitions.

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
  - `specs/permission-config/spec.md`
  - `specs/export-center-management/spec.md`
  - `specs/template-management/spec.md`
- Affected code (expected):
  - `backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `backend-go/internal/transport/http/handler/template_handler.go`
  - `backend-go/internal/transport/http/router.go`
  - `backend-go/internal/transport/http/middleware/permission.go`
  - `backend-go/internal/service/*` (dataset/chart/export related parity logic)
- Risk profile:
  - High: security semantics mismatch (row/column permissions)
  - High: silent compatibility stubs returning success
  - Medium: template route-group mismatch with Java clients

## Execution Policy
This change's `tasks.md` is the only execution contract for Atlas/Hephaestus. No separate plan document is allowed.

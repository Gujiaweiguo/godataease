## Why

T6 exploration uncovered three production bugs in the user/role lifecycle layer: (1) two frontend API functions (`userOptionForRoleApi`, `userSelectedForRoleApi`) have reversed path segments relative to the backend canonical routes, causing 404s on role-option and role-selection calls from the RoleTab member management dialog; (2) `UserService.SetAuditService()` is never wired in `router.go`, making password-reset audit logging dead code in production; (3) `EditRole` contains a duplicate system-role guard that is harmless but confusing. These are the highest-impact, lowest-risk fixes to ship before deeper lifecycle alignment work.

## What Changes

- Fix `userOptionForRoleApi` path from `/user/role/option` to `/role/user/option` in `apps/frontend/src/api/user.ts`
- Fix `userSelectedForRoleApi` path from `/user/role/selected/:page/:limit` to `/role/user/selected` (pagination is in the request body, not URL params) in `apps/frontend/src/api/user.ts`
- Wire `userService.SetAuditService(auditService)` after `NewUserService(...)` in `apps/backend-go/internal/transport/http/router.go`
- Remove duplicate system-role guard in `EditRole` at `apps/backend-go/internal/service/role_service.go` L104-113
- Add integration test verifying audit log is created for password reset

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `role-in-user-management`: fixes frontend API path mismatch for role-option and role-selected endpoints used in the member management dialog
- `audit-go`: wires audit service dependency for UserService, enabling password-reset audit trail

## Impact

- **Frontend**: `apps/frontend/src/api/user.ts` — 2 path fixes (non-breaking, the canonical routes already exist)
- **Backend**: `apps/backend-go/internal/transport/http/router.go` — 1 line addition to wire audit service
- **Backend**: `apps/backend-go/internal/service/role_service.go` — remove 7 lines of duplicate code
- **Tests**: `apps/backend-go/internal/service/user_service_integration_test.go` — add audit verification test
- **No API contract changes** — canonical routes already exist; frontend was pointing to wrong paths
- **Rollback**: trivial — revert the single commit; no schema changes, no migration

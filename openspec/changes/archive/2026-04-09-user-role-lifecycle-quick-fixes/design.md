## Design

Three independent fixes, each atomic and independently verifiable.

### Fix 1: Frontend API path mismatch

**Problem**: `userOptionForRoleApi` and `userSelectedForRoleApi` in `api/user.ts` use reversed path segments (`/user/role/...`) vs the backend canonical routes (`/role/user/...`). This causes 404s on role-option and role-selection calls from the RoleTab member management dialog.

**Solution**: Change two path strings in `api/user.ts`:
- `/user/role/option` → `/role/user/option`
- `/user/role/selected/:page/:limit` → `/role/user/selected` (pagination is in request body, not URL params)

**Files**: `apps/frontend/src/api/user.ts` lines 10, 12-13

**Risk**: None — the backend routes at `/role/user/option` and `/role/user/selected` already exist and work.

### Fix 2: Wire audit service for UserService

**Problem**: `UserService.SetAuditService()` exists but is never called in `router.go`. Password reset audit logs (`ResetPasswordWithAudit`) silently discard because `auditSvc` is nil.

**Solution**: Add `userService.SetAuditService(auditService)` after line 162 in `router.go`, between userService creation and roleRepo initialization.

**Files**: `apps/backend-go/internal/transport/http/router.go` (1 line addition after L162)

**Risk**: Minimal — `ResetPasswordWithAudit` already handles nil auditSvc gracefully; wiring it just enables the audit path.

### Fix 3: Remove duplicate system-role guard in EditRole

**Problem**: `role_service.go` EditRole method has a duplicate system-role check at L104-106 and L111-113, plus a stale `if err != nil` at L107-109 that shadows the earlier error check.

**Solution**: Remove lines 107-113 (stale error check + duplicate system-role guard).

**Files**: `apps/backend-go/internal/service/role_service.go` (remove 7 lines)

**Risk**: None — the first check at L104-106 is preserved.

## Approach

- All three fixes go in a single commit (they're tightly coupled as "T6 quick fixes").
- Add one integration test verifying audit log creation on password reset.
- Run `make test` + `make test-integration` + `npm run lint` + `npm run ts:check` for verification.

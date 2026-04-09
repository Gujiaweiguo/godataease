## Tasks

- [x] T1. Fix frontend API path mismatch for role-user endpoints
  - File: `apps/frontend/src/api/user.ts`
  - Change `/user/role/option` → `/role/user/option` (line 10)
  - Change `/user/role/selected/:page/:limit` → `/role/user/selected` with page/limit in request body (lines 12-13)
  - Verification: `npm run lint` + `npm run ts:check`

- [x] T2. Wire audit service for UserService in router
  - File: `apps/backend-go/internal/transport/http/router.go`
  - Add `userService.SetAuditService(auditService)` after line 162 (after `NewUserService` creation)
  - Verification: `make test`

- [x] T3. Remove duplicate system-role guard in EditRole
  - File: `apps/backend-go/internal/service/role_service.go`
  - Remove lines 107-113 (stale error check + duplicate RoleTypeSystem guard)
  - Keep first check at lines 104-106
  - Verification: `make test`

- [x] T4. Add integration test for audit log on password reset
  - File: `apps/backend-go/internal/service/user_service_integration_test.go`
  - Add test `TestUserServiceIntegration_ResetPassword_WithAuditLog` that creates a user, resets password, and verifies an audit log entry exists
  - Verification: `TEST_DB_HOST=127.0.0.1 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration`

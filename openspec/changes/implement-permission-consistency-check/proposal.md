## Why

`CheckPermissionConsistency` in `resource_perm_repo.go` currently returns a hardcoded `Consistent: true` stub with zero counts and no database queries. The permission-config spec requires that by-user and by-resource views resolve to the same effective authorization state, but there is no runtime verification mechanism. Without a real cross-check, silently divergent permission states can go undetected.

## What Changes

- Replace the always-true stub in `ResourcePermissionRepository.CheckPermissionConsistency()` with a real cross-view SQL query that compares effective permissions from user-perspective (`GetUserResources`) against resource-perspective (`GetResourceUsers`) for all active users and governed resource types.
- Populate `PermissionConsistencyResult` with actual `UserCount`, `ResourceCount`, and `Inconsistencies` entries when divergence is detected.
- Add unit tests (fake store) and integration tests (MySQL) covering: consistent state, divergent state, empty data, and large dataset performance.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `permission-config`: The existing `Permission Dual-Perspective Consistency` requirement already mandates that both views persist to the same authorization model. This change implements the verification method that the spec already requires.

## Impact

- **Files modified**: `apps/backend-go/internal/repository/resource_perm_repo.go` (implementation), `apps/backend-go/internal/service/resource_perm_service_test.go` (unit tests), new integration test file.
- **APIs**: No new HTTP endpoints. Existing `CheckPermissionConsistency` service method returns real data instead of a stub.
- **Database**: Read-only queries against existing tables (`sys_user_perm`, `sys_role_perm`, `sys_user_role`, `sys_perm`, `sys_resource`, `sys_resource_perm`, `sys_user`). No schema changes.
- **Risk**: Low — purely additive read-only logic. No mutation paths affected.

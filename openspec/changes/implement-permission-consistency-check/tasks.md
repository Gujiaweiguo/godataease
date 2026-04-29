## 1. Repository Implementation

- [x] 1.1 Implement real `CheckPermissionConsistency` in `resource_perm_repo.go`: batch-collect user-view effective set (all active users × governed perm keys) via single SQL joining `sys_user_perm` + `sys_role_perm` + `sys_user_role` + `sys_perm`, grouped by `user_id` and `perm_key` prefix
- [x] 1.2 Batch-collect resource-view effective set: query all registered resources from `sys_resource` for governed types, then for each resource collect `(userID, permKey)` from `GetResourceUsers` pattern (direct + role-inherited), unioned across all resources
- [x] 1.3 In-memory diff: compare user-view set vs resource-view set, populate `Inconsistencies` with entries containing `UserID`, `ResourceID`, `ResourceType`, `UserView`, `ResourceView`
- [x] 1.4 Add performance guard: if `SELECT COUNT(*) FROM sys_user WHERE del_flag = 0` > 10,000, return early with `Consistent: true` and skip note

## 2. Unit Tests

- [x] 2.1 Add unit test `TestCheckPermissionConsistency_ConsistentState` in `resource_perm_service_test.go`: mock repo returns matching user/resource sets → `Consistent: true`, zero inconsistencies
- [x] 2.2 Add unit test `TestCheckPermissionConsistency_DivergentState`: mock repo returns mismatching sets → `Consistent: false`, non-empty inconsistencies with correct fields
- [x] 2.3 Add unit test `TestCheckPermissionConsistency_EmptySystem`: no users, no resources → `Consistent: true`, `UserCount: 0`, `ResourceCount: 0`
- [x] 2.4 Add unit test `TestCheckPermissionConsistency_SkipsLargeUserBase`: user count > 10,000 → early return, no full scan

## 3. Integration Tests

- [x] 3.1 Add integration test file `resource_perm_repo_consistency_integration_test.go` with `//go:build integration` tag
- [x] 3.2 Test consistent state: seed users, roles, perms, resource perms → call `CheckPermissionConsistency` → assert `Consistent: true` with correct counts
- [x] 3.3 Test divergent state: seed data with intentionally divergent grants (user has perm not on resource, or resource has perm not on user) → assert `Consistent: false` with expected inconsistency entries

## 4. Verification

- [x] 4.1 Run `make test` — all existing + new tests pass
- [x] 4.2 Run `make drift-check` — no contract drift introduced

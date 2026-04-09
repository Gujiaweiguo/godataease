## 1. Compatibility route semantic correction

- [x] 1.1 Update compatibility bridge registration so `/user/org/option` always maps to `user.GetUserOptions` and no longer conditionally maps to `org.ListOrgs`.
- [x] 1.2 Confirm route scope remains limited to `/user/org/option` without changing `/org/list`, `/org/tree`, or `/org/mounted` semantics in this slice.

## 2. Focused regression coverage

- [x] 2.1 Add compatibility bridge tests proving `/user/org/option` returns user-option payload semantics when both user and org handlers are present.
- [x] 2.2 Add compatibility bridge tests proving `/user/org/option` still returns user-option payload semantics when org handler is absent.

## 3. Verification

- [x] 3.1 Run focused handler tests for compatibility bridge changes.
- [x] 3.2 Run `make test` in `apps/backend-go` and fix only issues caused by this change.
- [x] 3.3 Run `TEST_DB_HOST=127.0.0.1 make test-integration` in `apps/backend-go` if permission/org-context-sensitive behavior requires integration confirmation.

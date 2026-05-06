## 1. Deferred Permission Dimension Registry

- [ ] 1.1 Create `DeferredDimensionRegistry` in `domain/permission/deferred_registry.go` with entries for `sysParams`, `whiteList`, and `dept`. Each entry has: dimension name, stable error code (e.g., `DEFERRED_DIMENSION_SYS_PARAMS`), and human-readable message. Provide `IsDeferred(dimName) bool` and `GetRejectionError(dimName) error` lookups.
- [ ] 1.2 Replace scattered hardcoded rejection messages in `data_permission_admin_service.go` (sysParams target rejection) and `SaveRowPermission` (whiteList rejection) with calls to `DeferredDimensionRegistry.GetRejectionError()`. Preserve existing test assertions by updating expected messages to match registry values.
- [ ] 1.3 Write unit tests for `DeferredDimensionRegistry`: all three dimensions return correct error codes, unknown dimensions return a generic deferred error, `IsDeferred` returns true for known dimensions and false for unknown.

## 2. Permission Mutation Audit Trail

- [ ] 2.1 Create `auditPermissionMutation()` shared helper in `service/audit_helper.go` (or extend existing audit package). Signature: `auditPermissionMutation(ctx context.Context, op string, actorID, targetID, resourceID, orgID int64, details map[string]interface{})`. Uses the same audit entry schema as C2.
- [ ] 2.2 Add audit calls to `ResourcePermissionService` grant/revoke methods: `GrantResourcePermission`, `RevokeResourcePermission`, and batch variants. Each call records operation type, actor, target user/role, resource, org, and timestamp.
- [ ] 2.3 Add audit calls to `DataPermissionAdminService` save methods: `SaveRowPermission`, `SaveColumnPermission`. Each call records operation type, actor, target dataset, org, and timestamp.
- [ ] 2.4 Write unit tests for audit helper and verify audit entries are produced for grant, revoke, save-row, and save-column operations.
- [ ] 2.5 Write integration tests for permission mutation audit trail using MySQL test database with `//go:build integration` tag.

## 3. Org-Scoped Permission Service Guards

- [ ] 3.1 Add `requireGovernedOrgContext()` call to `ResourcePermissionService` mutation methods: `GrantResourcePermission`, `RevokeResourcePermission`, batch grant/revoke. Reject with explicit error if org context missing or caller lacks org-scoped admin authority.
- [ ] 3.2 Add org scope filtering to `ResourcePermissionService` query methods: `GetUserPermissions`, `GetResourcePermissions`, `ListPermissions`. When org context is present and caller is not admin, filter results to the org boundary.
- [ ] 3.3 Add `requireGovernedOrgContext()` call to `DataPermissionAdminService` save methods: `SaveRowPermission`, `SaveColumnPermission`. Validate that the target dataset belongs to the active org.
- [ ] 3.4 Scope `CheckPermissionConsistency` to accept optional `orgID` parameter. When provided, filter both user-view and resource-view queries by org membership. When absent (admin), perform global check for backward compatibility.
- [ ] 3.5 Write unit tests for org-scoped rejection: cross-org grant rejected, cross-org dataset permission save rejected, org-scoped consistency check only scans within boundary.
- [ ] 3.6 Write integration tests for org-scoped permission operations using MySQL test database with `//go:build integration` tag.

## 4. Permission Middleware Org-Context Integration

- [ ] 4.1 Add org-context extraction to `PermissionMiddleware`: extract orgID from JWT claims (already populated by auth middleware), validate user membership via `userRoleRepo.IsUserInOrg()`, fail closed (403) if membership cannot be confirmed. Admin users bypass org membership check.
- [ ] 4.2 Add org-context extraction to `RowPermissionMiddleware`: validate that the target dataset belongs to the current org before enforcing row permissions. Reject with 403 if dataset-org mismatch.
- [ ] 4.3 Update `permission_compat_handler.go` to pass org context through to underlying service calls, ensuring legacy compat routes also enforce org scope.
- [ ] 4.4 Write unit tests for middleware org-context integration: valid org passes, invalid org returns 403, missing org returns 403, admin bypasses org check.
- [ ] 4.5 Write integration tests for middleware org enforcement on governed permission routes.

## 5. Organization-Scoped Permission Query Guards

- [ ] 5.1 Add org-scoped query guards to permission handler methods in `perm_handler.go` and `data_permission_handler.go`: extract org context from request, pass to service layer, ensure response is scoped to org for non-admin users.
- [ ] 5.2 Update `permission_compat_handler.go` query methods to respect org scope: `menuPermission`, `busiPermission` queries filter by org when org context is present.
- [ ] 5.3 Write unit tests for org-scoped query guards: non-admin sees only org resources, admin sees all, invalid org context returns error.
- [ ] 5.4 Write integration tests for org-scoped permission query guards using MySQL test database.

## 6. Regression and Rollout Safety

- [ ] 6.1 Run backend verification (`make test`, `make drift-check`) and confirm no regressions in existing permission tests.
- [ ] 6.2 Run integration tests (`TEST_DB_HOST=172.19.0.2 TEST_DB_PASSWORD=Admin168 TEST_DB_NAME=dataease_test make test-integration`) for all permission-related integration test suites.
- [ ] 6.3 Perform final scope check: confirm C3 only adds org-scoped guards, audit trail, and deferred registry — does not change permission SQL generation, cache strategy, or resource inheritance behavior.

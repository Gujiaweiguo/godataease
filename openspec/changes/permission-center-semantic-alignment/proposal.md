## Why

C1 `iam-foundation-semantics` established organization-scoped governance as the mandatory contract for all IAM write paths, and C2 `user-role-lifecycle-alignment` extended this to role assignment, user transfer, and member consistency. The permission center currently operates at global scope — all permission queries and mutations ignore organization boundaries — creating an org-isolation bypass vector. The deferred dimensions (sysParams, whiteList, dept auth targets) are tracked with scattered ad-hoc messages rather than a centralized registry, and permission mutations lack the audit trail pattern established in C2.

## What Changes

- Adopt org-scoped context for all governed permission write paths: save, grant, revoke, and consistency-check operations must consume active org ID and reject cross-org mutations.
- Add audit trail to governed permission mutations using the C2 audit pattern (operation type, actor, target, org, timestamp).
- Centralize deferred permission dimensions (sysParams, whiteList, AuthTargetTypeDept) into a `DeferredDimensionRegistry` with stable error codes, replacing scattered ad-hoc rejection messages.
- Integrate org-context resolution into permission middleware so `RowPermissionMiddleware` and `PermissionMiddleware` validate org membership before enforcement, failing closed when org context is unavailable.
- Scope `CheckPermissionConsistency` to operate within org boundaries so it only scans users and resources belonging to the active organization.
- **BREAKING**: Permission mutation requests that previously succeeded at global scope will now require valid org context and will be rejected if the caller lacks org-scoped admin authority or targets resources outside their org boundary.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `permission-config`: add org-scoped administration contracts, audit requirements for governed permission mutations, and centralized deferred-dimension registry with stable error codes.
- `role-management`: ensure role-permission binding operations respect org scope established in C1/C2.
- `organization-management`: add org-scoped permission query guards so admin permission views only surface resources within the org boundary.

## Impact

- **Affected backend modules**: `resource_perm_service.go`, `data_permission_admin_service.go`, `permission.go` middleware, `permission_middleware.go`, `resource_perm_repo.go`, `permission_compat_handler.go`, and related domain models.
- **Affected APIs**: permission save/query/grant/revoke flows will require org-context header and reject cross-org mutations; consistency-check endpoint will scope to active org.
- **Affected tests**: existing permission tests must be updated to provide org context; new tests for org-scoped rejection and audit coverage.
- **Dependencies**: builds on C1 org-scoped governance and C2 audit/org-membership patterns; no new infrastructure.
- **Rollback**: org-scoped guards can be feature-flagged to revert to global-scope behavior while preserving audit and deferred-registry improvements.

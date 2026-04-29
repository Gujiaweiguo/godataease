## Context

The permission-config spec mandates that by-user and by-resource views resolve to the same effective authorization state. The backend already has:
- `GetUserResources(userID, resourceType)` — queries effective permissions from the user perspective (direct + role-inherited).
- `GetResourceUsers(resourceID, resourceType)` — queries effective permissions from the resource perspective (direct + role-inherited).
- `CheckPermissionConsistency()` — currently returns a hardcoded `Consistent: true` stub in `resource_perm_repo.go:391-400`.
- Domain types `PermissionConsistencyResult` and `PermissionInconsistencyVO` are fully defined with `UserView`/`ResourceView` fields.

The data model involves 7 tables: `sys_user_perm`, `sys_role_perm`, `sys_user_role`, `sys_perm`, `sys_resource`, `sys_resource_perm`, `sys_user`.

## Goals / Non-Goals

**Goals:**
- Replace the always-true stub with a real cross-view consistency check.
- Detect divergences between user-perspective and resource-perspective effective permission sets.
- Populate `Inconsistencies` with concrete `userID, resourceID, resourceType, userView, resourceView` tuples.
- Keep the check read-only (no mutation).

**Non-Goals:**
- Auto-remediation of inconsistencies (detection only).
- Exposing a new HTTP endpoint (the service method already exists).
- Performance optimization for very large datasets (acceptable to scan all governed users in a single check).
- Caching or incremental checking (full scan on each invocation is acceptable for admin-only usage).

## Decisions

### D1: Cross-check strategy — set-based diff on (userID, permKey) pairs

For each governed resource type (dashboard, screen, dataset, datasource):
1. Collect the "user view" effective set: all `(userID, permKey)` pairs visible from `GetUserResources`.
2. Collect the "resource view" effective set: all `(userID, permKey)` pairs visible from `GetResourceUsers` across all registered resources.
3. Diff the two sets: entries present in one but not the other are inconsistencies.

**Rationale**: This reuses the existing `GetUserResources`/`GetResourceUsers` query patterns and the `sys_perm.perm_key` namespace for resource-type scoping. It avoids inventing a new query paradigm.

**Alternative considered**: Single mega-SQL with FULL OUTER JOIN — rejected because it would be harder to debug and maintain than two separate collection queries + in-memory diff.

### D2: Scope — only check governed resource types

Only the four governed resource types (`dashboard`, `screen`, `dataset`, `datasource`) are checked. This aligns with what `resourcePermKeyPrefix()` already covers.

### D3: Implementation layer — repository only

The implementation lives entirely in `ResourcePermissionRepository.CheckPermissionConsistency()`. No service-layer changes needed — the service already delegates to `s.repo.CheckPermissionConsistency()`.

### D4: Performance guard — bail out at 10,000 users

If `SELECT COUNT(*) FROM sys_user WHERE del_flag = 0` exceeds 10,000, the method returns `Consistent: true` with a note that the check was skipped for performance reasons. This prevents the admin-only endpoint from causing unacceptable load on very large deployments.

## Risks / Trade-offs

- **[Risk] N+1 query pattern**: Iterating over every user and calling `GetUserResources` individually could be slow. → **Mitigation**: Batch the query — collect all active user IDs in one query, then build the user-view set in a single SQL that joins `sys_user_perm` + `sys_role_perm` + `sys_user_role` + `sys_perm` grouped by user_id and perm_key prefix. Same for resource-view.
- **[Risk] Large result sets**: Many users × many permissions = large in-memory diff. → **Mitigation**: Cap at 10,000 users (D4). In-memory diff on sets of `(userID, permKey)` strings is bounded.
- **[Risk] No auto-fix**: Detecting inconsistencies without fixing them may feel incomplete. → **Mitigation**: Acceptable for now — the spec only requires detection. Admin can use existing save APIs to correct.

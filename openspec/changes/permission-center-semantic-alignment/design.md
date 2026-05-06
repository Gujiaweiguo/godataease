## Context

C1 `iam-foundation-semantics` and C2 `user-role-lifecycle-alignment` established organization-scoped governance for all IAM write paths. The permission center — covering menu, resource, row, column, and export permissions — still operates at global scope. This creates a gap where permission mutations can bypass org isolation boundaries established downstream.

**Current state**: The permission system has 5 service layers (`ResourcePermissionService`, `RowPermissionService`, `ColumnPermissionService`, `ExportPermissionService`, `DataPermissionAdminService`), 3 repositories, 2 middleware layers, and a compat handler bridge. None consume org context. Audit trail exists only in C2's role/lifecycle services.

**Constraints**: Must not break existing permission checks for admin users. Must align with C2's audit pattern. Must keep deferred dimensions stable (sysParams/whiteList already clarified in prior changes).

## Goals / Non-Goals

### Goals
- All governed permission write paths consume and validate org context
- Permission mutations produce audit entries consistent with C2 pattern
- Deferred dimensions tracked in a single registry with stable error codes
- Permission middleware resolves org context and validates membership
- `CheckPermissionConsistency` operates within org scope
- Admin bypass preserved (admin sees across orgs as today)

### Non-Goals
- Frontend permission-center UI redesign (future change)
- Implementing deferred dimensions (sysParams, whiteList, dept) — only tracking them
- Changing permission cache TTL or invalidation strategy
- Row/column permission SQL generation changes
- Resource group inheritance changes

## Decisions

### D1: Permission write paths adopt org-scoped context via existing `requireGovernedOrgContext()` pattern

**Choice**: Reuse the `requireGovernedOrgContext()` guard established in C1/C2 for all governed permission mutation entry points.

**Rationale**: C1 already defined this as the standard pattern for org-scoped governance. `ResourcePermissionService`, `DataPermissionAdminService`, and compat handlers will call `requireGovernedOrgContext()` before mutations.

**Alternatives considered**:
- Middleware-only org check (rejected: too coarse, doesn't give service-layer control for audit)
- New permission-specific org resolver (rejected: duplicates C1 pattern)

### D2: Permission audit uses C2's `auditPermissionMutation()` helper pattern

**Choice**: Extract a shared `auditPermissionMutation(op, actor, target, orgID, details)` helper into a common audit package, used by both C2 role services and C3 permission services.

**Rationale**: C2 already established the audit-entry schema (operation type, actor, target, org, timestamp). Using the same schema for permission mutations maintains audit consistency.

**Alternatives considered**:
- Separate audit table for permissions (rejected: fragments audit trail)
- Event-sourced audit log (rejected: over-engineering for current needs)

### D3: Centralized `DeferredDimensionRegistry` replaces scattered rejection messages

**Choice**: Create a `DeferredDimensionRegistry` in `domain/permission/` that maps dimension name → stable error code + message. Services query the registry instead of hardcoding rejection messages.

**Rationale**: Currently, sysParams rejection lives in `DataPermissionAdminService`, whiteList rejection lives in `SaveRowPermission`, and dept is silently ignored. Centralizing these into one registry with stable error codes (e.g., `DEFERRED_DIMENSION_SYS_PARAMS`, `DEFERRED_DIMENSION_WHITELIST`, `DEFERRED_DIMENSION_DEPT`) makes the contract testable and maintainable.

**Alternatives considered**:
- Keep scattered messages but make them consistent (rejected: still fragile)
- Configuration-driven registry in DB (rejected: over-engineering for 3 entries)

### D4: Permission middleware resolves org from JWT claims + membership validation

**Choice**: `PermissionMiddleware` and `RowPermissionMiddleware` extract org context from JWT claims (already populated by auth middleware), validate org membership via `userRoleRepo.IsUserInOrg()`, and fail closed (403) if membership cannot be confirmed.

**Rationale**: C2's `TransferUserOrg` and `AssignRolesToUser` already validate org membership. The permission middleware should use the same check. Admin users bypass org membership check.

**Alternatives considered**:
- Org header in every request (rejected: redundant with JWT claims)
- Org lookup from resource ownership (rejected: adds latency, circular dependency risk)

### D5: `CheckPermissionConsistency` filters by org scope

**Choice**: The consistency checker receives org ID as parameter, filters both user-view and resource-view queries by org membership, and reports inconsistencies only within the org boundary.

**Rationale**: Without org scoping, the checker would compare permissions across orgs, producing false positives when users legitimately have different permissions in different orgs.

**Alternatives considered**:
- Keep global scope with org-aware comparison (rejected: complex, error-prone)
- Per-org parallel checks (rejected: no current use case)

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Org-scoped guards break existing admin workflows | Admin bypass preserved; feature flag to revert to global scope |
| Middleware org lookup adds latency to every permission check | Cache org membership in Redis (already cached in C2 pattern) |
| Consistency checker scope change produces different results | New tests compare global vs org-scoped results; migration guide |
| DeferredDimensionRegistry becomes stale if dimensions are implemented | Registry is code-level (not DB), updated when dimensions ship |

## Migration Plan

1. Add `DeferredDimensionRegistry` (non-breaking, additive)
2. Add org-context extraction to permission middleware (feature-flagged)
3. Add org-scoped guards to permission services (feature-flagged)
4. Add audit trail to permission mutations (non-breaking, additive)
5. Scope consistency checker to org (feature-flagged)
6. Enable feature flags in staging, verify admin workflows intact
7. Enable feature flags in production
8. Remove feature flags after stabilization period

**Rollback**: Disable feature flags to revert to global-scope behavior. Audit entries and deferred registry remain (non-destructive additions).

## Open Questions

- Should the deferred dimension registry be exposed via an admin endpoint for frontend consumption? (Leaning yes for transparency, but not blocking C3)
- Should `ExportPermissionService` also get org-scoped guards in C3 or defer to a later change? (Leaning include in C3 since it's a thin layer over `ResourcePermissionService`)

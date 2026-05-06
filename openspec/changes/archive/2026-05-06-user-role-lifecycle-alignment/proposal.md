## Why

C1 (`iam-foundation-semantics`) froze the IAM governance foundation: org-context guards on governed writes, built-in role baselines, last-role BLOCK as an intentional deviation, org delete policy, and legacy compat route classification. Those contracts are in place on `main`, but the user-role lifecycle operations that depend on them still lack explicit org-scoped alignment. Three gaps remain open.

First, the last-role policy is frozen as BLOCK, which is a documented C1 deviation from the official spec that expects cascade user deletion. This deviation needs to evolve into a configurable, auditable policy rather than remaining a hardcoded reject.

Second, user-role assignment, removal, and transfer between organizations are undefined at the spec level. Current code handles individual operations, but there is no contract for org-scoped transactional assignment, idempotent role binding, or explicit user org membership transfer.

Third, C1 classified compat routes into three buckets. The FRONTEND MIGRATION bucket routes still use legacy paths without canonical-path mapping, so frontend migration progress cannot be tracked or verified.

## What Changes

- Evolve the last-role policy from BLOCK to a configurable safety policy with audit trail, so administrators can choose between BLOCK, WARN+ALLOW, or CASCADE based on organizational governance requirements.
- Define user-role assignment lifecycle with org-scoped transactions: assigning roles to users within an organization scope, removing roles with policy-governed safety checks, and idempotent binding guarantees.
- Define explicit user org membership transfer as a first-class operation with audit trail, rather than an implicit side effect of role manipulation.
- Enforce role-member operations (add/remove/query) verify org scope consistency before executing.
- Classify and map FRONTEND MIGRATION bucket routes to their canonical paths, providing a migration tracking surface for frontend teams.

## Capabilities

### New Capabilities
- `user-role-lifecycle`: User-role assignment lifecycle, membership transfer between organizations, last-role policy evolution from BLOCK to configurable governance, and org-scoped role-member operations.

### Modified Capabilities
- `user-management`: User-role operations surfaced in the user admin page must respect org-scoped lifecycle contracts, including assignment, removal, and transfer workflows.
- `role-management`: Member lifecycle operations must enforce org-scoped consistency and respect the evolved last-role configurable policy instead of hardcoded BLOCK.
- `organization-management`: Membership transfer and cross-org user movement become explicit governed operations with audit requirements.

## Impact

- **Backend service layer**: `role_service.go`, `user_service.go`, `org_service.go` gain new lifecycle methods and policy evaluation logic.
- **Backend handlers**: New REST endpoints for transfer and policy configuration; existing member-operation handlers gain org-scope validation.
- **Integration tests**: New test coverage for configurable last-role policy, org-scoped assignment transactions, and transfer operations.
- **API surface**: New endpoints for user org transfer and last-role policy query/update. Existing member-operation endpoints gain stricter org-scope validation.
- **Frontend**: User management and role management pages need to handle configurable last-role policy responses and org transfer workflows.
- **Compat routes**: FRONTEND MIGRATION bucket routes get canonical-path mapping metadata; no removal of PERMANENT SHIM routes.
- **Dependencies**: C1 `iam-foundation-semantics` must be fully merged. DB migration `20260307_alter_sys_role_for_official_spec.sql` must be applied.

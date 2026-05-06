## Context

C1 (`iam-foundation-semantics`) established three building blocks that C2 now extends. First, `requireGovernedOrgContext()` is the enforcement root for governed IAM writes, consumed by org, role, and user service methods. Second, `SysRole` carries `OrgID`, `IsBuiltin`, and `Readonly` fields that classify roles for downstream workflows. Third, `bindUserToOrgBaseline` is the authoritative user-org binding entry point. The last-role policy is frozen as BLOCK at `role_service.go` line 349, recorded as an intentional C1 deviation.

Current code handles individual role-member operations but lacks transactional org-scoped assignment, configurable last-role policy, and explicit org membership transfer. Compat routes classified as FRONTEND MIGRATION in C1 have no canonical-path mapping yet.

DB migration `20260307_alter_sys_role_for_official_spec.sql` must be applied before C2 code changes to align the schema with the Go model.

## Goals / Non-Goals

**Goals:**
- Evolve the last-role safety policy from hardcoded BLOCK to a configurable governance policy (BLOCK / WARN+ALLOW / CASCADE) with audit trail.
- Define user-role assignment as an org-scoped transactional operation with idempotent guarantees.
- Define user org membership transfer as an explicit, auditable operation (not implicit role removal side effect).
- Enforce that role-member add/remove/query operations verify org scope consistency before executing.
- Map FRONTEND MIGRATION bucket compat routes to canonical paths with tracking metadata.

**Non-Goals:**
- Do not redesign menu/resource/row/column permission enforcement (belongs to C3 `permission-center-semantic-alignment`).
- Do not remove PERMANENT SHIM compat routes (frozen as permanent in C1).
- Do not redesign DUAL-SUPPORT TRANSITION routes (C3+ territory).
- Do not introduce UI redesign beyond what the lifecycle workflows require.
- Do not alter the org delete policy (already frozen in C1).

## Decisions

### 1. Last-role policy evolves from BLOCK to configurable three-mode policy with audit

The last-role safety check at `role_service.go` UnmountUser will read a configurable policy instead of always rejecting. Three modes:

- **BLOCK** (default): Reject removal. Current C1 behavior, no breaking change on upgrade.
- **WARN+ALLOW**: Allow removal after recording a warning audit entry. User keeps account with zero roles.
- **CASCADE**: Remove role and disable the user account. User account stays in DB with `enabled = false`, not hard-deleted.

Each policy evaluation writes an audit log entry capturing the policy mode, actor, target user, affected role, and timestamp.

**Alternative considered:** immediate CASCADE-only adoption.
**Why rejected:** CASCADE deletes carry risk without a UI confirmation flow. Three-mode policy lets organizations choose their risk tolerance. Default BLOCK preserves backward compatibility.

### 2. User-role assignment uses org-scoped transactional binding

Assigning roles to a user within an org scope is a single transactional operation. The service method:
1. Validates org context via `requireGovernedOrgContext()`.
2. Verifies target user exists in the same org scope (or is being added to it).
3. Creates user-role associations idempotently (duplicate bindings are no-ops, not errors).
4. Records audit entry with assignment metadata.

If the user is not yet a member of the target org, the assignment transaction also creates the org membership baseline via `bindUserToOrgBaseline`.

**Alternative considered:** separate API calls for membership binding and role assignment.
**Why rejected:** splitting the operation creates partial-state windows where membership exists without roles or roles are assigned to non-members. Transactional binding eliminates this class of bug.

### 3. User org transfer is an explicit operation with two-phase semantics

Transferring a user from org A to org B is an explicit `TransferUserOrg` operation, not an implicit side effect of removing roles in A and adding roles in B. The operation:

1. Validates that the target org B exists and is active.
2. Within a single DB transaction: removes user-role bindings from org A, creates baseline membership in org B, and assigns a default role in B (determined by the target org's built-in baseline).
3. Records a transfer audit entry capturing source org, target org, user, actor, and resulting role assignment.

The user account itself is not duplicated or moved. Only the org-scoped membership and role bindings change.

**Alternative considered:** implicit transfer via role manipulation.
**Why rejected:** implicit transfer has no audit trail, no atomicity guarantee, and no clear policy for what happens when removal in org A triggers last-role. Explicit transfer avoids all three issues.

### 4. Role-member operations verify org scope consistency

Every role-member add, remove, and query operation checks that the role and the target user share the same org scope before proceeding. This check uses the same `requireGovernedOrgContext()` enforcement root from C1.

For cross-org user discovery (e.g., adding an external user to a role), the operation validates that the resulting association is created within the target org scope, even if the user has memberships in other orgs.

**Alternative considered:** allow cross-org role-member associations.
**Why rejected:** cross-org associations violate the org-scoped isolation contract established in C1 and make permission inheritance unpredictable.

### 5. FRONTEND MIGRATION bucket routes get canonical-path mapping metadata

C1 classified compat routes into three buckets. For the FRONTEND MIGRATION bucket, C2 adds a mapping table in the compat route registration that records:

- Legacy path (e.g., `/de2api/system/role/update`)
- Canonical path (e.g., `/api/role/edit`)
- Migration status: `pending`, `in-progress`, `migrated`

This mapping is exposed as a read-only admin endpoint for frontend teams to track migration progress. It does not change routing behavior; compat routes continue to delegate to canonical handlers as before.

**Alternative considered:** remove FRONTEND MIGRATION routes immediately.
**Why rejected:** C1 explicitly deferred route removal to avoid breaking frontend callers. Mapping metadata provides visibility without disruption.

## Risks / Trade-offs

- **[Risk] Configurable last-role policy adds runtime configuration surface** → **Mitigation:** default is BLOCK (current behavior). Policy is stored in a simple key-value governance config table with validation constraints. Only org-level admins can change the policy.

- **[Risk] User org transfer is a high-impact operation that could orphan access** → **Mitigation:** transfer is auditable, requires target-org admin authority, and assigns a default role in the target org. Source-org role removal follows the configured last-role policy.

- **[Risk] Org-scoped transactional binding may introduce deadlocks under concurrent assignment** → **Mitigation:** use row-level locking on the user-role junction table within the transaction scope. Keep transactions short (validate, insert, audit).

- **[Risk] Compat route mapping metadata could become stale** → **Mitigation:** mapping status is updated as part of the same change set that migrates a frontend call. CI can verify mapping status matches actual frontend usage.

- **[Risk] CASCADE mode disables user accounts, which may surprise administrators** → **Mitigation:** CASCADE mode audit entries are prominently tagged. UI confirmation flow (C2 frontend work) must display the consequence before allowing the operation.

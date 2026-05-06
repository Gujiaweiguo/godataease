## Context

The governance gap plan has already frozen four C1 policy decisions that must become the implementation baseline for all later IAM work: organization isolation, built-in role baseline, last-role safety policy, organization delete policy, and legacy compatibility routing policy. Today these semantics are spread across `auth_service.go`, `org_service.go`, `role_service.go`, and multiple compatibility handlers, so later user/role and permission work could easily drift if C1 is not captured as a single contract first.

Current code already contains partial implementation: token bootstrap and org switching exist, last-role removal is blocked in `role_service.go`, organization deletion uses child rejection plus soft delete, and compat routes are registered through multiple bridge handlers. The missing piece is not raw capability, but a unified design that defines which semantics are authoritative, which deviations are intentional, and where later changes are allowed to extend behavior.

## Goals / Non-Goals

**Goals:**
- Make organization context the authoritative runtime contract for governed IAM write paths.
- Freeze the intentional deviation that last-role removal is blocked rather than cascading user deletion.
- Freeze organization delete semantics as child rejection + soft delete + auditable deferred resource disposition.
- Classify legacy IAM compatibility routes into permanent shim, frontend migration, and dual-support transition buckets.
- Provide downstream changes (`user-role-lifecycle-alignment`, `permission-center-semantic-alignment`) with a stable prerequisite contract.

**Non-Goals:**
- Do not redesign menu/resource/row/column permission enforcement details in this change.
- Do not complete full user/role lifecycle alignment (that belongs to C2).
- Do not remove all compatibility routes immediately.
- Do not introduce UI redesign or non-governance refactors.

## Decisions

### 1. Foundation semantics are captured by modifying existing governance capabilities, not by inventing a new parallel IAM spec
C1 changes behavior already covered by `organization-management`, `role-management`, `user-management`, and `api-compatibility-bridge`. Reusing those capability specs keeps archive history attached to the modules that will actually enforce the policies.

**Alternative considered:** create a brand-new `iam-foundation-semantics` capability.
**Why rejected:** it would duplicate requirements that already live in org/role/user/compat specs and make later implementation traceability worse.

### 2. Organization context is the enforcement root for governed IAM writes
Governed create/update/delete flows for organizations, users, and roles must consume the active runtime organization context rather than relying only on token bootstrap side effects or caller discipline. This narrows future implementation work to deterministic service-layer scope checks.

**Alternative considered:** keep org isolation at token/bootstrap level only.
**Why rejected:** that leaves service-layer writes vulnerable to cross-org drift and contradicts the frozen policy lock.

### 3. Last-role policy is recorded as an intentional deviation
The official manual implies deleting a user when its last role is removed, but the governed baseline keeps the current safer behavior: reject removal of the final role. The spec must say this explicitly so later C2 work treats it as a documented policy, not as an accidental mismatch.

**Alternative considered:** change spec to require cascade delete immediately.
**Why rejected:** it is high-risk, lacks current UI confirmation flow, and would expand C1 beyond minimal governance locking.

### 4. Organization delete policy is split into immediate deterministic behavior plus deferred resource disposition
C1 only freezes the allowed immediate behavior: reject deletion when children exist, otherwise soft-delete the organization and preserve auditable resource disposition state. Resource cleanup remains a deferred downstream obligation, not an unresolved ambiguity.

**Alternative considered:** require full cascade disposal in C1.
**Why rejected:** it couples foundation locking to cross-domain cleanup work that belongs to C2.

### 5. Compatibility routes are governed by route-family policy buckets
C1 classifies legacy routes instead of removing them wholesale: some stay as permanent shims, some are explicit frontend migration targets, and some remain dual-support during transition. This keeps later route cleanup reviewable and prevents compat fixes from being mistaken for semantic completion.

**Alternative considered:** declare all compat routes temporary and remove them aggressively.
**Why rejected:** current frontend and external callers still depend on several route families, and abrupt removal would make downstream verification noisy and misleading.

## Risks / Trade-offs

- **[Risk] Organization-scope rules may be interpreted inconsistently across services** → **Mitigation:** express them in modified org/role/user specs with explicit governed-write scenarios.
- **[Risk] Intentional deviation on last-role behavior may confuse future implementers** → **Mitigation:** encode the deviation directly in the role-management delta and call out why cascade delete is rejected for C1.
- **[Risk] Deferred organization resource disposal could be mistaken for complete parity** → **Mitigation:** specs explicitly distinguish immediate soft-delete semantics from downstream disposal work.
- **[Risk] Compatibility-route buckets may become stale as frontend migration advances** → **Mitigation:** capture them as governance rules in `api-compatibility-bridge` so later changes must update the spec when route families move between buckets.

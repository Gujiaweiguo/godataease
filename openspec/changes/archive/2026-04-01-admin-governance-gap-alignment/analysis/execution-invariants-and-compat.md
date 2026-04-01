# Admin Governance Alignment Invariants and Compatibility Policy

## T3. Shared Invariants

### Invariant I1 — Organization Context Is First-Class
- All governed user, role, and permission workflows must consume an explicit organization context.
- Search, assignment, and authorization behavior must remain scoped to that context.

### Invariant I2 — Organization Isolation Must Stay Intact
- Same-level and parent/child organizations do not implicitly inherit each other's governed resources.
- Any foundation change that affects queries, assignment, or authorization must preserve isolation semantics.

### Invariant I3 — Built-In Role Baseline Must Be Stable
- System-level immutable roles and organization-scoped built-in roles must remain distinguishable.
- Custom roles must inherit only from allowed built-in baselines.

### Invariant I4 — Last-Role Policy Must Be Deterministic
- The system must not leave “last role removal” as an accidental side effect.
- The fork must explicitly choose and test one policy before lifecycle work can complete.

### Invariant I5 — Organization Delete Policy Must Be Deterministic
- Child-organization preconditions and resource-disposition behavior must be defined together.
- Deleting a leaf organization must produce an auditable, deterministic result.

### Invariant I6 — Effective Permission Must Converge Across Views
- Menu/resource/data permissions must resolve to one effective authorization state.
- By-user and by-resource views cannot diverge semantically.

### Invariant I7 — Authorization Denial Must Be Distinguishable From Missing Resource
- Permission denials must not degrade into generic 404 or placeholder-success behavior.
- This applies to menu routes, permission APIs, and governed BI resource flows.

## Final Change Split

### Recommended 3-Change Semantic Split
1. `iam-foundation-semantics`
2. `user-role-lifecycle-alignment`
3. `permission-center-semantic-alignment`

### Why the 4-Change Default Is Rejected
- `org-management-foundation` and `user-role-center` share the same org-scope and role-baseline invariants.
- `permission-menu-resource` and `permission-row-column` share the same effective-permission model and should not diverge before foundation semantics stabilize.
- Page adjacency is weaker than semantic coupling for rollback and regression control.

## T4. Compatibility Policy

### Canonical Route Principle
- Canonical Go-aligned routes are preferred when both canonical and compatibility aliases exist.
- Compatibility aliases remain only where frontend migration has not yet finished and governed regression evidence still depends on them.

### Route Family Decisions

#### User legacy routes
- Affected examples from `apps/frontend/src/api/user.ts`:
  - `/user/pager/:page/:limit`
  - `/user/create`
  - `/user/edit`
  - `/user/delete/:uid`
  - `/user/enable`
- Policy:
  - treat as compatibility surface, not canonical source of truth
  - prefer frontend migration to canonical `/system/user/*` or governed current-org paths
  - add backend alias only when a governed flow still depends on the legacy path and migration cannot be completed safely in the same wave

#### Permission compatibility routes
- Affected examples from `apps/frontend/src/api/auth.ts` and `permission_compat_handler.go`:
  - `/auth/menuPermission`
  - `/auth/busiPermission`
  - `/auth/userPerspective`
  - `/auth/busiTargetPermission`
  - `/auth/menuTargetPermission`
  - `/auth/saveMenuPer`
  - `/auth/saveBusiPer`
  - `/auth/saveBusiTargetPer`
  - `/auth/saveMenuTargetPer`
- Policy:
  - keep existing compat routes only where they are already functionally complete
  - incomplete target-style routes must be explicitly classified as gaps and cannot be treated as parity-complete
  - new implementation should prefer canonical governed permission services behind both views

#### Menu target routes
- `menuTargetPerApi` and `menuTargetPerSaveApi` remain explicit governed gaps.
- Policy:
  - do not silently keep them as pseudo-supported paths
  - either implement working canonical-backed behavior or retire/migrate the caller with explicit decision record

#### `/user/org/option`
- Policy:
  - classify as unresolved compatibility gap
  - resolve by one of three explicit choices only: add alias, migrate frontend caller, or retire usage

## Verification Order for Early Waves
1. baseline freeze and gap matrix
2. invariant lock and compatibility policy
3. foundation policy decisions
4. lifecycle alignment
5. permission-center convergence

## Blocking Decisions for T5+
- Decide whether this fork keeps “block last-role removal” or switches to “remove user on last-role removal”.
- Decide whether organization delete uses soft-delete with governed resource disposition, synchronous cascade, or asynchronous cleanup.
- Decide whether unsupported third-party user-source fields are deferred or minimally added in the lifecycle wave.

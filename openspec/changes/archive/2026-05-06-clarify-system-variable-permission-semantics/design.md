## Context

The permission-center work completed in the governance gap plan intentionally converged the governed surface around menu permissions, resource permissions, and row/column permissions. The current runtime and admin save paths already enforce that row permissions only support `user` and `role` targets in T8 (`apps/backend-go/internal/service/data_permission_admin_service.go`), but several permission-facing contracts still carry residual system-variable semantics:

- backend permission DTO comments still describe `authTargetType` as `user, role, sysParams` (`apps/backend-go/internal/domain/permission/row_permission.go`)
- frontend filter-tree components still branch on `authTargetType === 'sysParams'` and expose `${sysParams.roles}`-style options in permission-adjacent flows (`apps/frontend/src/views/visualized/data/dataset/auth-tree/FilterFiled.vue` and sibling components)
- system variable management APIs remain valid for variable definition/value CRUD (`apps/frontend/src/api/variable.ts`, backend `SystemVariableHandler` / `SystemVariableService`), which is correct for `system-variable-management`, but those APIs do not imply permission-center authorization support

This creates a contract gap: some code paths still look like system-variable permission assignment is partially supported even though the governed T8 implementation explicitly deferred it.

## Goals / Non-Goals

**Goals:**
- Make the deferred status of system-variable permission assignment explicit and consistent across permission-facing backend contracts, validation, and frontend UI copy/affordances.
- Preserve system variable definition/value management as a separate supported capability.
- Ensure unsupported system-variable permission payloads and UI paths fail explicitly or disappear instead of looking partially implemented.
- Keep the change small enough to merge independently without reopening broader T8/P2 enforcement work.

**Non-Goals:**
- Do not implement system-variable permission assignment semantics.
- Do not introduce new permission targets beyond the current governed `user` / `role` scope.
- Do not expand into whitelist persistence, broader P2 inheritance work, or `RowPermissionMiddleware` runtime replacement.
- Do not change the existing supported `/sysVariable/*` CRUD behavior for managing variable definitions and values.

## Decisions

### 1. Treat this as a permission-contract clarification, not a new capability

**Decision:** Modify `permission-config` and `system-variable-management` specs instead of adding a new capability.

**Rationale:** The problem is semantic overlap between two existing capabilities: permission configuration and system variable CRUD. A new capability would falsely suggest that system-variable permission assignment is being introduced, when the real goal is to document and enforce that it is still deferred.

**Alternative considered:** Add a new `system-variable-permission-management` capability. Rejected because it would create a misleading product contract before the feature exists.

### 2. Fail explicit beats silent compatibility for unsupported permission targets

**Decision:** Permission-facing save/query contracts should explicitly reject or hide system-variable authorization targets instead of tolerating residual `sysParams` shapes.

**Rationale:** The current admin save flow already rejects unsupported row permission targets and whitelist semantics. Extending that explicitness to the remaining contracts keeps the permission center honest and removes “looks supported” ambiguity.

**Alternative considered:** Keep accepting legacy-looking payloads but document them as no-op. Rejected because silent acceptance is exactly what makes the deferred state misleading.

### 3. Separate management CRUD from permission assignment semantics

**Decision:** Keep `/sysVariable/*` management APIs and related frontend CRUD screens untouched unless they directly imply permission-center support.

**Rationale:** `system-variable-management` is an existing supported capability for variable definitions and values. The change should only remove or relabel permission-assignment implications, not destabilize valid variable management workflows.

**Alternative considered:** Collapse or disable system variable management APIs together with permission clarification. Rejected because variable CRUD is already a legitimate supported feature and is not the source of ambiguity.

### 4. Prefer UI affordance removal or explicit unsupported messaging over hidden partial branching

**Decision:** In frontend permission-adjacent components that still branch on `sysParams`, either remove the branch from governed permission flows or replace it with an explicit unsupported state/message.

**Rationale:** A branch that remains reachable but non-functional creates more support burden than either hiding it or making it intentionally unavailable. Which option to apply can vary by component, but the user-visible outcome must be unambiguous.

**Alternative considered:** Leave the branch in place and rely on backend rejection. Rejected because the ambiguity starts in the UI and should be resolved before the request is sent.

### 5. No database or cache migration in this change

**Decision:** Keep the change at contract, validation, and UI-affordance level only.

**Rationale:** No new persistence model is being introduced. The risk lies in misleading semantics, not schema shape.

**Alternative considered:** Add compatibility markers or migration data to track unsupported historical rows. Rejected because current plan evidence treats the topic as deferred, not partially migrated.

## Risks / Trade-offs

- **[Risk] Hidden dependency on ambiguous `sysParams` behavior in internal workflows** → **Mitigation:** tighten behavior with explicit unsupported responses/messages, and keep rollback limited to UI/contract clarification if a real dependent workflow appears.
- **[Risk] Frontend and backend may become inconsistent if only one side is clarified** → **Mitigation:** update both permission-facing UI affordances and backend validation/contracts in the same implementation wave.
- **[Risk] Overreach into broader T8/P2 work** → **Mitigation:** treat middleware replacement, whitelist persistence, and runtime enforcement expansion as explicit non-goals in specs/tasks.
- **[Risk] Readers may confuse system variable CRUD with system-variable permission support even after changes** → **Mitigation:** spec deltas must explicitly distinguish “supported variable management” from “unsupported permission assignment semantics.”

## Migration Plan

1. Update spec deltas for `permission-config` and `system-variable-management` to encode the clarified boundary.
2. Implement backend permission-contract cleanup and validation alignment for unsupported system-variable permission targets.
3. Implement frontend affordance cleanup / explicit unsupported messaging in permission-facing flows that still surface `sysParams` semantics.
4. Add focused tests covering explicit unsupported behavior and preservation of supported system-variable CRUD behavior.
5. Rollback plan: revert the clarification changes if they block a real workflow, but keep the deferred boundary documented so follow-up work does not regress into ambiguous semantics.

## Open Questions

- Which frontend permission-adjacent components with `sysParams` branches are still reachable from governed user flows versus only from editor-internal or future-deferred paths?
- Should unsupported system-variable permission requests return a dedicated domain error/message, or reuse the existing unsupported-target validation shape used for other deferred targets?
- Do any existing docs or translation keys outside OpenSpec need explicit “not supported in current permission center” wording to prevent operator confusion?

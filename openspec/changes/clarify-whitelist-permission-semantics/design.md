## Context

The governed permission-center work has already made whitelist behavior effectively unsupported in the unified row/column permission flow, but the contract still looks ambiguous to consumers:

- backend write validation still accepts a `whiteList` field in `RowPermissionForm` and then rejects non-empty values with the internal-planning message `whiteList is not supported in T8` (`apps/backend-go/internal/service/data_permission_admin_service.go`)
- backend read models still serialize `whiteListUser`, `whiteListRole`, and `whiteListDept` in `RowPermissionDTO` / `DataPermRow`, while `buildRowPermissionPage` returns an always-empty `WhiteList: []int64{}` (`apps/backend-go/internal/domain/permission/row_permission.go`, `apps/backend-go/internal/domain/permission/data_perm_row.go`, `apps/backend-go/internal/service/data_permission_admin_service.go`)
- the unified permission center already strips `whiteList` before submit and does not expose whitelist editing UI (`apps/frontend/src/views/system/permission/DataPermission.vue`)
- legacy dataset auth-tree code still contains adjacent deferred auth-target semantics (`apps/frontend/src/views/visualized/data/dataset/auth-tree/FilterFiled.vue`), which is a boundary to document but not to normalize in this change

The main `permission-config` spec now requires deferred dimensions to be traceable or explicitly marked as deferred instead of being silently accepted. This change exists to make whitelist semantics meet that bar without accidentally turning deferred persistence into a supported feature.

## Goals / Non-Goals

**Goals:**
- Replace milestone-coupled whitelist rejection text with stable deferred-semantics contract language.
- Clarify the backend read/write contract so whitelist-related fields no longer imply supported persistence or editing semantics.
- Preserve the current unified permission-center UX boundary where whitelist editing is unavailable.
- Record the intentional split between governed permission-center behavior and legacy/deferred adjacent flows.

**Non-Goals:**
- Do not implement whitelist persistence or runtime whitelist enforcement.
- Do not remove legacy/deferred dataset auth-tree `sysParams` branches as part of this change.
- Do not redesign row/column permission data models beyond what is needed for deferred-contract clarity.
- Do not touch middleware replacement, broader P2 inheritance work, or other deferred permission-center expansion items.

## Decisions

### 1. Keep read-path compatibility; clarify semantics instead of removing fields now

**Decision:** Preserve the existing read-path field shape for whitelist-related response fields in this change, and clarify their deferred meaning through spec contract, service-level shaping, and tests rather than removing them outright.

**Rationale:** `whiteListUser`, `whiteListRole`, `whiteListDept`, and `WhiteList` are already visible in response payloads. Removing them in the same change would be a breaking contract cleanup and would make the proposal's deferred-clarification goal compete with compatibility risk. A compatibility-preserving clarification is the smallest safe step.

**Alternative considered:** Remove whitelist-related read fields entirely. Rejected because it introduces API compatibility risk before the contract has been explicitly stabilized and communicated.

### 2. Use stable deferred error language on write paths

**Decision:** Replace `whiteList is not supported in T8` with a stable user-facing/service-facing error that identifies whitelist semantics as deferred in the permission center without referencing internal milestone labels.

**Rationale:** `T8` is an implementation-plan label, not an API contract. Consumers need an error that remains valid after planning waves change. The pattern established by the recent `sysParams is deferred and not supported in permission center` message is the right direction.

**Alternative considered:** Keep the current message and only document it. Rejected because the message itself is the ambiguity.

### 3. Treat whitelist fields as deferred contract surface, not supported persistence slots

**Decision:** Backend domain/service code should continue to expose whitelist-related fields only as inert deferred surface in this change; tests and spec wording should make it explicit that non-empty whitelist persistence/editing is unsupported.

**Rationale:** `DataPermRow` uses `gorm:"-"` for whitelist fields, and the unified permission center already strips `whiteList` before submission. This means the real behavior today is “field may exist, persistence is unsupported.” The design should codify that rather than pretending the fields are active storage.

**Alternative considered:** Retrofit persistence columns or shadow storage now. Rejected because that would be implementing whitelist support, not clarifying its deferred semantics.

### 4. Preserve the current unified permission-center UI boundary

**Decision:** The unified permission-center UI should continue not to submit or edit whitelist state; any frontend changes in this change should be limited to wording, error presentation, or regression coverage needed to make that boundary explicit.

**Rationale:** `DataPermission.vue` already removes `whiteList` from form submission and no longer offers whitelist editing. That is the correct governed behavior for the current phase and should not be reopened.

**Alternative considered:** Reintroduce read-only whitelist widgets in the unified permission center. Rejected because it risks implying partial support.

### 5. Document legacy auth-tree behavior as intentional boundary, not implementation target

**Decision:** Legacy dataset auth-tree components that still expose adjacent deferred semantics are documented as out of scope for normalization in this change unless they directly contradict the clarified whitelist contract.

**Rationale:** The proposal needs to acknowledge the split so readers do not assume the omission is accidental, but the smallest safe change is to clarify the governed permission-center contract first.

**Alternative considered:** Normalize all legacy auth-tree semantics together with whitelist clarification. Rejected because it broadens scope into a larger migration stream.

## Risks / Trade-offs

- **[Risk] Consumers may still misread preserved read-path whitelist fields as active support** → **Mitigation:** pair compatibility-preserving fields with stable deferred write errors, spec updates, and regression tests that explicitly encode unsupported persistence.
- **[Risk] Changing the whitelist error string can break brittle tests or downstream assertions** → **Mitigation:** update backend tests together with the message change and keep the new wording stable and contract-oriented.
- **[Risk] Legacy/deferred frontend paths remain visually inconsistent with the governed permission center** → **Mitigation:** record the split explicitly in specs/design and keep this change scoped to the governed contract rather than silently expanding into legacy cleanup.
- **[Risk] Future implementers may treat preserved fields as endorsement for persistence work** → **Mitigation:** make the deferred status explicit in design/spec/tasks so any future whitelist implementation must happen in a separate change.

## Migration Plan

1. Update the `permission-config` delta spec to require stable deferred whitelist semantics in read/write contracts.
2. Update backend service/domain behavior just enough to remove internal-plan wording and make deferred whitelist handling explicit in validation and response expectations.
3. Add or update backend tests to lock the new stable rejection behavior.
4. Add frontend regression coverage only where needed to prove the unified permission center still does not offer whitelist editing.
5. Rollback, if needed, by reverting the contract/message clarification while keeping the deferred whitelist feature unsupported; do not partially reintroduce editable whitelist behavior.

## Open Questions

- Should the stable deferred whitelist message mirror the recent `sysParams ... is deferred and not supported in permission center` wording exactly, or should whitelist use a separate but parallel phrase?
- Do we want to keep returning an always-empty `WhiteList: []` in row-permission page payloads, or should the service omit that field only when safe to do so without breaking existing consumers?
- Is there any user-visible documentation page outside OpenSpec that should be updated once the stable whitelist wording is finalized?

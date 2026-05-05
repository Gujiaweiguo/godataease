## Why

Row/column whitelist dimensions are deferred to P2, but the current codebase doesn't treat them as explicitly deferred. The backend still serializes `whiteListUser`, `whiteListRole`, `whiteListDept` fields in domain entities and returns an always-empty `whiteList` array in read responses. On the write path, submitting a non-empty `whiteList` produces `"whiteList is not supported in T8"`, an opaque rejection referencing an internal milestone label. The existing permission-config spec already requires that deferred dimensions be "traceable or explicitly marked as deferred instead of being silently accepted," but the implementation doesn't meet that bar. This gap creates confusion for consumers who see whitelist fields in payloads without knowing they're intentionally unsupported.

## What Changes

- Define an explicit contract for whitelist-related fields in row permission read/write paths: either omit them from responses or annotate them as deferred with a stable, consumer-facing error code.
- Replace the milestone-coupled rejection message (`"whiteList is not supported in T8"`) with a contract-stable message that identifies the feature as deferred without referencing internal planning labels.
- Align the `RowPermissionDTO` and `DataPermRow` domain model exposure of `whiteListUser`, `whiteListRole`, `whiteListDept` fields with the deferred-state contract (currently `gorm:"-"` computed fields serialized as empty strings in every response).
- Clarify that the legacy dataset auth-tree UI (`FilterFiled.vue` in `views/visualized/data/dataset/`) still surfaces `sysParams` filter options for backward compatibility, while the unified permission center (`DataPermission.vue`) correctly blocks `variable` filterType edits and strips `whiteList` before submission. Document this split as a known intentional boundary, not an accidental omission.

## Capabilities

### New Capabilities

None. This change clarifies and formalizes deferred semantics within an existing capability; it does not introduce new functionality.

### Modified Capabilities

- `permission-config`: Add explicit requirement-level scenarios for how whitelist and system-variable dimensions are handled in row/column permission read/write contracts. The existing spec already mandates that deferred dimensions be traceable or marked as deferred (see "Row and Column Permissions Must Enforce Governed Runtime Behavior" scenario). This change adds concrete contract requirements for the deferred-state handling, covering field exposure in read responses, rejection semantics on write, and error message stability. No change to runtime enforcement behavior for supported dimensions (user, role targets).

## Impact

- **Backend domain layer**: `RowPermissionDTO`, `DataPermRow` structs in `domain/permission/` expose whitelist-related JSON fields. Contract change may adjust or annotate these fields.
- **Backend service layer**: `data_permission_admin_service.go` `SaveRowPermission` validation message and `buildRowPermissionPage` response shape. The `unsupportedRowPermissionTargetTypeError` helper already distinguishes `sysParams` from other unsupported types.
- **Frontend permission center**: `DataPermission.vue` already strips `whiteList` and blocks `variable` filterType. No behavioral change expected, but contract documentation ensures this stays consistent.
- **Frontend legacy dataset auth-tree**: `FilterFiled.vue` and `options.js` still reference `sysParams` enums. Documented as an intentional legacy boundary, not a defect.
- **Tests**: Existing unit test in `data_permission_admin_service_test.go` asserts the `"whiteList is not supported in T8"` message. That assertion will need updating to match the new stable rejection message.
- **API compatibility note**: Removing read-path whitelist fields would be breaking, so the preferred direction for this change is to clarify deferred semantics through stable contract language and rejection behavior unless a later change explicitly approves a breaking contract cleanup.

## Context
This change is intentionally scoped to the MVP implementation: row-level permissions
with minimal UI integration. The full design for user, organization, and permission
management is tracked in `openspec/changes/add-user-org-permission-full/`.

## MVP Scope
- Row-level permission filtering (DataPermRow, RowPermissionFilter)
- Minimal row permission configuration UI (existing XpackComponent integration)
- Basic row permission API surface (dataset.ts integration)

## Deferred Scope
- User, organization, and role management
- Menu/resource/export permissions and inheritance
- Column-level permissions (disable/mask)
- Comprehensive tests, documentation, and API docs

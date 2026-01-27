# Change: Complete user, organization, and permission management

## Why
The MVP delivered row-level permissions with minimal UI, but the open-source edition
still lacks complete user, organization, role, and permission management needed for
multi-tenant use and fine-grained access control. This change completes the remaining
capabilities (menu/resource/column permissions, inheritance, and management UI) so
the system is production-ready for enterprise permission needs.

## What Changes
- Implement the remaining user, organization, and role management services and UI.
- Add menu/resource/export permissions with inheritance and assignment workflows.
- Add column-level permission masking/disable and configuration UI.
- Add organization switching and organization-scoped data isolation enforcement.
- Extend permission checks and query engine integration for full-scope permissions.
- Add tests, docs, and API documentation for the complete permission stack.

## Scope Notes
- Assumes MVP row-level permissions are already implemented in `add-user-org-permission`.
- Does not include SSO, audit logging, or permission templates.

## Impact
- Affected specs: user-management, organization-management, permission-config
- Affected code:
  - `core/core-backend/src/main/java/io/dataease/system/` (services, controllers, permission checks)
  - `core/core-frontend/src/views/system/` (management pages and permission UI)
  - `core/core-backend/src/main/resources/db/migration/` (schema and seed data)
- Depends on: `add-user-org-permission` (MVP row-level permissions)

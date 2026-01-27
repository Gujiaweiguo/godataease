# Change: Add User, Organization, and Permission Management

## Why
DataEase currently lacks comprehensive user management, organization management, and permission configuration capabilities in the open-source version. These features are essential for:
- Multi-tenant scenarios where different organizations need isolated data access
- Fine-grained access control to secure sensitive data and operations
- Role-based access control (RBAC) to manage user permissions efficiently
- Data row/column level permissions for compliance with data security requirements

## What Changes
- Add User Management: CRUD operations for users, user profile management
- Add Organization Management: Multi-organization support, organization hierarchy
- Add Permission Configuration:
  - Menu permissions: control access to functional modules
  - Resource permissions: control access to datasources, datasets, dashboards, screens
  - Data permissions: row-level and column-level access control
- Add Role Management: create and manage roles with associated permissions
- Implement permission inheritance and whitelist mechanisms

## Impact
- Affected specs: New capabilities (user-management, organization-management, permission-config)
- Affected code:
  - `core/core-backend/src/main/java/io/dataease/system/` - New modules
  - `core/core-frontend/src/views/system/` - New UI pages
  - `core/core-backend/src/main/resources/db/migration/` - Database schema changes
- Security: Significantly improves data security and access control

## Approval Decision
**Decision**: Approved for MVP implementation

**MVP Scope**:
- Row-level permissions only
- Basic permission configuration API
- Minimal UI for row permission configuration (existing XpackComponent integration)
- Basic integration tests and performance baseline

**Implementation Status**: Completed 2026-01-25
- Database migration: V2.10.21__data_perm_row.sql ✓
- Backend entities, services, controllers: DataPermRow, DataPermRowService, RowPermissionController ✓
- Row permission filter: RowPermissionFilter.java ✓
- Frontend integration: Existing XpackComponent + dataset.ts APIs ✓
- Compilation verified: Maven compile success ✓

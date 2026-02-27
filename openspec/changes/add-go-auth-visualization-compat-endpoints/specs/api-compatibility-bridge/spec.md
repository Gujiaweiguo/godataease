## ADDED Requirements

### Requirement: Permission Compatibility Endpoint Family Coverage
The compatibility bridge SHALL expose permission-management endpoints required by frontend role/menu administration flows.

#### Scenario: Permission endpoints are routable through compatibility prefix
- **WHEN** clients call legacy permission endpoints under compatibility path (for example `/de2api/auth/menuPermission`)
- **THEN** Go backend MUST route the request to implemented handlers instead of returning `404`
- **AND** response semantics remain compatible with frontend caller expectations

### Requirement: System Role Path Compatibility
The compatibility bridge SHALL support legacy `/system/role/*` API paths used by frontend admin pages.

#### Scenario: Create/update/delete role through legacy system path
- **WHEN** a client calls `/de2api/system/role/create`, `/de2api/system/role/update`, or `/de2api/system/role/delete/:roleId`
- **THEN** the backend MUST map these requests to canonical Go role operations
- **AND** return Java-compatible response envelope

## ADDED Requirements

### Requirement: Role-Menu Authorization Model
The system SHALL maintain role-to-menu authorization mappings in `sys_role_menu` table.

#### Scenario: Grant menu to role
- **WHEN** admin assigns a menu to a role
- **THEN** system creates role-menu mapping record
- **AND** users with that role gain visibility to the menu

#### Scenario: Revoke menu from role
- **WHEN** admin removes menu assignment from role
- **THEN** system deletes role-menu mapping
- **AND** users with that role lose menu visibility on next query

#### Scenario: Query role-menu assignments
- **WHEN** admin requests menus for a role
- **THEN** system returns list of assigned menu IDs
- **AND** includes menu metadata for tree selection UI

### Requirement: Role-Menu Authorization APIs
The system SHALL provide APIs to manage role-menu authorization.

#### Scenario: Query role menus endpoint
- **WHEN** GET /api/role/menu/:roleId is called
- **THEN** returns list of menu IDs assigned to role

#### Scenario: Save role menus endpoint
- **WHEN** POST /api/role/menu/save is called with roleId and menuIds array
- **THEN** system replaces all role-menu mappings idempotently
- **AND** returns success with code 000000

#### Scenario: Idempotent save behavior
- **WHEN** same role-menu assignments are saved multiple times
- **THEN** no duplicate records are created
- **AND** final state matches requested assignments

### Requirement: Admin Role Full Menu Access
The system SHALL grant full menu access to admin role by default.

#### Scenario: Bootstrap admin menu authorization
- **WHEN** system initializes with admin role
- **THEN** admin role has mappings to all menus
- **AND** admin users see all menus regardless of explicit assignments

#### Scenario: Admin role identification
- **WHEN** checking admin role
- **THEN** system identifies by role_code='admin' first
- **AND** falls back to role_id=1 if code not found

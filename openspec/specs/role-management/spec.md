# role-management Specification

## Purpose
Define role lifecycle and authorization scope requirements for role CRUD, assignments, and permission linkage.
## Requirements
### Requirement: Role Management
The system SHALL provide role management functionality for organizing user permissions.

#### Scenario: Create role
- **WHEN** admin creates a role with name and typeCode
- **THEN** the system creates a new role record
- **AND** returns the role ID

#### Scenario: Query roles
- **WHEN** admin queries role list with optional keyword
- **THEN** the system returns matching roles
- **AND** supports keyword filtering by name

#### Scenario: Update role
- **WHEN** admin updates a role's name or description
- **THEN** the system updates the record
- **AND** returns success with code 000000

#### Scenario: Delete role
- **WHEN** admin deletes a role by id
- **THEN** the system removes the record
- **AND** returns success with code 000000

#### Scenario: Get role detail
- **WHEN** admin requests role detail by id
- **THEN** the system returns full role information
- **AND** includes all role attributes

### Requirement: Role Data Scope
The system SHALL support data scope configuration for roles.

#### Scenario: Data scope types
- **WHEN** creating or editing a role
- **THEN** the system supports data scope values: all, custom, dept, dept_and_child, self

### Requirement: Role API Endpoints
The system SHALL provide the following API endpoints for role management.

#### Scenario: Query endpoint
- **WHEN** POST /api/role/query is called
- **THEN** returns list of roles matching keyword filter

#### Scenario: Create endpoint
- **WHEN** POST /api/role/create is called
- **THEN** creates new role and returns role ID

#### Scenario: Edit endpoint
- **WHEN** POST /api/role/edit is called
- **THEN** updates existing role

#### Scenario: Delete endpoint
- **WHEN** POST /api/role/delete/:id is called
- **THEN** deletes specified role

#### Scenario: Detail endpoint
- **WHEN** GET /api/role/detail/:id is called
- **THEN** returns role details

### Requirement: Role Menu Authorization Configuration
The system SHALL allow administrators to configure menu authorizations as part of role management workflows.

#### Scenario: Configure menus during role administration
- **WHEN** an administrator edits a role in role management
- **THEN** the administrator can view and update the role's granted menu set
- **AND** saved menu grants take effect in subsequent authorized menu queries

#### Scenario: New role bootstrap without menu grants
- **WHEN** an administrator creates a new role without explicit menu grants
- **THEN** the role starts with no business menu visibility by default
- **AND** access remains denied until menu grants are explicitly assigned

### Requirement: Legacy System Role Route Compatibility
The system SHALL support legacy role administration routes under `/system/role/*` for frontend compatibility.

#### Scenario: Frontend role page uses system role path
- **WHEN** frontend role page calls create, update, or delete operations using `/system/role/*`
- **THEN** backend MUST map requests to canonical role management handlers
- **AND** response behavior MUST match role management contract semantics

### Requirement: Role API Name Mapping Compatibility
The system SHALL handle legacy action-name differences between frontend and canonical role handlers.

#### Scenario: Update route maps to canonical edit operation
- **WHEN** frontend calls legacy role update endpoint
- **THEN** backend MUST execute canonical role edit logic
- **AND** returned payload and status MUST remain compatible with frontend handling

### Requirement: Role Member Lifecycle Management
The system SHALL provide complete role member lifecycle APIs and UI workflows, including add organization user, add external user, and remove member from role.

#### Scenario: Add existing organization user to role
- **WHEN** an administrator selects users already under the current organization and confirms assignment
- **THEN** the system MUST create missing user-role associations idempotently
- **AND** duplicate associations MUST NOT be created

#### Scenario: Add external user by account or ID
- **WHEN** an administrator searches an external user by exact account or user ID and assigns role
- **THEN** the system MUST create user-role association within target organization scope
- **AND** the operation MUST be audited

### Requirement: Last Role Safety Guard
The system SHALL enforce last-role safety semantics when removing a role-member relation.

#### Scenario: Remove member with multiple remaining roles
- **WHEN** an administrator removes one role from a user who still has other roles
- **THEN** the system MUST remove only the target relation
- **AND** the user account MUST remain active

#### Scenario: Remove member with unique remaining role
- **WHEN** an administrator removes the user's last role in all organizations
- **THEN** the system MUST execute configured safety policy
- **AND** the resulting behavior MUST be deterministic and documented

### Requirement: Custom Role Inheritance Constraint
The system SHALL enforce that custom roles inherit from allowed built-in organizational roles and cannot exceed inherited permission boundaries.

#### Scenario: Create custom role inheriting organization admin
- **WHEN** an administrator creates a custom role with inheritance base "organization-admin"
- **THEN** the system MUST validate inherited scope and cap grantable permissions to parent bounds

#### Scenario: Attempt to exceed parent permissions
- **WHEN** an administrator assigns permissions beyond inherited parent scope
- **THEN** the system MUST reject the operation with a validation error

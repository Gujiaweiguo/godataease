# role-management Specification

## Purpose
TBD - created by archiving change implement-go-role-module. Update Purpose after archive.
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


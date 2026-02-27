## ADDED Requirements

### Requirement: Role Permission Save API Availability
The system SHALL provide role permission save APIs required by frontend role management page.

#### Scenario: Save resource permissions for role
- **WHEN** frontend submits role permission save request to `/system/role/permission/save`
- **THEN** backend MUST process the request through Go permission service path
- **AND** return success or explicit failure without `404`

### Requirement: Menu and Business Permission Query APIs
The system SHALL provide menu-permission and business-permission query APIs used by permission configuration UI.

#### Scenario: Query menu and business permission trees
- **WHEN** frontend requests `/auth/menuPermission` or `/auth/busiPermission`
- **THEN** backend MUST return permission dataset in contract-compatible structure
- **AND** not fallback to static file route or generic `404`

### Requirement: Menu and Business Permission Save APIs
The system SHALL provide save APIs for menu and business permission assignments.

#### Scenario: Persist permission assignment changes
- **WHEN** frontend posts permission updates to `/auth/saveMenuPer` or `/auth/saveBusiPer`
- **THEN** backend MUST persist effective authorization state or return explicit validation/auth error
- **AND** MUST NOT return placeholder success for unimplemented logic

## ADDED Requirements

### Requirement: User management scope expands to include role workflows
The system SHALL expand the user-management capability so that governed role workflows are hosted inside the user-management page.

#### Scenario: Default user-management experience preserves user workflows
- **WHEN** a user opens the user-management page
- **THEN** existing in-scope user CRUD and maintenance workflows MUST remain available from the User tab
- **AND** the expanded page structure MUST NOT remove the governed user-management functions already available before the refactor

#### Scenario: User-management page provides in-context access to role workflows
- **WHEN** a user needs to manage roles from the user-management area
- **THEN** the user-management capability MUST provide a Role tab within the same governed page
- **AND** the role workflows exposed there MUST remain part of the user-management experience for this change scope

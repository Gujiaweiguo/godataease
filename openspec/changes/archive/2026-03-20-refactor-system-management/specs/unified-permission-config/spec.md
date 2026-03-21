## ADDED Requirements

### Requirement: Unified permission configuration entry
The system SHALL provide a unified permission configuration center instead of scattering permission workflows across multiple management pages.

#### Scenario: Permission center exposes three governed tabs
- **WHEN** a user opens the permission configuration page
- **THEN** the page MUST expose exactly three tabs for menu permission, resource permission, and row/column permission
- **AND** switching between tabs MUST keep the user inside one governed permission-management entry path

### Requirement: Role-based menu authorization remains available from the unified center
The system SHALL let administrators manage menu authorization for roles from the permission configuration center.

#### Scenario: Menu permission tab authorizes governed menu visibility
- **WHEN** a user opens the menu permission tab and saves role-menu authorization changes
- **THEN** the system MUST persist those changes through the governed permission workflow
- **AND** the resulting menu visibility MUST remain consistent with the saved role authorization state

### Requirement: Resource permission supports governed assignment views
The system SHALL provide a unified resource-permission workflow that supports the governed assignment perspectives in this change scope.

#### Scenario: Resource permission tab loads governed resources and assignees
- **WHEN** a user opens the resource permission tab
- **THEN** the page MUST load governed resource families for datasource, dataset, dashboard, and big-screen permissions
- **AND** the workflow MUST support the configured assignment perspectives without forcing the user into a separate permission page

### Requirement: Row and column permission workflows remain first-class within the unified center
The system SHALL expose row and column permission workflows from the same permission center.

#### Scenario: Row/column permission tab manages dataset data-access rules
- **WHEN** a user opens the row/column permission tab
- **THEN** the page MUST expose governed dataset-level row-filter and column-control workflows
- **AND** saving those rules MUST keep row-permission and column-permission behavior reachable from the unified permission center

## ADDED Requirements

### Requirement: Role Tab Hosts Governed Role Administration Workflows
The system SHALL use the Role tab inside user management as the primary hosted surface for governed role administration workflows added in this change.

#### Scenario: User management opens role administration tab
- **WHEN** an administrator enters the Role tab within user management
- **THEN** the page MUST expose the governed role list, role member operations, and inheritance-boundary workflows for this change
- **AND** the administrator MUST NOT need a separate standalone role page to complete those tasks

#### Scenario: Role tab keeps user-management route context
- **WHEN** an administrator switches between User and Role tabs during administration
- **THEN** the workflow MUST preserve the governed user-management route context
- **AND** role operations MUST remain reachable without rebuilding navigation state

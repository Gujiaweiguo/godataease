# role-in-user-management Specification

## Purpose
Define the information-architecture rules that move governed role-management workflows into the user-management area instead of keeping role management as a separate primary system-management page.

## Requirements

### Requirement: Role workflows moved into user management
The system SHALL host role-management workflows inside the user-management area instead of keeping role management as a separate primary system-management page.

#### Scenario: User management exposes user and role tabs
- **WHEN** a user opens the governed user-management page
- **THEN** the page MUST expose separate User and Role tabs within the same management surface
- **AND** switching between those tabs MUST preserve the governed user-management route context

### Requirement: Role operations remain available after relocation
The system SHALL preserve in-scope role operations after moving role workflows into user management.

#### Scenario: Role tab exposes governed role workflows
- **WHEN** a user opens the Role tab within user management
- **THEN** the page MUST expose the role list and the governed in-scope role operations for this change
- **AND** those workflows MUST remain executable without requiring a separate standalone role page entry

### Requirement: Standalone role menu is no longer part of primary system-management navigation
The system SHALL remove the standalone role-management menu from the governed system-management primary navigation.

#### Scenario: Role workflows are reachable only from user management in the new IA
- **WHEN** a user browses the system-management menu structure after this change
- **THEN** the standalone role-management menu entry MUST be hidden from the governed navigation
- **AND** role workflows MUST instead remain reachable through the Role tab inside user management

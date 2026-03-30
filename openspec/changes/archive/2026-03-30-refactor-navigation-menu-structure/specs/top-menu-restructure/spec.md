## MODIFIED Requirements

### Requirement: Four-entry top navigation
The system SHALL restructure the governed primary navigation into six first-level entries: Workbench, Visualization, Data Management, Organization & Permission, System Settings, and Toolbox.

#### Scenario: Authorized administrator sees six first-level entries
- **WHEN** an authenticated administrator loads the application shell after menu bootstrap succeeds
- **THEN** the primary navigation MUST expose exactly six governed first-level entries for this change scope
- **AND** the entries MUST remain ordered as Workbench, Visualization, Data Management, Organization & Permission, System Settings, and Toolbox

#### Scenario: Non-administrator does not see admin-governed groups
- **WHEN** an authenticated non-administrator loads the application shell after menu bootstrap succeeds
- **THEN** the primary navigation MUST continue to expose Workbench, Visualization, Data Management, and Toolbox according to role authorization
- **AND** Organization & Permission and System Settings MUST remain hidden unless role-menu authorization grants them

### Requirement: Visualization and data-management grouping
The system SHALL expose the governed second-level menus under Visualization and Data Management without mixing export-center entry points into Data Management.

#### Scenario: Visualization group exposes governed submenu items
- **WHEN** a user opens the Visualization first-level entry
- **THEN** the submenu MUST expose dashboard, big-screen, and template-related entries governed by the migrated menu structure
- **AND** those entries MUST remain reachable through the resolved menu path instead of behaving like missing routes

#### Scenario: Data Management group exposes governed submenu items
- **WHEN** a user opens the Data Management first-level entry
- **THEN** the submenu MUST expose datasource and dataset related entries governed by the migrated menu structure
- **AND** export-center workflows MUST NOT appear under Data Management after the restructure

### Requirement: Duplicate and deprecated top-menu entries removed from the primary IA
The system SHALL remove duplicate or deprecated primary menu entries and header shortcuts from the governed information architecture for this change.

#### Scenario: Removed help entry points do not remain in shell navigation
- **WHEN** a user browses the governed primary navigation and shell header after the restructure
- **THEN** help-documentation, product-forum, technical-blog, and enterprise-trial entry points MUST NOT appear in the primary IA
- **AND** the removed entries MUST NOT leave behind an empty More menu trigger or equivalent placeholder

#### Scenario: Relocated system capabilities remain reachable through governed grouping
- **WHEN** menu-management and export-center capabilities are relocated under new governed parent groups
- **THEN** the user-facing navigation state MUST remain internally consistent
- **AND** menu-management MUST be reachable through System Settings while export-center MUST be reachable through Toolbox

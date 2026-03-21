# top-menu-restructure Specification

## Purpose
Define the governed top-level information architecture for the post-refactor application shell so primary navigation remains compact, role-consistent, and semantically stable.

## Requirements

### Requirement: Four-entry top navigation
The system SHALL restructure the top navigation into four first-level entries: Workbench, Visualization, Data Management, and System Management.

#### Scenario: Authorized user sees four first-level entries
- **WHEN** an authenticated user loads the application shell after menu bootstrap succeeds
- **THEN** the top navigation MUST expose exactly four governed first-level entries for this change scope
- **AND** the entries MUST remain ordered as Workbench, Visualization, Data Management, and System Management

### Requirement: Visualization and data-management grouping
The system SHALL expose the re-grouped second-level menus under Visualization and Data Management.

#### Scenario: Visualization group exposes governed submenu items
- **WHEN** a user opens the Visualization first-level entry
- **THEN** the submenu MUST expose dashboard, big-screen, and template-related entries governed by the migrated menu structure
- **AND** those entries MUST remain reachable through the resolved menu path instead of behaving like missing routes

#### Scenario: Data Management group exposes governed submenu items
- **WHEN** a user opens the Data Management first-level entry
- **THEN** the submenu MUST expose datasource, dataset, and export-center related entries in the governed grouping
- **AND** those entries MUST remain reachable through the resolved menu path instead of behaving like missing routes

### Requirement: Duplicate and deprecated top-menu entries removed from the primary IA
The system SHALL remove duplicate or deprecated primary menu entries from the governed top-level information architecture for this change.

#### Scenario: Duplicate permission menu is absent from primary navigation
- **WHEN** a user browses the governed first-level and second-level navigation structure
- **THEN** the duplicate menu-permission entry MUST NOT appear as a separate menu item
- **AND** permission workflows MUST instead be reachable through the unified permission configuration entry

#### Scenario: Hidden or relocated legacy entries do not create ambiguous navigation state
- **WHEN** a legacy or relocated menu entry is no longer shown in the governed top navigation
- **THEN** the user-facing navigation state MUST remain internally consistent
- **AND** relocated functionality MUST still be reachable through the new governed grouping

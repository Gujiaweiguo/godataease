## ADDED Requirements

### Requirement: Permission configuration becomes a unified configuration center
The system SHALL expand permission configuration from a narrower permission-maintenance surface into a unified permission configuration center.

#### Scenario: Permission configuration page exposes unified IA
- **WHEN** a user opens permission configuration
- **THEN** the page MUST present a unified information architecture for menu permission, resource permission, and row/column permission workflows
- **AND** the user MUST NOT need a duplicate dedicated menu-permission page to complete those governed tasks

#### Scenario: Existing permission workflows remain semantically reachable after consolidation
- **WHEN** permission-management workflows previously reached through scattered entry points are consolidated
- **THEN** the permission-config capability MUST still expose the governed authorization and data-access workflows required by this change
- **AND** the consolidated page MUST preserve consistent save-and-revisit behavior for the in-scope permission tabs

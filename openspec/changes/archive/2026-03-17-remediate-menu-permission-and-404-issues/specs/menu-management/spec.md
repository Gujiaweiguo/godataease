## MODIFIED Requirements

This is a delta spec that `openspec/specs/menu-management/spec.md`.

### Requirement: Menu Query
The system SHALL provide menu query functionality for navigation.

#### Scenario: Query menu tree returns only healthy runtime menus
- **WHEN** user requests runtime menu list
- **THEN** the system returns hierarchical menu structure
- **AND** menus are sorted by menuSort
- **AND** parent-child relationships are preserved
- **AND** returned runtime menus MUST NOT include entries with broken route or invalid component mappings

### Requirement: Menu Query Must Be Data-Driven
The system SHALL generate menu trees from persistent menu data and SHALL NOT rely on hardcoded menu item lists for runtime menu responses.

#### Scenario: Runtime menu data remains aligned with actual reachable pages
- **WHEN** menu records are loaded for runtime navigation
- **THEN** the system MUST derive navigation from persisted menu data
- **AND** visible menus MUST remain consistent with actual registered routes and page components
- **AND** menu data MUST NOT surface entries that are known to be unreachable in the current runtime

### Requirement: Menu Admin Console
The system SHALL provide a frontend menu administration console for persisted menu data in `core_menu`.

#### Scenario: Administrator views unhealthy menu configuration
- **WHEN** an authorized administrator inspects menu management data
- **THEN** the system SHOULD support identifying menus with missing route targets, invalid component paths, or other runtime navigation inconsistencies
- **AND** such issues SHOULD be actionable for remediation

#### Scenario: Administrator edits menu metadata that would create broken navigation
- **WHEN** an administrator edits menu name, path, component, icon, sort, or hidden flag and submits
- **THEN** the system MUST persist valid updates through menu APIs
- **AND** invalid path/component combinations MUST be rejected with explicit validation errors
- **AND** updates that would create known broken runtime navigation MUST NOT be silently accepted

### Requirement: Runtime Menu Health Governance
The system SHALL ensure that runtime-visible menus remain consistent with registered frontend navigation targets.

#### Scenario: Visible menu resolves to registered frontend target
- **WHEN** a menu is visible to an authorized user
- **THEN** the menu target MUST resolve to a valid frontend route
- **AND** the associated page component MUST be loadable in the active application build
- **AND** clicking the visible menu MUST NOT fail due to missing route registration

#### Scenario: Menu becomes unhealthy after route drift
- **WHEN** persisted menu configuration drifts from actual frontend route or component definitions
- **THEN** the inconsistency MUST be detectable during governance verification
- **AND** the menu MUST be corrected before being considered healthy for release

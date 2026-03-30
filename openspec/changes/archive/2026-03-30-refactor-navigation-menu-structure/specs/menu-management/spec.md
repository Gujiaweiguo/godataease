## MODIFIED Requirements

### Requirement: Menu Query
The system SHALL provide menu query functionality for navigation, including grouped admin-governance menus and event-type child menus.

#### Scenario: Query menu tree
- **WHEN** user requests menu list
- **THEN** the system returns hierarchical menu structure
- **AND** menus are sorted by menuSort
- **AND** parent-child relationships are preserved

#### Scenario: Query menu tree returns grouped toolbox and admin settings layout
- **WHEN** user requests runtime menu list after the restructure
- **THEN** the returned tree MUST expose Organization & Permission, System Settings, and Toolbox according to persisted parent-child relationships and role authorization
- **AND** Toolbox MUST contain Data Export Center as a child node instead of flattening it into a top-level event entry

### Requirement: Menu Admin Console
The system SHALL provide a frontend menu administration console for persisted menu data in `core_menu`, including grouped menus and event-type child entries.

#### Scenario: View full menu tree
- **WHEN** an authorized administrator opens menu management page
- **THEN** the system MUST display all menu nodes in hierarchical order by `menu_sort`
- **AND** the tree MUST reflect Organization & Permission, System Settings, Toolbox, and their governed children exactly as persisted

#### Scenario: Edit menu metadata
- **WHEN** an administrator edits menu name, path, component, icon, sort, or hidden flag and submits
- **THEN** the system MUST persist updates through menu APIs
- **AND** invalid path/component combinations MUST be rejected with explicit validation errors
- **AND** event-type child menus such as Data Export Center MUST remain classifiable without being forced into route-only metadata

## ADDED Requirements

### Requirement: Menu restructuring must preserve governed parent assignment
The system SHALL preserve explicit parent assignment for relocated menu nodes during navigation restructuring.

#### Scenario: Admin-governance pages move under new parent groups
- **WHEN** menu-management, user-management, organization-management, role-management, and permission-management nodes are relocated under new first-level parents
- **THEN** each node MUST be returned under its governed parent in menu queries
- **AND** the resulting tree MUST NOT duplicate the same node under both old and new parents

#### Scenario: Removed header help entries do not remain as orphan menu records
- **WHEN** help-link entry points are removed from the shell navigation
- **THEN** the persisted menu dataset returned for runtime navigation MUST NOT expose those removed entries as visible nodes
- **AND** runtime clients MUST NOT need frontend-only exclusion rules to hide them

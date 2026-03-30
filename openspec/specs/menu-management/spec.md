# menu-management Specification

## Purpose
Define menu domain behaviors and admin capabilities for querying, managing, sorting, and safely deleting hierarchical menus used by runtime navigation.
## Requirements
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

#### Scenario: Menu structure
- **WHEN** menu is returned
- **THEN** each menu has path, component, name, meta, and children
- **AND** meta contains title and icon

### Requirement: Menu API Endpoints
The system SHALL provide the following API endpoint for menu management.

#### Scenario: Query endpoint
- **WHEN** GET /api/menu/query is called
- **THEN** returns list of root menus with nested children

### Requirement: Menu Management CRUD
The system SHALL provide full menu management operations for persisted menu data, including create, update, delete, and ordering adjustments.

#### Scenario: Administrator creates a menu item
- **WHEN** an administrator submits menu metadata (name, path, component, icon, parent, sort, visibility)
- **THEN** the system persists a new menu record in `core_menu`
- **AND** the menu tree query reflects the new item in correct hierarchy and sort order

#### Scenario: Administrator updates and reorders menus
- **WHEN** an administrator updates menu metadata or sort order
- **THEN** the system updates persisted menu records atomically
- **AND** subsequent menu queries return the updated ordering and metadata

#### Scenario: Administrator deletes menu item safely
- **WHEN** an administrator deletes a menu item with children
- **THEN** the system rejects deletion unless child-handling policy is satisfied
- **AND** the system returns a structured error describing dependency constraint

### Requirement: Menu Query Must Be Data-Driven
The system SHALL generate menu trees from persistent menu data and SHALL NOT rely on hardcoded menu item lists for runtime menu responses.

#### Scenario: Query menu tree after restart
- **WHEN** the application restarts
- **THEN** menu query endpoints return menu trees derived from persisted records
- **AND** no hardcoded static list is required to render standard menus

#### Scenario: Runtime menu data remains aligned with actual reachable pages
- **WHEN** menu records are loaded for runtime navigation
- **THEN** the system MUST derive navigation from persisted menu data
- **AND** visible menus MUST remain consistent with actual registered routes and page components
- **AND** menu data MUST NOT surface entries that are known to be unreachable in the current runtime

#### Scenario: Empty menu dataset handling
- **WHEN** no active menu records exist for a scope
- **THEN** menu query endpoints return an empty list with success response
- **AND** clients do not receive synthetic hardcoded menus

### Requirement: Menu Admin Console
The system SHALL provide a frontend menu administration console for persisted menu data in `core_menu`, including grouped menus and event-type child entries.

#### Scenario: View full menu tree
- **WHEN** an authorized administrator opens menu management page
- **THEN** the system MUST display all menu nodes in hierarchical order by `menu_sort`
- **AND** the tree MUST reflect Organization & Permission, System Settings, Toolbox, and their governed children exactly as persisted

#### Scenario: Administrator views unhealthy menu configuration
- **WHEN** an authorized administrator inspects menu management data
- **THEN** the system SHOULD support identifying menus with missing route targets, invalid component paths, or other runtime navigation inconsistencies
- **AND** such issues SHOULD be actionable for remediation

#### Scenario: Edit menu metadata
- **WHEN** an administrator edits menu name, path, component, icon, sort, or hidden flag and submits
- **THEN** the system MUST persist updates through menu APIs
- **AND** invalid path/component combinations MUST be rejected with explicit validation errors
- **AND** event-type child menus such as Data Export Center MUST remain classifiable without being forced into route-only metadata

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

### Requirement: Safe Menu Delete
The system SHALL enforce safe deletion rules for menu nodes with descendants.

#### Scenario: Delete menu with children
- **WHEN** an administrator attempts to delete a menu node that has child nodes
- **THEN** the system MUST reject deletion unless child-handling policy is satisfied
- **AND** the response MUST include dependency details

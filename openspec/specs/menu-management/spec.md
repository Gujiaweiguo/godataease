# menu-management Specification

## Purpose
Define menu domain behaviors and admin capabilities for querying, managing, sorting, and safely deleting hierarchical menus used by runtime navigation.
## Requirements
### Requirement: Menu Query
The system SHALL provide menu query functionality for navigation.

#### Scenario: Query menu tree
- **WHEN** user requests menu list
- **THEN** the system returns hierarchical menu structure
- **AND** menus are sorted by menuSort
- **AND** parent-child relationships are preserved

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

#### Scenario: Empty menu dataset handling
- **WHEN** no active menu records exist for a scope
- **THEN** menu query endpoints return an empty list with success response
- **AND** clients do not receive synthetic hardcoded menus

### Requirement: Menu Admin Console
The system SHALL provide a frontend menu administration console for persisted menu data in `core_menu`.

#### Scenario: View full menu tree
- **WHEN** an authorized administrator opens menu management page
- **THEN** the system MUST display all menu nodes in hierarchical order by `menu_sort`
- **AND** each row MUST expose editable metadata fields

#### Scenario: Edit menu metadata
- **WHEN** an administrator edits menu name, path, component, icon, sort, or hidden flag and submits
- **THEN** the system MUST persist updates through menu APIs
- **AND** invalid path/component combinations MUST be rejected with explicit validation errors

### Requirement: Safe Menu Delete
The system SHALL enforce safe deletion rules for menu nodes with descendants.

#### Scenario: Delete menu with children
- **WHEN** an administrator attempts to delete a menu node that has child nodes
- **THEN** the system MUST reject deletion unless child-handling policy is satisfied
- **AND** the response MUST include dependency details

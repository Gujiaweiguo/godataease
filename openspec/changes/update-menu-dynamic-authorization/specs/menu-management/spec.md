## ADDED Requirements
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

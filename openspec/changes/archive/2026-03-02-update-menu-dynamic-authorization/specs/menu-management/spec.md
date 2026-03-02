## ADDED Requirements

### Requirement: Menu CRUD Operations
The system SHALL provide menu CRUD operations for dynamic menu management.

#### Scenario: Create menu
- **WHEN** admin creates a menu with name, pid, icon, path, and sort
- **THEN** the system creates a new menu record in `core_menu`
- **AND** returns the menu ID

#### Scenario: Update menu
- **WHEN** admin updates menu properties
- **THEN** the system updates the record
- **AND** changes reflect in subsequent menu queries

#### Scenario: Delete menu
- **WHEN** admin deletes a menu without children
- **THEN** the system removes the record
- **AND** related role-menu mappings are removed

#### Scenario: Delete menu with children rejected
- **WHEN** admin attempts to delete a menu with child menus
- **THEN** the system rejects the operation
- **AND** returns error indicating child menus exist

### Requirement: Menu Ordering and Visibility
The system SHALL support menu ordering and visibility control.

#### Scenario: Update menu sort order
- **WHEN** admin changes menu sort value
- **THEN** menu appears in updated position in tree queries

#### Scenario: Hide menu from navigation
- **WHEN** admin sets menu hidden flag to true
- **THEN** the menu is excluded from navigation tree output
- **AND** direct route access is still controlled by authorization

### Requirement: Menu Tree Query
The system SHALL provide menu tree query for management UI.

#### Scenario: Query full menu tree
- **WHEN** admin requests menu list
- **THEN** system returns hierarchical menu tree with all menus
- **AND** includes parent-child relationships

### Requirement: Menu API Endpoints
The system SHALL provide the following API endpoints for menu management.

#### Scenario: Query endpoint
- **WHEN** GET /api/menu/query is called
- **THEN** returns hierarchical menu tree

#### Scenario: Create endpoint
- **WHEN** POST /api/menu/create is called with menu payload
- **THEN** creates new menu and returns menu ID

#### Scenario: Update endpoint
- **WHEN** POST /api/menu/edit is called with menu ID and updates
- **THEN** updates existing menu

#### Scenario: Delete endpoint
- **WHEN** POST /api/menu/delete/:id is called
- **THEN** deletes specified menu if no children exist

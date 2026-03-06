## ADDED Requirements

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

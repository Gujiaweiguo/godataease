## ADDED Requirements

### Requirement: Visual form field editor
The system SHALL provide a native form editor page where admin users define form fields, configure field types, set validation rules, and arrange field order. The editor SHALL load and save form definitions via `getFormById`, `createForm`, and `updateForm` APIs.

#### Scenario: Open editor for existing form
- **WHEN** user navigates to the form editor with a form id route parameter
- **THEN** the system calls `getFormById` and populates the editor with the existing field definitions
- **AND** field types, labels, validation settings, and order are rendered for editing

#### Scenario: Open editor for new form
- **WHEN** user navigates to the form editor without a form id (new form mode)
- **THEN** the editor displays an empty canvas with an "Add Field" action
- **AND** the parent folder pid is pre-populated from route parameters

#### Scenario: Add a new field to the form
- **WHEN** user clicks "Add Field" and selects a field type (text, number, decimal, date, select, etc.)
- **THEN** a new field configuration block appears in the editor
- **AND** the field has editable properties: label name, field type, required flag, and type-specific settings

#### Scenario: Reorder fields via drag-and-drop
- **WHEN** user drags a field block to a new position
- **THEN** the field order updates in the editor layout
- **AND** the serialized `forms` field definition reflects the new order on save

#### Scenario: Remove a field from the form
- **WHEN** user deletes a field block from the editor
- **THEN** the field is removed from the editor canvas
- **AND** the change is staged for the next save operation

#### Scenario: Save form definition
- **WHEN** user clicks "Save" on the form editor
- **THEN** the system serializes the field configuration into the `forms` JSON string
- **AND** calls `createForm` for new forms or `updateForm` for existing forms
- **AND** navigates back to the form management tree on success

### Requirement: Datasource and table binding in editor
The system SHALL allow the admin to bind a form to a specific datasource and configure whether to use an existing table or create a new one.

#### Scenario: Bind form to a new table
- **WHEN** user selects a datasource and leaves "use existing table" unchecked
- **THEN** the system sends `useExistsTable: false` with a `tableName` in the save request
- **AND** the backend creates the physical table on save

#### Scenario: Bind form to an existing table
- **WHEN** user selects a datasource and enables "use existing table"
- **THEN** the system calls `getBuiltInTables` to list available tables
- **AND** user selects a table name from the list
- **AND** the system sends `useExistsTable: true` with the selected `tableName` in the save request

### Requirement: Index configuration in editor
The system SHALL allow the admin to configure table indexes for forms that create new tables.

#### Scenario: Enable index creation
- **WHEN** user toggles "create index" on and configures index columns
- **THEN** the system sends `createIndex: true` with the `tableIndexes` configuration in the save request

#### Scenario: Index section hidden for existing tables
- **WHEN** user enables "use existing table"
- **THEN** the index configuration section is hidden or disabled
- **AND** the save request sends `createIndex: false`

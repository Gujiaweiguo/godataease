## ADDED Requirements

### Requirement: System Variable Definition Management
The Go backend SHALL provide management APIs for system variable definitions required by the frontend.

#### Scenario: Create and edit system variable definitions
- **WHEN** a client calls `/sysVariable/create` or `/sysVariable/edit` with a valid variable payload
- **THEN** the backend persists the variable definition deterministically
- **AND** the resulting state is observable through subsequent query or detail requests

#### Scenario: Query and inspect system variable definitions
- **WHEN** a client calls `/sysVariable/query` or `/sysVariable/detail/:id`
- **THEN** the backend returns variable records and detail data in a contract compatible with the current frontend management view

#### Scenario: Delete a system variable definition
- **WHEN** a client calls `/sysVariable/delete/:id` for an existing variable definition
- **THEN** the backend deletes or deactivates the target variable deterministically according to management rules

### Requirement: System Variable Value Management
The Go backend SHALL provide management APIs for the selectable values attached to a system variable.

#### Scenario: Create and edit system variable values
- **WHEN** a client calls `/sysVariable/value/create` or `/sysVariable/value/edit` with a valid variable-value payload
- **THEN** the backend persists the variable value deterministically
- **AND** the resulting state is observable through subsequent selection requests

#### Scenario: Delete one or more system variable values
- **WHEN** a client calls `/sysVariable/value/delete/:id` or `/sysVariable/value/batchDel`
- **THEN** the backend deletes the targeted variable values deterministically
- **AND** leaves remaining values queryable for the associated variable definition

### Requirement: System Variable Value Selection and Paging
The Go backend SHALL provide selection APIs for variable values that support the frontend editor experience.

#### Scenario: Page selected values for a variable
- **WHEN** a client calls `/sysVariable/value/selected/:page/:limit` with selection filters
- **THEN** the backend returns deterministic paged variable-value results compatible with the current frontend selection UI

#### Scenario: List selected values by variable identifier
- **WHEN** a client calls `/sysVariable/value/selected/:id`
- **THEN** the backend returns the selected values associated with the specified system variable definition

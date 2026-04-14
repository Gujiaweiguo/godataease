# datasource-table-exploration-canonical Specification

## Purpose
Define the governed canonical datasource table exploration routes and the compatibility-safe migration boundary for frontend callers.

## Requirements
### Requirement: Datasource Table List Must Have A Canonical Route
The system SHALL provide datasource table listing through a canonical route at `POST /api/ds/tables`.

#### Scenario: Query datasource tables through canonical route
- **WHEN** a client calls `POST /api/ds/tables` with a valid datasource table query payload
- **THEN** the system MUST return datasource table records using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve the table-listing semantics currently used by datasource exploration pages

### Requirement: Datasource Table Status Must Have A Canonical Route
The system SHALL provide datasource table status lookup through a canonical route at `POST /api/ds/tableStatus`.

#### Scenario: Query datasource table status through canonical route
- **WHEN** a client calls `POST /api/ds/tableStatus` with a valid datasource table status request
- **THEN** the system MUST return table synchronization status information using the established compatibility-safe response envelope
- **AND** the canonical route MUST preserve explicit unknown or non-success semantics when no status evidence exists

### Requirement: Datasource Table Field Metadata Must Have A Canonical Route
The system SHALL provide datasource table field metadata through a canonical route at `POST /api/ds/tableField`.

#### Scenario: Query datasource table fields through canonical route
- **WHEN** a client calls `POST /api/ds/tableField` with a valid datasource table field request
- **THEN** the system MUST return table field metadata using the existing field contract expected by datasource exploration flows
- **AND** the canonical route MUST preserve deterministic success and explicit failure semantics for invalid datasource or table inputs

### Requirement: Datasource Schema Discovery Must Have A Canonical Route
The system SHALL provide datasource schema discovery through a canonical route at `POST /api/ds/schema`.

#### Scenario: Query datasource schemas through canonical route
- **WHEN** a client calls `POST /api/ds/schema` with a valid schema discovery request
- **THEN** the system MUST return schema names using the established compatibility-safe response envelope
- **AND** the canonical route MUST remain compatible with existing datasource editor schema selection flows

### Requirement: Canonical Datasource Table Exploration Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource table exploration callers to migrate to canonical routes without removing the legacy compatibility aliases in the same change.

#### Scenario: Frontend table exploration callers switch safely to canonical routes
- **WHEN** the frontend datasource API layer migrates table list, table status, table field, and schema callers to canonical `/api/ds/*` routes
- **THEN** the corresponding `/datasource/*` compatibility routes MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting the frontend route selection without requiring emergency backend route restoration

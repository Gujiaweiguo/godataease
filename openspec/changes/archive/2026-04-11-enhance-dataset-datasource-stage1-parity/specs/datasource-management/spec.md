## ADDED Requirements

### Requirement: Datasource Table Status Must Reflect Real Synchronization State
The system SHALL return datasource table status using real synchronization evidence or explicit unknown-state semantics rather than a fixed success placeholder.

#### Scenario: Query table status with synchronization history
- **WHEN** a client calls `POST /datasource/getTableStatus` for a datasource table that has synchronization history or task records
- **THEN** the system MUST derive the returned status from the latest relevant synchronization evidence
- **AND** the response MUST include a stable status value and the latest available update time for frontend display

#### Scenario: Query table status without synchronization history
- **WHEN** a client calls `POST /datasource/getTableStatus` for a datasource table that has no usable synchronization evidence
- **THEN** the system MUST return an explicit unknown, uninitialized, or equivalent non-success-assuming state
- **AND** the endpoint MUST NOT report a synthetic success status solely because the route is reachable

### Requirement: Datasource Delete Compatibility And Canonical Write Paths Must Remain Consistent
The system SHALL provide a normative datasource delete write path while keeping historical compatibility delete routes behaviorally consistent during migration.

#### Scenario: Delete datasource through normative write route
- **WHEN** a client calls the normative datasource delete write route for a deletable datasource or datasource folder
- **THEN** the system MUST execute the same recursive delete business logic used by compatibility aliases
- **AND** the route MUST return deterministic success or failure semantics compatible with existing callers

#### Scenario: Compatibility delete route matches canonical delete semantics
- **WHEN** a client calls a historical datasource delete compatibility alias for the same target resource
- **THEN** the compatibility route MUST resolve through the same underlying delete logic as the normative write route
- **AND** both paths MUST preserve the same pre-delete validation, permission checks, and business failure semantics

#### Scenario: Datasource delete failure remains explicit
- **WHEN** datasource deletion is blocked by missing resource, authorization denial, dependent relations, or recursive delete failure
- **THEN** the system MUST return explicit non-success semantics
- **AND** the result MUST remain distinguishable from route absence and from a successful no-op response

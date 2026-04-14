## ADDED Requirements

### Requirement: Datasource Tree Query Must Have A Governed Canonical Route
The system SHALL provide datasource tree query capability through a governed canonical route in addition to existing compatibility aliases.

#### Scenario: Query datasource tree through canonical management route
- **WHEN** a client calls `POST /api/ds/tree` for datasource discovery data
- **THEN** the system MUST return datasource tree nodes using the established `BusiTreeNode` contract and `code/data/msg` envelope
- **AND** the canonical result MUST remain consistent with compatibility datasource tree semantics for supported requests

### Requirement: Datasource Detail Read Must Have A Governed Canonical Route
The system SHALL provide datasource detail read capability through a governed canonical route for datasource management flows.

#### Scenario: Query datasource detail through canonical management route
- **WHEN** a client calls `GET /api/ds/:id` for a valid datasource
- **THEN** the system MUST return datasource detail data using the established datasource detail contract
- **AND** missing or unauthorized datasource outcomes MUST remain distinguishable from route absence

### Requirement: Datasource Core Write Lifecycle Must Have Governed Canonical Routes
The system SHALL provide canonical datasource create, update, and delete routes for the core datasource lifecycle while keeping compatibility behavior aligned.

#### Scenario: Canonical datasource create remains parity-safe
- **WHEN** a client calls `POST /api/ds/save` with a valid datasource creation request
- **THEN** the system MUST execute the same governed creation rules used by compatibility datasource save flows
- **AND** validation or connectivity failures MUST remain explicit and testable

#### Scenario: Canonical datasource update remains parity-safe
- **WHEN** a client calls `POST /api/ds/update` with a valid datasource update request
- **THEN** the system MUST execute the same governed update rules used by compatibility datasource update flows
- **AND** authorization or business-rule failures MUST remain explicit and testable

#### Scenario: Canonical datasource delete remains parity-safe
- **WHEN** a client calls `POST /api/ds/delete/:id` for a datasource delete request
- **THEN** the system MUST execute the same governed delete logic used by compatibility datasource delete flows
- **AND** the response MUST preserve deterministic success or explicit failure semantics

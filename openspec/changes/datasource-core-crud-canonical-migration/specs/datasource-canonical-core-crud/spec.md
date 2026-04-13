## ADDED Requirements

### Requirement: Datasource Tree Must Have A Canonical Read Route
The system SHALL provide a canonical datasource tree route at `POST /api/ds/tree` for datasource discovery flows that currently depend on compatibility aliases.

#### Scenario: Query datasource tree through canonical route
- **WHEN** a client calls `POST /api/ds/tree` with datasource tree query parameters
- **THEN** the system MUST return datasource tree nodes using the established `code/data/msg` envelope
- **AND** the returned node structure MUST remain compatible with existing frontend datasource tree consumers

### Requirement: Datasource Detail Must Have A Canonical Read Route
The system SHALL provide a canonical datasource detail route for datasource read-by-id flows used by datasource management pages.

#### Scenario: Query datasource detail through canonical route
- **WHEN** a client calls `GET /api/ds/:id` for an existing datasource
- **THEN** the system MUST return datasource detail information compatible with the existing datasource detail contract
- **AND** the canonical response MUST remain behaviorally consistent with the compatibility detail route for the same datasource

### Requirement: Datasource Core Write Paths Must Have Canonical Routes
The system SHALL provide canonical datasource create, update, and delete routes for the primary datasource lifecycle operations.

#### Scenario: Create datasource through canonical route
- **WHEN** a client calls `POST /api/ds/save` with a valid datasource creation payload
- **THEN** the system MUST create the datasource through the same governed business logic used by the compatibility path
- **AND** the response MUST preserve deterministic success or validation-failure semantics

#### Scenario: Update datasource through canonical route
- **WHEN** a client calls `POST /api/ds/update` with a valid datasource update payload
- **THEN** the system MUST update the datasource through the same governed business logic used by the compatibility path
- **AND** the response MUST preserve deterministic success or validation-failure semantics

#### Scenario: Delete datasource through canonical route
- **WHEN** a client calls `POST /api/ds/delete/:id` for a deletable datasource or datasource folder
- **THEN** the system MUST execute the same governed delete behavior used by the compatibility path
- **AND** deletion failure MUST remain distinguishable from missing-route behavior and successful no-op responses

### Requirement: Canonical Datasource Core CRUD Migration Must Preserve Compatibility Safety
The system SHALL allow frontend datasource core CRUD callers to migrate to canonical routes without removing compatibility aliases during the same change.

#### Scenario: Frontend core CRUD callers switch to canonical datasource routes safely
- **WHEN** the frontend datasource API layer migrates tree, detail, save, update, or delete callers to canonical `/api/ds/*` routes
- **THEN** existing compatibility `/datasource/*` aliases MUST remain available during the migration window
- **AND** rollback MUST remain possible by reverting frontend route selection without requiring emergency backend route restoration

## ADDED Requirements

### Requirement: Datasource UI Metadata And Sync Record Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource type listing, finish-page preference, recent-use tracking, and sync record pagination capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource UI metadata and sync record use canonical routes consistently
- **WHEN** a client requests datasource types, finish-page status, finish-page dismissal, recently used types, or sync record listing
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource UI Metadata And Sync Record Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource UI metadata and sync record response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource UI metadata and sync record remain contract-safe
- **WHEN** the frontend datasource API layer switches types, showFinishPage, setShowFinishPage, latestUse, and syncRecord callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource ID, missing authentication, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

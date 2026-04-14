## ADDED Requirements

### Requirement: Datasource Preview And Sync Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource preview and synchronization capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource preview and sync use canonical routes consistently
- **WHEN** a client performs datasource preview data retrieval or synchronization actions
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource Preview And Sync Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource preview and sync response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource preview and sync remain contract-safe
- **WHEN** the frontend datasource API layer switches preview and sync callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource preview/sync flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource, invalid sync target, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

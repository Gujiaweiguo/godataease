## ADDED Requirements

### Requirement: Datasource File Ingest Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource file upload and remote file loading capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource file ingest uses canonical routes consistently
- **WHEN** a client performs datasource file upload or remote file loading
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource File Ingest Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource file ingest response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource file ingest remains contract-safe
- **WHEN** the frontend datasource API layer switches file ingest callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource ingest flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid upload input, invalid remote source, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

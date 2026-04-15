## ADDED Requirements

### Requirement: Datasource Validation And Checking Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource validate-by-ID, name/type repeat checking, and API datasource validity checking capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource validation and checking uses canonical routes consistently
- **WHEN** a client performs datasource validation by ID, name/type repeat checking, or API datasource validity checking
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource Validation And Checking Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource validation and checking response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource validation and checking remains contract-safe
- **WHEN** the frontend datasource API layer switches validation/checking callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource validation/checking flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource ID, duplicate name/type, invalid API datasource, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

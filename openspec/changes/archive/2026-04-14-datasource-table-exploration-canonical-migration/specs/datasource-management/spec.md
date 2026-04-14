## ADDED Requirements

### Requirement: Datasource Table Exploration Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource table exploration capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource table exploration uses canonical routes consistently
- **WHEN** a client performs datasource table exploration for table listing, schema lookup, status lookup, or field inspection
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with the corresponding compatibility datasource routes

### Requirement: Datasource Exploration Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource exploration response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource table exploration remains contract-safe
- **WHEN** the frontend datasource API layer switches table exploration callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource exploration pages MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid datasource, invalid table, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

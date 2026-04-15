## ADDED Requirements

### Requirement: Datasource Get Variants And PerDelete Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource password-hidden detail, simplified info, and permanent deletion capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource get variants and perDelete use canonical routes consistently
- **WHEN** a client requests datasource details with hidden passwords, simplified datasource info, or performs permanent deletion
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource Get Variants And PerDelete Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource get variants and permanent deletion response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource get variants and perDelete remain contract-safe
- **WHEN** the frontend datasource API layer switches get variants and perDelete callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** non-existent datasource ID, invalid ID format, or unmet deletion precondition outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

## ADDED Requirements

### Requirement: Datasource Tree/Folder Management Must Be Available Through Governed Canonical Routes
The system SHALL provide datasource move, rename, and folder creation capabilities through governed canonical datasource routes in addition to compatibility aliases.

#### Scenario: Datasource tree/folder management uses canonical routes consistently
- **WHEN** a client performs datasource or folder move, rename, or folder creation
- **THEN** the system MUST expose governed canonical routes for those operations under `/api/ds/*`
- **AND** the canonical routes MUST remain behaviorally consistent with corresponding compatibility datasource routes

### Requirement: Datasource Tree/Folder Management Responses Must Preserve Compatibility-Safe Contracts
The system SHALL preserve datasource tree/folder management response shapes and explicit failure semantics when callers migrate from compatibility routes to canonical routes.

#### Scenario: Canonical datasource tree/folder management remains contract-safe
- **WHEN** the frontend datasource API layer switches tree/folder management callers from `/datasource/*` to `/api/ds/*`
- **THEN** datasource tree/folder management flows MUST continue receiving compatibility-safe envelopes and expected payload structures
- **AND** invalid move target, rename conflict, duplicate folder name, or unavailable backend outcomes MUST remain explicit and testable rather than silently degrading into empty success responses

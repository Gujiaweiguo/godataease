## ADDED Requirements

### Requirement: Interactive Visualization Tree Parity
The system SHALL provide `dataVisualization/interactiveTree` as an authorization-filtered visualization resource tree for dashboard and big-screen discovery flows.

#### Scenario: Interactive tree returns real dashboard resources
- **WHEN** frontend requests `POST /api/dataVisualization/interactiveTree` with `dashboard` scope
- **THEN** the system MUST return real dashboard resource nodes derived from visualization data rather than synthetic authorization placeholders
- **AND** each returned node MUST preserve the frontend tree contract required by interactive consumers

#### Scenario: Interactive tree returns real big-screen resources
- **WHEN** frontend requests `POST /api/dataVisualization/interactiveTree` with `dataV` scope
- **THEN** the system MUST return real big-screen resource nodes derived from visualization data rather than synthetic authorization placeholders
- **AND** the node hierarchy MUST remain usable for downstream selection and navigation flows

### Requirement: Interactive Tree Authorization Filtering
Interactive visualization trees SHALL filter unauthorized visualization resources without breaking the remaining tree structure.

#### Scenario: Unauthorized visualization nodes are filtered safely
- **WHEN** a user lacks access to part of the dashboard or big-screen resource tree
- **THEN** unauthorized nodes MUST be excluded from the response
- **AND** remaining nodes MUST keep valid identifiers, parent relationships, and leaf semantics

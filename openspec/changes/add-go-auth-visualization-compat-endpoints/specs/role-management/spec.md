## ADDED Requirements

### Requirement: Legacy System Role Route Compatibility
The system SHALL support legacy role administration routes under `/system/role/*` for frontend compatibility.

#### Scenario: Frontend role page uses system role path
- **WHEN** frontend role page calls create, update, or delete operations using `/system/role/*`
- **THEN** backend MUST map requests to canonical role management handlers
- **AND** response behavior MUST match role management contract semantics

### Requirement: Role API Name Mapping Compatibility
The system SHALL handle legacy action-name differences between frontend and canonical role handlers.

#### Scenario: Update route maps to canonical edit operation
- **WHEN** frontend calls legacy role update endpoint
- **THEN** backend MUST execute canonical role edit logic
- **AND** returned payload and status MUST remain compatible with frontend handling

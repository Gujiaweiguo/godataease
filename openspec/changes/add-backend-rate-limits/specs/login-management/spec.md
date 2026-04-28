## ADDED Requirements

### Requirement: Local Login Must Enforce Burst-Safe Request Limits
The system SHALL enforce request-throttling on local login endpoints so repeated login attempts cannot be performed at unbounded frequency.

#### Scenario: Local login remains available within request budget
- **WHEN** a client calls `POST /login/localLogin` or `POST /api/login/localLogin` within the configured login request budget
- **THEN** the system MUST continue to validate credentials using the existing login semantics
- **AND** successful and failed credential outcomes MUST preserve their current response behavior

#### Scenario: Local login exceeds request budget
- **WHEN** a client exceeds the configured request budget for local login attempts within the active rate-limit window
- **THEN** the system MUST reject the login attempt with an explicit throttling response
- **AND** the system MUST NOT continue into credential validation for the throttled request

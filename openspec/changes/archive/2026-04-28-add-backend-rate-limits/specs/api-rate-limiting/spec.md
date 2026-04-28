## ADDED Requirements

### Requirement: Backend Must Support Route-Scoped Rate Limiting
The system SHALL provide route-scoped backend rate limiting that can be attached to selected HTTP endpoints without changing unrelated route behavior.

#### Scenario: Limited route remains available within budget
- **WHEN** a client sends requests to a rate-limited route within the configured request budget
- **THEN** the system MUST process the request using the existing route semantics
- **AND** the system MUST NOT change the normal success envelope solely because rate limiting is enabled

#### Scenario: Limited route exceeds request budget
- **WHEN** a client exceeds the configured request budget for a rate-limited route within the active time window
- **THEN** the system MUST reject the request with an explicit throttling response
- **AND** the system MUST use HTTP `429 Too Many Requests`

### Requirement: Rate Limiting Must Use Route-Appropriate Identity Keys
The system SHALL key rate-limiting decisions by a stable request identity appropriate to the protected route surface.

#### Scenario: Unauthenticated route uses client network identity
- **WHEN** the system rate-limits an unauthenticated route such as local login
- **THEN** the limiter MUST key request budgets by client network identity
- **AND** the system MUST NOT require authenticated user context to apply throttling

#### Scenario: Authenticated route uses current user identity
- **WHEN** the system rate-limits an authenticated route such as datasource validation or audit export/download
- **THEN** the limiter MUST key request budgets by the authenticated user identity when available
- **AND** the system MAY fall back to client network identity only if authenticated identity is unavailable unexpectedly

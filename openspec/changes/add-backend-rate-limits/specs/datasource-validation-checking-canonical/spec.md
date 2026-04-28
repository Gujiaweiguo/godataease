## ADDED Requirements

### Requirement: Datasource Validation Routes Must Enforce Request Limits
The system SHALL enforce request-throttling on canonical datasource validation routes so validation cannot be abused as a high-frequency probe surface.

#### Scenario: Datasource validate by payload remains available within request budget
- **WHEN** an authenticated client calls `POST /api/ds/validate` within the configured datasource-validation request budget
- **THEN** the system MUST preserve the existing datasource validation behavior and response envelope
- **AND** the request MUST still be subject to the existing authentication and validation checks

#### Scenario: Datasource validate by ID remains available within request budget
- **WHEN** an authenticated client calls `GET /api/ds/validate/:id` within the configured datasource-validation request budget
- **THEN** the system MUST preserve the existing datasource validation-by-ID behavior and response envelope

#### Scenario: Datasource validation exceeds request budget
- **WHEN** an authenticated client exceeds the configured request budget for datasource validation routes within the active rate-limit window
- **THEN** the system MUST reject the validation request with an explicit throttling response
- **AND** the throttled request MUST NOT proceed into datasource validation execution

## ADDED Requirements

### Requirement: Audit Export Flows Must Enforce Request Limits
The system SHALL enforce request-throttling on audit export and export-download routes so authenticated users cannot repeatedly generate or fetch audit exports at unbounded frequency.

#### Scenario: Audit export remains available within request budget
- **WHEN** an authenticated client calls `POST /api/audit/export` within the configured audit-export request budget
- **THEN** the system MUST preserve the existing audit export behavior and response envelope
- **AND** the request MUST still return export metadata when business validation succeeds

#### Scenario: Audit export download remains available within request budget
- **WHEN** an authenticated client calls `GET /api/audit/download` within the configured audit-export request budget
- **THEN** the system MUST preserve the existing audit download validation and file-delivery behavior

#### Scenario: Audit export or download exceeds request budget
- **WHEN** an authenticated client exceeds the configured request budget for audit export or download requests within the active rate-limit window
- **THEN** the system MUST reject the request with an explicit throttling response
- **AND** the throttled request MUST NOT continue into export generation or file delivery

## ADDED Requirements

### Requirement: Audit Export Flows Must Enforce Authorization Boundaries
The system SHALL restrict audit export and export-download routes to authenticated users who hold the governed audit authorization for the audit feature surface.

#### Scenario: Authorized user exports audit logs
- **WHEN** an authenticated user with audit feature authorization calls `POST /api/audit/export`
- **THEN** the system MUST preserve the existing audit export behavior and response envelope
- **AND** the request MUST remain subject to the existing audit export rate limits

#### Scenario: Authorized user downloads audit export
- **WHEN** an authenticated user with audit feature authorization calls `GET /api/audit/download`
- **THEN** the system MUST preserve the existing audit download validation and file-delivery behavior
- **AND** the request MUST remain subject to the existing audit export/download rate limits

#### Scenario: Authenticated user without audit authorization is denied export access
- **WHEN** an authenticated user without the governed audit authorization calls `POST /api/audit/export` or `GET /api/audit/download`
- **THEN** the system MUST reject the request with HTTP `403`
- **AND** the request MUST NOT continue into export generation or file delivery

## ADDED Requirements

### Requirement: Datasource Validation Routes Must Enforce Authorization Boundaries
The system SHALL restrict datasource validation routes to authenticated users with datasource validation authority appropriate to the target being validated.

#### Scenario: Existing datasource validation by ID requires datasource authorization
- **WHEN** an authenticated user calls `GET /api/ds/validate/:id` or an equivalent compatibility alias for a datasource they are authorized to manage
- **THEN** the system MUST preserve the existing datasource validation behavior and response envelope
- **AND** the request MUST remain subject to the existing datasource validation rate limits

#### Scenario: Draft datasource validation requires datasource management authorization
- **WHEN** an authenticated user calls `POST /api/ds/validate` or an equivalent compatibility alias with a draft datasource payload that does not yet resolve to a stable datasource resource ID
- **THEN** the system MUST require datasource management authorization equivalent to the governed datasource feature surface
- **AND** authorized requests MUST preserve the existing datasource validation behavior and response envelope

#### Scenario: User without datasource validation authority is denied
- **WHEN** an authenticated user without the required datasource validation authority calls `POST /api/ds/validate`, `GET /api/ds/validate/:id`, or an equivalent compatibility alias
- **THEN** the system MUST reject the request with HTTP `403`
- **AND** the request MUST NOT continue into datasource validation execution

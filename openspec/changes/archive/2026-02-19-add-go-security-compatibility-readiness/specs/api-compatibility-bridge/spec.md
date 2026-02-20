## ADDED Requirements

### Requirement: Template Java Route Group Compatibility
The Go backend SHALL provide Java-compatible template route groups during migration.

#### Scenario: Call template management route with Java prefix
- **WHEN** a client calls `/api/templateManage/*` during migration
- **THEN** the request is routed to equivalent Go template management capability
- **AND** response semantics (`code/msg/data`) match Java client expectations

#### Scenario: Call template market route with Java prefix
- **WHEN** a client calls `/api/templateMarket/*` during migration
- **THEN** the request is routed to equivalent Go template market capability
- **AND** pagination and metadata fields remain Java-compatible

### Requirement: Compatibility Stub Failure Semantics
Compatibility endpoints that are not implemented SHALL fail explicitly and consistently.

#### Scenario: Invoke unimplemented compatibility endpoint
- **WHEN** a client calls a compatibility endpoint that is marked as not implemented
- **THEN** the backend returns explicit non-success semantics (e.g., HTTP 501 with mapped `code/msg`)
- **AND** the response MUST NOT return silent success with empty placeholder payload

### Requirement: Contract Diff Gate for Critical Compatibility Routes
Migration pipeline SHALL enforce contract diff gates for critical compatibility endpoints.

#### Scenario: Validate Java vs Go contract parity
- **WHEN** contract diff suite runs for critical compatibility routes
- **THEN** status, `code`, and payload semantics are compared against Java baseline
- **AND** release gate fails if parity threshold is not met

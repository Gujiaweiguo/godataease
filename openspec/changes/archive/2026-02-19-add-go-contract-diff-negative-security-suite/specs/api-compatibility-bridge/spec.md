## ADDED Requirements

### Requirement: Unauthorized vs Forbidden Semantic Parity
The contract diff system SHALL enforce semantic parity for authentication and authorization failures between Java and Go APIs.

#### Scenario: Distinguish unauthenticated and unauthorized access
- **WHEN** a request is sent without valid authentication credentials
- **THEN** both Java and Go responses MUST return unauthenticated semantics (`401` + expected error contract)
- **AND** responses MUST NOT be misclassified as authorization failures (`403`)

#### Scenario: Enforce forbidden semantics for insufficient privileges
- **WHEN** an authenticated principal lacks required permission
- **THEN** both Java and Go responses MUST return forbidden semantics (`403` + expected error contract)
- **AND** the gate MUST fail if status or error semantics diverge

### Requirement: Row-Level Access and Column Masking Parity
The contract diff system SHALL validate negative security parity for row-level access control and sensitive column masking.

#### Scenario: Block unauthorized row visibility
- **WHEN** a low-privilege principal queries data outside authorized scope
- **THEN** Java and Go responses MUST both hide unauthorized rows consistently
- **AND** any unauthorized row exposure MUST be classified as blocking failure

#### Scenario: Enforce sensitive column masking consistency
- **WHEN** a principal without sensitive-field permission accesses protected columns
- **THEN** Java and Go responses MUST apply equivalent masking semantics
- **AND** unmasked sensitive values in either side MUST fail the gate

### Requirement: Export and Download Authorization Negative Suite
The contract diff system SHALL validate authorization controls for export and download flows under negative security scenarios.

#### Scenario: Reject expired or invalid token download
- **WHEN** export/download APIs are called with expired or invalid credentials
- **THEN** Java and Go responses MUST reject requests with equivalent security semantics
- **AND** successful access in either side MUST be treated as blocking regression

#### Scenario: Reject cross-user or cross-tenant artifact access
- **WHEN** a principal tries to download artifacts owned by another principal or tenant
- **THEN** Java and Go responses MUST deny access consistently
- **AND** the gate MUST archive scenario evidence and fail on any bypass

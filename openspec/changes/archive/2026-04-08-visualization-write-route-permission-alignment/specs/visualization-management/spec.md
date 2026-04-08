## MODIFIED Requirements

### Requirement: Visualization Core CRUD
The system SHALL provide visualization core CRUD capability in Go backend, and in-scope visualization write operations MUST preserve governed authorization semantics before mutating dashboard or big-screen resources.

#### Scenario: Create or update visualization
- **WHEN** client submits visualization definition payload
- **THEN** the system persists visualization metadata and content
- **AND** returns success with Java-compatible response envelope

#### Scenario: Query visualization detail
- **WHEN** client requests visualization detail by identifier
- **THEN** the system returns complete visualization definition for rendering

#### Scenario: Governed visualization write succeeds with sufficient permission
- **WHEN** an authenticated user invokes an in-scope visualization write route with the required governed authorization on the target visualization resource or governed parent scope
- **THEN** the backend MUST allow the mutation to proceed through the normal visualization service flow
- **AND** the resulting write behavior MUST remain compatible with existing visualization CRUD contracts

#### Scenario: Governed visualization write remains explicit when permission is insufficient
- **WHEN** a user invokes an in-scope visualization write route without sufficient governed authorization on the target resource or parent scope
- **THEN** the backend MUST return explicit authorization-denied semantics
- **AND** the failure MUST remain distinguishable from malformed payload, missing visualization resource, or generic route failure

#### Scenario: Visualization write route cannot resolve a safe authorization target
- **WHEN** an in-scope visualization write request does not provide the governed resource or parent information required to authorize that write safely
- **THEN** the backend MUST fail closed before mutating visualization state
- **AND** the system MUST NOT continue into an Auth-only write path as a fallback

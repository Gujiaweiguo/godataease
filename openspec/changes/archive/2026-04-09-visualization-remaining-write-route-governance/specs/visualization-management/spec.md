## MODIFIED Requirements

### Requirement: Visualization Core CRUD
The system SHALL provide visualization core CRUD capability in Go backend, and in-scope visualization write operations MUST preserve governed authorization semantics before mutating dashboard or big-screen resources.

#### Scenario: Remaining root visualization mutation route succeeds with sufficient permission
- **WHEN** an authenticated user invokes `/dataVisualization/updateBase`, `/dataVisualization/move`, `/dataVisualization/updatePublishStatus`, or `/dataVisualization/recoverToPublished` with sufficient governed authorization on the target visualization resource
- **THEN** the backend MUST allow the mutation to proceed through the normal visualization service flow
- **AND** the resulting write behavior MUST remain compatible with existing visualization CRUD contracts

#### Scenario: Remaining root visualization mutation route remains explicit when permission is insufficient
- **WHEN** a user invokes one of the remaining in-scope root visualization mutation routes without sufficient governed authorization on the target visualization resource
- **THEN** the backend MUST return explicit authorization-denied semantics
- **AND** the failure MUST remain distinguishable from malformed payload, missing visualization resource, or generic route failure

#### Scenario: Remaining root visualization mutation route fails closed when the governed target cannot be resolved
- **WHEN** one of the remaining in-scope root visualization mutation routes cannot provide the governed visualization resource information required for safe authorization
- **THEN** the backend MUST fail closed before mutating visualization state
- **AND** the system MUST NOT continue into an Auth-only write path as a fallback

## MODIFIED Requirements

### Requirement: Legacy `/user/org/option` endpoint returns user-option payload semantics
- **WHEN** a client invokes compatibility endpoint `/user/org/option`
- **THEN** the response MUST preserve user-option payload semantics expected by legacy consumers
- **AND** the endpoint MUST remain aligned with user-option behavior instead of returning organization-list payload semantics
- **AND** the route family MUST be classified explicitly as either permanent shim, frontend migration target, or dual-support transition path

### Requirement: Organization-Scoped User Membership Baseline
The system SHALL define user lifecycle operations against an explicit organization-scoped membership baseline that can be reused by later role and permission changes.

#### Scenario: Administrator creates user within organization scope
- **WHEN** an administrator creates a user for a target organization
- **THEN** the system MUST persist the user's organization-scoped membership baseline required by downstream role assignment
- **AND** the create workflow MUST validate the request against the active governed organization context
- **AND** later role workflows MUST be able to discover that user through the same organization scope

#### Scenario: Administrator queries users for organization administration
- **WHEN** an administrator opens a user list under a given organization context
- **THEN** the system MUST return users according to that organization scope
- **AND** the result MUST remain consistent with the organization context established by foundation bootstrap

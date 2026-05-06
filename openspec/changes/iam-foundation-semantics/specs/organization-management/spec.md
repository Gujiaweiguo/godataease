## MODIFIED Requirements

### Requirement: Organization Data Isolation
The system MUST treat organization context as the authoritative runtime scope for governed IAM administration and downstream resource access.

#### Scenario: User from Org A cannot access Org B resources
- **WHEN** user from Organization A attempts to access dashboard from Organization B
- **THEN** system checks resource's organization
- **AND** system denies access if organizations don't match
- **AND** system displays a governed permission-denied error instead of silently crossing organization scope

#### Scenario: Governed IAM write uses current organization context
- **WHEN** an administrator creates, edits, or deletes governed users or roles under an active organization context
- **THEN** the service-layer workflow MUST validate the target organization against the active governed context
- **AND** writes outside that scope MUST be rejected deterministically

### Requirement: Organization Delete Resource Disposition Policy
The system MUST enforce a deterministic organization delete policy that is consistent with documented behavior and existing child-organization safety constraints.

#### Scenario: Delete organization with child organizations
- **WHEN** an administrator attempts to delete an organization that still has children
- **THEN** the system MUST reject the request
- **AND** the response MUST include a clear dependency reason

#### Scenario: Delete leaf organization under C1 policy
- **WHEN** an administrator deletes an organization without child organizations
- **THEN** the system MUST soft-delete the organization record
- **AND** the system MUST preserve audit visibility for affected resources and memberships
- **AND** resource disposition follow-up MUST remain explicitly deferred to downstream lifecycle work rather than being treated as implicitly complete

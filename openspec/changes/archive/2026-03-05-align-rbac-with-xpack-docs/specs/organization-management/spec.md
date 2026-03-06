## ADDED Requirements

### Requirement: Organization Delete Resource Disposition Policy
The system SHALL enforce a deterministic organization delete policy that is consistent with documented behavior and existing child-organization safety constraints.

#### Scenario: Delete organization with child organizations
- **WHEN** an administrator attempts to delete an organization that still has children
- **THEN** the system MUST reject the request
- **AND** the response MUST include a clear dependency reason

#### Scenario: Delete leaf organization
- **WHEN** an administrator deletes an organization without child organizations
- **THEN** the system MUST execute the configured resource disposition policy for that organization
- **AND** affected resources MUST be traceable through audit logs

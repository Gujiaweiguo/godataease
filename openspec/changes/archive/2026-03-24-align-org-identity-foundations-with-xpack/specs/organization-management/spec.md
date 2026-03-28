## ADDED Requirements

### Requirement: Organization Context and Switching Contract
The system SHALL treat organization context as a first-class runtime contract for identity-aware administration flows.

#### Scenario: Current user has access to multiple organizations
- **WHEN** the system returns current-user runtime context after login or refresh
- **THEN** the response MUST include the current organization identity
- **AND** the response MUST include the set of organizations the user is allowed to switch into

#### Scenario: User switches current organization
- **WHEN** a user selects another authorized organization
- **THEN** the system MUST update the current organization context deterministically
- **AND** downstream role and permission workflows MUST consume the new organization context instead of stale scope

## ADDED Requirements

### Requirement: Built-in Role Model Baseline
The system SHALL define a stable built-in role baseline that distinguishes global system roles from organization-scoped default roles before advanced role workflows are introduced.

#### Scenario: Organization becomes available for IAM workflows
- **WHEN** an organization is initialized for governed IAM administration
- **THEN** the system MUST expose the built-in organizational role baseline required by downstream role workflows
- **AND** those built-in roles MUST be discoverable through shared role queries under organization scope

#### Scenario: System-level role is queried alongside organization roles
- **WHEN** an administrator queries roles under a given runtime context
- **THEN** the system MUST distinguish immutable global system roles from organization-scoped built-in roles
- **AND** later role workflow and permission changes MUST consume the same role classification semantics

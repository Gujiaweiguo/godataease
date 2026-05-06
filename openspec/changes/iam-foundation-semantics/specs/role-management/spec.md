## MODIFIED Requirements

### Requirement: Last-Role Policy Must Be Deterministic and Explicitly Recorded
The system MUST define and document a deterministic last-role policy before governed role-member removal is considered complete.

#### Scenario: Administrator removes a user's last governed role
- **WHEN** an administrator removes the user's last remaining role under the governed semantics
- **THEN** the system MUST reject the removal instead of deleting the user automatically
- **AND** the block behavior MUST be documented as an intentional deviation from the manual's cascade-delete implication
- **AND** later user-role lifecycle work MUST reuse the same policy rather than redefining it implicitly

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

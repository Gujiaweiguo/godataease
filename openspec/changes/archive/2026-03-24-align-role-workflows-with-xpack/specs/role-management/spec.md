## ADDED Requirements

### Requirement: Governed Role Workflow Boundary
The system SHALL keep role workflows focused on role lifecycle and member operations without absorbing unified permission-center responsibilities.

#### Scenario: Administrator manages role lifecycle
- **WHEN** an administrator performs create, edit, enable, disable, or delete actions on a role
- **THEN** the system MUST apply organization-aware role lifecycle rules
- **AND** authorization-center concerns MUST remain outside this workflow boundary

#### Scenario: Administrator saves role changes before permission-center rollout
- **WHEN** a role workflow is completed without changing role permission assignments
- **THEN** the role operation MUST remain valid and auditable on its own
- **AND** the system MUST NOT require the unified permission center to complete the lifecycle action

### Requirement: Role Membership Workflow Distinguishes Organization and External Users
The system SHALL provide distinct governed flows for adding organization users and external users into role membership.

#### Scenario: Add organization user into role
- **WHEN** an administrator selects users already available under the current organization scope
- **THEN** the system MUST create missing role-member relations idempotently
- **AND** the operation MUST remain bound to the active organization context

#### Scenario: Add external user into role
- **WHEN** an administrator searches a user outside the default organization list by exact account or identifier
- **THEN** the system MUST validate that the assignment is allowed for the target organization scope
- **AND** the operation MUST emit auditable role-member change evidence

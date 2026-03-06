## ADDED Requirements

### Requirement: Role Member Lifecycle Management
The system SHALL provide complete role member lifecycle APIs and UI workflows, including add organization user, add external user, and remove member from role.

#### Scenario: Add existing organization user to role
- **WHEN** an administrator selects users already under the current organization and confirms assignment
- **THEN** the system MUST create missing user-role associations idempotently
- **AND** duplicate associations MUST NOT be created

#### Scenario: Add external user by account or ID
- **WHEN** an administrator searches an external user by exact account or user ID and assigns role
- **THEN** the system MUST create user-role association within target organization scope
- **AND** the operation MUST be audited

### Requirement: Last Role Safety Guard
The system SHALL enforce last-role safety semantics when removing a role-member relation.

#### Scenario: Remove member with multiple remaining roles
- **WHEN** an administrator removes one role from a user who still has other roles
- **THEN** the system MUST remove only the target relation
- **AND** the user account MUST remain active

#### Scenario: Remove member with unique remaining role
- **WHEN** an administrator removes the user's last role in all organizations
- **THEN** the system MUST execute configured safety policy
- **AND** the resulting behavior MUST be deterministic and documented

### Requirement: Custom Role Inheritance Constraint
The system SHALL enforce that custom roles inherit from allowed built-in organizational roles and cannot exceed inherited permission boundaries.

#### Scenario: Create custom role inheriting organization admin
- **WHEN** an administrator creates a custom role with inheritance base "organization-admin"
- **THEN** the system MUST validate inherited scope and cap grantable permissions to parent bounds

#### Scenario: Attempt to exceed parent permissions
- **WHEN** an administrator assigns permissions beyond inherited parent scope
- **THEN** the system MUST reject the operation with a validation error

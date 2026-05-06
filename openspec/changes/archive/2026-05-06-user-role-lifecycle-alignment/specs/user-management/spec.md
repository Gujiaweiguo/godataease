# user-management Delta Spec for user-role-lifecycle-alignment

> This is a delta spec for the `user-management` capability. Only new or changed requirements are listed below. See `openspec/specs/user-management/spec.md` for the full baseline.

## ADDED Requirements

### Requirement: User-Role Operations Respect Org-Scoped Lifecycle Contracts
The system SHALL ensure that user-role operations surfaced in the user admin page follow the org-scoped lifecycle contracts defined by `user-role-lifecycle`.

#### Scenario: Administrator assigns role to user from user management page
- **WHEN** an administrator assigns a role to a user from the user-management admin surface
- **THEN** the operation MUST use org-scoped transactional assignment as defined by user-role-lifecycle
- **AND** the user management page MUST reflect the updated role state after successful assignment

#### Scenario: Administrator removes role from user via user management
- **WHEN** an administrator removes a role from a user via the user-management page
- **THEN** the system MUST apply the configured last-role safety policy before completing removal
- **AND** the user management page MUST display the policy outcome (blocked, warning, or cascaded)

#### Scenario: Administrator initiates org transfer from user management
- **WHEN** an administrator initiates a user org transfer from the user-management admin surface
- **THEN** the operation MUST follow the explicit transfer contract defined by user-role-lifecycle
- **AND** the user list in the source org MUST reflect the transfer immediately after completion

### Requirement: User Management Displays Last-Role Policy Status
The system SHALL display the effective last-role safety policy status in the user management admin context.

#### Scenario: User management shows policy indicator during role removal
- **WHEN** an administrator views the role removal confirmation for a user who has only one remaining role
- **THEN** the user management page MUST display the effective last-role policy for the current org
- **AND** the confirmation MUST indicate what will happen if the removal proceeds under the current policy

# role-management Delta Spec for user-role-lifecycle-alignment

> This is a delta spec for the `role-management` capability. Only new or changed requirements are listed below. See `openspec/specs/role-management/spec.md` for the full baseline.

## MODIFIED Requirements

### Requirement: Last Role Safety Guard
The system SHALL enforce last-role safety semantics when removing a role-member relation using the configurable three-mode policy defined by `user-role-lifecycle`.

#### Scenario: Remove member with multiple remaining roles
- **WHEN** an administrator removes one role from a user who still has other roles
- **THEN** the system MUST remove only the target relation
- **AND** the user account MUST remain active

#### Scenario: Remove member with unique remaining role under BLOCK policy
- **WHEN** an administrator removes the user's last role and the effective policy is BLOCK
- **THEN** the system MUST reject the removal
- **AND** the system MUST return a governed error indicating last-role safety policy blocked the operation

#### Scenario: Remove member with unique remaining role under WARN+ALLOW policy
- **WHEN** an administrator removes the user's last role and the effective policy is WARN+ALLOW
- **THEN** the system MUST remove the role association
- **AND** the user account MUST remain active with zero roles
- **AND** the system MUST record a warning audit entry

#### Scenario: Remove member with unique remaining role under CASCADE policy
- **WHEN** an administrator removes the user's last role and the effective policy is CASCADE
- **THEN** the system MUST remove the role association and disable the user account
- **AND** the system MUST record a critical audit entry

## ADDED Requirements

### Requirement: Role Member Operations Verify Org Scope Consistency
The system SHALL verify org scope consistency for all role-member add, remove, and query operations.

#### Scenario: Add member validates org scope match
- **WHEN** an administrator adds a user to a role via role-management workflows
- **THEN** the system MUST verify the role and user share the same org scope
- **AND** operations with org scope mismatch MUST be rejected

#### Scenario: Role member list is org-scoped
- **WHEN** an administrator views the member list for a role
- **THEN** the system MUST return only members within the role's org scope
- **AND** cross-org members MUST NOT appear in the list

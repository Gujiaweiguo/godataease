# user-role-lifecycle Specification (NEW)

## Purpose
This capability defines the full user-role assignment lifecycle, including org-scoped role binding, configurable last-role safety policy, explicit user org membership transfer, and org-consistent role-member operations. It extends the C1 foundation semantics without modifying frozen org-delete or permission enforcement contracts.

## Requirements

### Requirement: Configurable Last-Role Safety Policy
The system SHALL provide a configurable last-role safety policy with three modes: BLOCK (default), WARN+ALLOW, and CASCADE.

#### Scenario: Last-role removal with BLOCK policy (default)
- **WHEN** an administrator removes a user's last remaining role and the configured policy is BLOCK
- **THEN** the system MUST reject the removal
- **AND** the system MUST return an error indicating the operation is blocked by last-role safety policy
- **AND** the system MUST record an audit entry with policy mode BLOCK, actor, target user, and affected role

#### Scenario: Last-role removal with WARN+ALLOW policy
- **WHEN** an administrator removes a user's last remaining role and the configured policy is WARN+ALLOW
- **THEN** the system MUST remove the role association
- **AND** the user account MUST remain active with zero assigned roles
- **AND** the system MUST record a warning-level audit entry capturing the policy decision

#### Scenario: Last-role removal with CASCADE policy
- **WHEN** an administrator removes a user's last remaining role and the configured policy is CASCADE
- **THEN** the system MUST remove the role association
- **AND** the system MUST set the user's enabled state to false
- **AND** the user account MUST NOT be hard-deleted
- **AND** the system MUST record a critical-level audit entry capturing the cascade action

#### Scenario: Last-role policy defaults to BLOCK on fresh installation
- **WHEN** the system starts with no previously configured last-role policy
- **THEN** the effective policy MUST be BLOCK
- **AND** no configuration action is required to maintain current behavior

#### Scenario: Administrator queries current last-role policy
- **WHEN** an administrator queries the last-role safety policy for their organization
- **THEN** the system MUST return the current effective policy mode
- **AND** the response MUST include the policy mode name and last-updated timestamp

#### Scenario: Administrator updates last-role policy
- **WHEN** an org-level administrator changes the last-role safety policy
- **THEN** the system MUST validate the requested mode is one of BLOCK, WARN+ALLOW, or CASCADE
- **AND** the system MUST persist the new policy for that organization scope
- **AND** the system MUST record an audit entry capturing the policy change, old mode, new mode, and actor

### Requirement: Org-Scoped Transactional Role Assignment
The system SHALL assign roles to users within an explicit org scope as a single transactional operation.

#### Scenario: Assign role to user within same org scope
- **WHEN** an administrator assigns a role to a user who is already a member of the target org
- **THEN** the system MUST validate org context via the governed org context enforcement
- **AND** the system MUST create the user-role association idempotently within a single DB transaction
- **AND** duplicate associations MUST NOT be created and MUST NOT produce errors

#### Scenario: Assign role to user not yet in target org
- **WHEN** an administrator assigns a role to a user who is not yet a member of the target org
- **THEN** the system MUST create the org membership baseline via the authoritative binding entry point
- **AND** the system MUST create the user-role association within the same transaction
- **AND** both operations MUST succeed or both MUST roll back

#### Scenario: Role assignment respects built-in role baseline
- **WHEN** an administrator assigns a built-in readonly role to a user
- **THEN** the system MUST allow the assignment
- **AND** the system MUST NOT allow modification of the built-in role's own attributes during assignment

#### Scenario: Concurrent role assignments to same user in same org
- **WHEN** two administrators concurrently assign different roles to the same user in the same org
- **THEN** both assignments MUST succeed without deadlock
- **AND** the user MUST end up with both role associations

### Requirement: Explicit User Org Membership Transfer
The system SHALL provide an explicit user org membership transfer operation that moves a user's role bindings from source org to target org atomically.

#### Scenario: Transfer user between organizations
- **WHEN** an org-level administrator transfers a user from source org A to target org B
- **THEN** the system MUST validate that target org B exists and is active
- **AND** the system MUST remove all user-role bindings for the user in source org A
- **AND** the system MUST create baseline membership in target org B
- **AND** the system MUST assign the target org's default built-in role to the user in org B
- **AND** all operations MUST execute within a single DB transaction

#### Scenario: Transfer user to non-existent target org
- **WHEN** an administrator attempts to transfer a user to an org that does not exist or is soft-deleted
- **THEN** the system MUST reject the transfer
- **AND** no user-role bindings in the source org MUST be modified

#### Scenario: Transfer triggers last-role policy in source org
- **WHEN** a user transfer removes the user's last role in the source org
- **THEN** the system MUST apply the configured last-role safety policy for the source org
- **AND** if the policy is BLOCK, the transfer MUST be rejected
- **AND** if the policy is WARN+ALLOW or CASCADE, the transfer MUST proceed with the corresponding side effect

#### Scenario: Transfer audit trail
- **WHEN** a user org transfer completes successfully
- **THEN** the system MUST record a transfer audit entry capturing source org ID, target org ID, user ID, actor ID, assigned role in target org, and timestamp
- **AND** the audit entry MUST be queryable by org administrators of both source and target orgs

### Requirement: Role-Member Org Scope Consistency Enforcement
The system SHALL verify that role-member add, remove, and query operations maintain org scope consistency.

#### Scenario: Add member to role within same org
- **WHEN** an administrator adds a user as a member of a role
- **THEN** the system MUST verify that the role's org scope matches the operation's org context
- **AND** the system MUST verify that the user is a member of the same org scope
- **AND** if both match, the association MUST be created

#### Scenario: Add member to role in different org scope
- **WHEN** an administrator attempts to add a user to a role in a different org scope
- **THEN** the system MUST reject the operation
- **AND** the system MUST return an error indicating org scope mismatch

#### Scenario: Query role members filtered by org scope
- **WHEN** an administrator queries members of a role
- **THEN** the system MUST return only members within the role's org scope
- **AND** members from other org scopes MUST NOT appear in the result

#### Scenario: Remove member from role verifies org scope
- **WHEN** an administrator removes a user from a role
- **THEN** the system MUST verify that the user-role association belongs to the current org scope
- **AND** removal MUST proceed only after org scope verification passes

### Requirement: Lifecycle Audit Trail
The system SHALL record audit entries for all user-role lifecycle operations.

#### Scenario: Role assignment audit
- **WHEN** a role is assigned to a user
- **THEN** the system MUST record an audit entry with operation type, actor ID, user ID, role ID, org ID, and timestamp

#### Scenario: Role removal audit
- **WHEN** a role is removed from a user
- **THEN** the system MUST record an audit entry with operation type, actor ID, user ID, role ID, org ID, last-role policy applied (if applicable), and timestamp

#### Scenario: Org transfer audit
- **WHEN** a user is transferred between orgs
- **THEN** the system MUST record an audit entry with transfer-specific metadata as defined by the transfer requirement

#### Scenario: Policy change audit
- **WHEN** the last-role safety policy is changed
- **THEN** the system MUST record an audit entry with old policy mode, new policy mode, actor ID, org ID, and timestamp

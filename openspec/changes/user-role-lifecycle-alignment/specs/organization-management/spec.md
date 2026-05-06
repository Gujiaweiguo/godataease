# organization-management Delta Spec for user-role-lifecycle-alignment

> This is a delta spec for the `organization-management` capability. Only new or changed requirements are listed below. See `openspec/specs/organization-management/spec.md` for the full baseline.

## ADDED Requirements

### Requirement: Organization User Membership Transfer Is a Governed Operation
The system SHALL treat user membership transfer between organizations as an explicit governed operation with audit trail.

#### Scenario: Administrator transfers user into their organization
- **WHEN** an org-level administrator transfers a user from another organization into their own
- **THEN** the system MUST execute the transfer as a single atomic operation
- **AND** the source org role bindings MUST be removed within the same transaction
- **AND** a default role in the target org MUST be assigned within the same transaction
- **AND** an audit entry MUST record source org, target org, user, actor, and resulting assignment

#### Scenario: Transfer is rejected when target org is inactive or missing
- **WHEN** an administrator attempts to transfer a user into an organization that is soft-deleted or inactive
- **THEN** the system MUST reject the transfer
- **AND** no membership state in any org MUST change

#### Scenario: Transfer respects last-role policy in source org
- **WHEN** a transfer would remove the user's last role in the source organization
- **THEN** the system MUST apply the configured last-role safety policy for the source org
- **AND** if the policy is BLOCK, the transfer MUST be rejected entirely
- **AND** if the policy is WARN+ALLOW or CASCADE, the transfer MUST proceed with the corresponding side effect on the user's source-org membership

### Requirement: Organization Membership Transfer Requires Audit Visibility
The system SHALL make membership transfer audit entries visible to administrators of both source and target organizations.

#### Scenario: Source org admin queries transfer history
- **WHEN** an administrator of the source org queries user transfer audit entries
- **THEN** the system MUST return transfer entries where their org appears as source
- **AND** each entry MUST include target org ID, user ID, actor, and timestamp

#### Scenario: Target org admin queries transfer history
- **WHEN** an administrator of the target org queries user transfer audit entries
- **THEN** the system MUST return transfer entries where their org appears as target
- **AND** each entry MUST include source org ID, user ID, actor, and timestamp

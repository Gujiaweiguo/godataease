## ADDED Requirements

### Requirement: Permission Dual-Perspective Consistency
The system SHALL provide both "configure by user" and "configure by resource" views, and both views SHALL persist to the same authorization model with equivalent effective results.

#### Scenario: Grant permission in user perspective
- **WHEN** an administrator grants a resource permission in user perspective
- **THEN** the same grant MUST be visible in resource perspective without additional synchronization steps

#### Scenario: Revoke permission in resource perspective
- **WHEN** an administrator revokes a grant in resource perspective
- **THEN** the same revocation MUST be visible in user perspective immediately after data refresh

### Requirement: Resource Group Inheritance Effective on New Resources
The system SHALL apply inherited permissions from resource groups to newly created resources in that group.

#### Scenario: Create resource under granted group
- **WHEN** a new dashboard/dataset is created under a group that already has grants
- **THEN** the new resource MUST inherit effective grants for all authorized users/roles

#### Scenario: Query inherited permission
- **WHEN** a client checks resource permission for an inherited target
- **THEN** the system MUST return permission as granted without requiring manual re-authorization

## MODIFIED Requirements

### Requirement: Permission Dual-Perspective Consistency

The system SHALL provide both "configure by user" and "configure by resource" views, and both views SHALL persist to the same authorization model with equivalent effective results.

The system SHALL implement `CheckPermissionConsistency` as a real cross-view verification that queries both perspectives from the database and reports any divergences, rather than returning a hardcoded always-true result.

#### Scenario: Grant permission in user perspective
- **WHEN** an administrator grants a resource permission in user perspective
- **THEN** the same grant MUST be visible in resource perspective without additional synchronization steps

#### Scenario: Revoke permission in resource perspective
- **WHEN** an administrator revokes a grant in resource perspective
- **THEN** the same revocation MUST be visible in user perspective immediately after data refresh

#### Scenario: Consistency check detects no divergence
- **WHEN** `CheckPermissionConsistency` is called on a system where user-view and resource-view permissions are identical
- **THEN** the method SHALL return `Consistent: true` with accurate `UserCount` and `ResourceCount` reflecting the number of users and resources scanned
- **AND** `Inconsistencies` SHALL be empty

#### Scenario: Consistency check detects divergence
- **WHEN** `CheckPermissionConsistency` is called and a permission `(userID, permKey)` pair exists in one perspective but not the other
- **THEN** the method SHALL return `Consistent: false`
- **AND** `Inconsistencies` SHALL contain at least one entry with `UserID`, `ResourceID`, `ResourceType`, `UserView`, and `ResourceView` describing the divergence

#### Scenario: Consistency check on empty system
- **WHEN** `CheckPermissionConsistency` is called on a system with no users, no resources, or no permissions
- **THEN** the method SHALL return `Consistent: true` with `UserCount: 0` and `ResourceCount: 0`

#### Scenario: Consistency check skips oversized user base
- **WHEN** `CheckPermissionConsistency` is called and the active user count exceeds 10,000
- **THEN** the method SHALL return `Consistent: true` and the result SHALL indicate that the check was skipped for performance reasons
- **AND** the method SHALL NOT perform the full cross-view scan

## MODIFIED Requirements

This is a delta spec that `openspec/changes/enhance-role-pagination/specs/role-management/spec.md`.

### Requirement: Role Pagination Query
The system SHALL provide a paginated role query API with total count.

#### Scenario: Query roles with pagination
- **WHEN** admin queries roles with `current=1, size=10`
- **THEN** system returns first page of roles
- **AND** total count of matching roles
- **AND** pagination metadata (current, size)

#### Scenario: Query roles with keyword filter and pagination
- **WHEN** admin queries roles with keyword "admin" and `current=1, size=10`
- **THEN** system returns roles containing "admin" in name
- **AND** total count reflects filtered results
#### Scenario: Query roles with role type filter
- **WHEN** admin queries roles with roleType="system" then `current=1, size=10`
- **THEN** system returns only system roles
- **AND** excludes custom roles


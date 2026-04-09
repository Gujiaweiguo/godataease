## MODIFIED Requirements

### Requirement: Role-based menu authorization remains available from the unified center
The system SHALL let administrators manage menu authorization for roles from the permission configuration center.

#### Scenario: Menu permission tab uses canonical role-menu authorization APIs for role-scoped load/save
- **WHEN** an administrator selects a role and queries or saves menu authorization from the menu permission tab
- **THEN** the workflow MUST use canonical role-menu authorization APIs
- **AND** the role-scoped menu authorization result MUST remain consistent with previously saved state

## ADDED Requirements

### Requirement: Menu Authorization Drives Navigation Visibility
The system SHALL bind role-menu authorization outcomes to navigation visibility decisions.

#### Scenario: Grant menu to role
- **WHEN** an administrator grants a menu node to a role
- **THEN** users with that role MUST see the menu in navigation after authorization refresh

#### Scenario: Revoke menu from role
- **WHEN** an administrator revokes a previously granted menu node from a role
- **THEN** users with that role MUST no longer see the menu in navigation
- **AND** direct navigation to revoked route MUST be denied

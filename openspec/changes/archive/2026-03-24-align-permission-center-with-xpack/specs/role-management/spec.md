## ADDED Requirements

### Requirement: Role Authorization Carrier Boundary
The system SHALL treat roles as authorization carriers inside the unified permission center without moving role lifecycle or member-management responsibilities into permission workflows.

#### Scenario: Administrator grants menu or resource authorization to role
- **WHEN** an administrator configures authorization for a role inside the permission center
- **THEN** the system MUST use the role as the persisted authorization carrier
- **AND** the workflow MUST NOT require redefining the role lifecycle or membership semantics

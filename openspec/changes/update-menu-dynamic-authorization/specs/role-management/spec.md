## ADDED Requirements
### Requirement: Role Menu Authorization Configuration
The system SHALL allow administrators to configure menu authorizations as part of role management workflows.

#### Scenario: Configure menus during role administration
- **WHEN** an administrator edits a role in role management
- **THEN** the administrator can view and update the role's granted menu set
- **AND** saved menu grants take effect in subsequent authorized menu queries

#### Scenario: New role bootstrap without menu grants
- **WHEN** an administrator creates a new role without explicit menu grants
- **THEN** the role starts with no business menu visibility by default
- **AND** access remains denied until menu grants are explicitly assigned

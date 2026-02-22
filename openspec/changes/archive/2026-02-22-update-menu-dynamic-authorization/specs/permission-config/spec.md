## ADDED Requirements
### Requirement: Role-Menu Authorization Mapping
The system SHALL persist role-to-menu authorization mappings and use them as the authoritative source for menu visibility decisions.

#### Scenario: Grant menu set to role
- **WHEN** an administrator saves menu assignments for a role
- **THEN** the system persists role-menu relations idempotently
- **AND** users with that role receive only granted menus in authorized menu responses

#### Scenario: Revoke menu from role
- **WHEN** an administrator revokes one or more menus from a role
- **THEN** users with that role lose visibility to revoked menus on next authorization fetch
- **AND** direct access to revoked menu routes is denied by authorization policy

### Requirement: Role-Menu Authorization APIs
The system SHALL provide APIs to query and save role-menu authorization state.

#### Scenario: Query role-menu assignments
- **WHEN** a client requests menu assignments for a role
- **THEN** the system returns granted menu IDs and metadata needed for authorization UI

#### Scenario: Save role-menu assignments
- **WHEN** a client submits a complete role-menu assignment set
- **THEN** the system validates role and menu existence before persistence
- **AND** the system returns success only after effective authorization state is stored

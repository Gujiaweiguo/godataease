## MODIFIED Requirements

This is a delta spec that `openspec/specs/permission-config/spec.md`.

### Requirement: Menu Permission Control
The system SHALL provide menu permission management including:
- Define menu structure and hierarchy
- Assign menu permissions to roles
- Control access to functional modules
- Menu permissions only assignable to roles, not directly to users

#### Scenario: Admin assigns menu permission to role
- **WHEN** administrator assigns a menu permission to a role
- **THEN** users with that role can access the corresponding module
- **AND** the corresponding menu item is visible to those users
- **AND** other menu items without permissions remain hidden

#### Scenario: User tries to access unauthorized menu
- **WHEN** user without menu permission tries to access a protected menu URL directly
- **THEN** system denies access
- **AND** system displays an insufficient-permission error or redirect behavior
- **AND** system MUST NOT misclassify the authorization failure as a generic `404`

### Requirement: Role-Menu Authorization Mapping
The system SHALL persist role-to-menu authorization mappings and use them as the authoritative source for menu visibility decisions.

#### Scenario: Grant menu set to role
- **WHEN** an administrator saves menu assignments for a role
- **THEN** the system persists role-menu relations idempotently
- **AND** users with that role receive only granted menus in authorized menu responses
- **AND** direct route access decisions remain consistent with the same effective authorization state

#### Scenario: Revoke menu from role
- **WHEN** an administrator revokes one or more menus from a role
- **THEN** users with that role lose visibility to revoked menus on next authorization fetch
- **AND** direct access to revoked menu routes is denied by authorization policy
- **AND** revoked access MUST NOT degrade into a misleading `404` caused by permission-path confusion

### Requirement: Role Permission Save API Availability
The system SHALL provide role permission save APIs required by frontend role management page.

#### Scenario: Save resource permissions for role
- **WHEN** frontend submits role permission save request to `/system/role/permission/save`
- **THEN** backend MUST process the request through Go permission service path
- **AND** return success or explicit failure without `404`

### Requirement: Menu and Business Permission Query APIs
The system SHALL provide menu-permission and business-permission query APIs used by permission configuration UI.

#### Scenario: Query menu and business permission trees
- **WHEN** frontend requests `/auth/menuPermission` or `/auth/busiPermission`
- **THEN** backend MUST return permission dataset in contract-compatible structure
- **AND** the request MUST NOT fallback to static file route or generic `404`

### Requirement: Menu and Business Permission Save APIs
The system SHALL provide save APIs for menu and business permission assignments.

#### Scenario: Persist permission assignment changes
- **WHEN** frontend posts permission updates to `/auth/saveMenuPer` or `/auth/saveBusiPer`
- **THEN** backend MUST persist effective authorization state or return explicit validation/auth error
- **AND** MUST NOT return placeholder success for unimplemented logic
- **AND** MUST NOT return generic `404` for supported permission flows

### Requirement: Authorization Result Must Be Semantically Distinguishable
The system SHALL keep authorization denial distinguishable from missing resource errors across menu access flows.

#### Scenario: Protected resource exists but user lacks permission
- **WHEN** a protected page or API exists but the current user lacks permission
- **THEN** the system MUST return authorization-denied behavior
- **AND** frontend and backend logs SHOULD preserve that semantic distinction
- **AND** operators MUST be able to distinguish this case from resource absence during troubleshooting

#### Scenario: Protected resource does not exist
- **WHEN** frontend requests a page or API that is not implemented or no longer registered
- **THEN** the system MAY return `404`
- **AND** the result MUST remain distinguishable from authorization denial

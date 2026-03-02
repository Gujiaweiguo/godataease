## ADDED Requirements

### Requirement: Menu Visibility Controlled by Role Authorization
Menu visibility in API responses SHALL be controlled by role-menu authorization mappings.

#### Scenario: User sees only authorized menus
- **WHEN** user requests menu tree or routes
- **THEN** only menus assigned to user's roles are included
- **AND** unauthorized menus are completely excluded from response

#### Scenario: User with no role gets empty menu
- **WHEN** user without any role requests menu
- **THEN** system returns empty menu list
- **AND** response status is 200 (not 500 or 403)

### Requirement: Unauthorized Menu Direct Access Denied
Direct access to unauthorized menu routes SHALL be denied.

#### Scenario: Direct URL access to unauthorized menu
- **WHEN** user navigates directly to URL of unauthorized menu
- **THEN** system denies access with 403 semantics
- **AND** displays permission denied message

### Requirement: Menu Authorization Changes Immediate Effect
Menu authorization changes SHALL take effect on next API call.

#### Scenario: Role menu assignment change reflects immediately
- **WHEN** admin changes role menu assignments
- **THEN** affected users see updated menus on next menu API call
- **AND** no server restart or cache clear required

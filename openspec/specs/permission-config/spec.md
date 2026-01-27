# permission-config Specification

## Purpose
TBD - created by archiving change add-user-org-permission. Update Purpose after archive.
## Requirements
### Requirement: Row-Level Permission Filtering
The system SHALL support row-level permission filtering for dataset queries based on
configured row permission rules.

#### Scenario: Apply row permission filters to a query
- **WHEN** a user executes a dataset query with row permission rules configured
- **THEN** the system applies those rules to the query
- **THEN** rows outside the allowed scope are excluded from results

### Requirement: Menu Permission Control
The system SHALL provide menu permission management including:
- Define menu structure and hierarchy
- Assign menu permissions to roles
- Control access to functional modules (workspace, dashboard, screen, dataset, datasource, system management)
- Menu permissions only assignable to roles, not directly to users

#### Scenario: Admin assigns dashboard menu permission to role
- **WHEN** administrator assigns "Dashboard" menu permission to "Viewer" role
- **THEN** all users with "Viewer" role can access dashboard module
- **THEN** dashboard menu item is visible to those users
- **THEN** other menu items without permissions remain hidden

#### Scenario: User tries to access unauthorized menu
- **WHEN** user without "User Management" menu permission tries to access user management URL directly
- **THEN** system denies access
- **THEN** system displays "Insufficient permissions" error
- **THEN** system redirects to previous page or home page

### Requirement: Resource Permission Control
The system SHALL provide resource permission management including:
- Define resource groups (datasources, datasets, dashboards, screens)
- Grant view, edit, export, manage permissions on resources
- Support permission inheritance from resource groups
- Assign resource permissions to users or roles
- Separate resource permissions from menu permissions

#### Scenario: Admin grants dashboard edit permission to user
- **WHEN** administrator grants "Edit" permission on specific dashboard to user
- **THEN** user can modify the dashboard
- **THEN** user cannot delete the dashboard without "Manage" permission
- **THEN** user cannot export dashboard without "Export" permission

#### Scenario: Admin grants datasource view permission to role
- **WHEN** administrator grants "View" permission on datasource group to "Data Analyst" role
- **THEN** all users with "Data Analyst" role can view all datasources in group
- **THEN** any new datasource added to group automatically inherits permission
- **THEN** users cannot edit datasources without explicit "Edit" permission

#### Scenario: User attempts to edit resource without permission
- **WHEN** user with "View" permission only tries to edit a dataset
- **THEN** system denies edit operation
- **THEN** system displays "Permission denied" error
- **THEN** system logs permission violation

### Requirement: Column Permission Control
The system SHALL support column-level permissions with the following modes:
- Disable: hide column entirely from query results
- Mask: replace sensitive values with masking rules

#### Scenario: Column is disabled for a user
- **WHEN** a user queries a dataset with a disabled column
- **THEN** the column is omitted from the query results

#### Scenario: Column is masked for a user
- **WHEN** a user queries a dataset with a masked column
- **THEN** the column is returned with values masked per configured rule

### Requirement: Export Permission Control
The system SHALL provide granular export permission control including:
- Resource export permission: control export of images, PDFs, templates
- Chart export permission: control export of chart data as Excel
- Detail export permission: control export of detailed data

#### Scenario: User with view permission cannot export
- **WHEN** user with only "View" permission tries to export dashboard as PDF
- **THEN** system denies export operation
- **THEN** system displays "No export permission" error
- **THEN** system hides export buttons if user lacks export permissions

### Requirement: Datasource View-Only Permission
The system SHALL support view-only permission for datasources with the following constraints:
- Allow viewing datasource basic info (name, description, connection type, table structure)
- Allow reading data through datasets (if dataset has public permission)
- Prohibit editing datasource configuration
- Prohibit creating datasets based on datasource
- Hide datasource from "Create Dataset" interface if no view permission
- Prevent editing existing datasets that depend on datasource without view permission

#### Scenario: User with datasource view permission creates dataset
- **WHEN** user with datasource "View" permission only tries to create dataset
- **THEN** "Create Dataset" interface hides datasource without view permission
- **THEN** user cannot select datasource without view permission
- **THEN** system prevents dataset creation if datasource lacks permission

#### Scenario: User edits dataset dependent on unauthorized datasource
- **WHEN** user with dataset edit permission tries to edit dataset
- **THEN** system checks if dependent datasource has view permission
- **THEN** system denies save operation if no permission
- **THEN** system displays "Insufficient permissions, unable to modify" error

### Requirement: Permission Configuration Perspectives
The system SHALL support two perspectives for permission configuration:
- "Configure by User" view: assign permissions to individual users
- "Configure by Resource" view: assign users/roles to resources
- Underlying model is the same, different presentation

#### Scenario: Admin configures permissions by user
- **WHEN** administrator selects "Configure by User" view
- **THEN** system displays list of users
- **THEN** clicking user shows all permissions (menus, resources, data)
- **THEN** administrator can toggle permissions for that user

#### Scenario: Admin configures permissions by resource
- **WHEN** administrator selects "Configure by Resource" view
- **THEN** system displays resource tree
- **THEN** clicking resource shows users/roles with access
- **THEN** administrator can add/remove users/roles

### Requirement: Permission Inheritance
The system SHALL support permission inheritance including:
- Resource groups automatically grant permissions to all resources within group
- New resources automatically inherit parent group permissions
- Role inheritance from parent roles if hierarchy exists

#### Scenario: New dashboard inherits group permissions
- **WHEN** administrator creates new dashboard under "Production Dashboards" group
- **THEN** dashboard automatically inherits group's permissions
- **THEN** users with group permission can access new dashboard
- **THEN** administrator can override inherited permissions if needed


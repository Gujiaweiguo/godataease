# permission-config Specification

## Purpose
This capability provides comprehensive permission configuration for DataEase, enabling fine-grained access control across menus, resources, data rows, and columns. It supports permission inheritance, role-based and user-based assignment, and granular export controls to meet enterprise security requirements for multi-tenant environments.
## Requirements
### Requirement: Row-Level Permission Filtering

系统 SHALL 在 Go 实现中保持与 Java 版本完全一致的行级权限过滤逻辑。

#### Scenario: Permission SQL generation
- **WHEN** 生成行级权限过滤 SQL
- **THEN** Go 实现 SHALL 生成与 Java 版本语法等效的 WHERE 子句

#### Scenario: Permission filter result
- **WHEN** 执行带权限过滤的查询
- **THEN** Go 实现 SHALL 返回与 Java 版本完全相同的过滤结果

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

系统 SHALL 在 Go 实现中保持与 Java 版本完全一致的列级权限控制。

#### Scenario: Column is disabled for a user
- **WHEN** a user queries a dataset with a disabled column
- **THEN** Go 实现 SHALL 以与 Java 版本相同的方式从结果中隐藏该列

#### Scenario: Column is masked for a user
- **WHEN** a user queries a dataset with a masked column
- **THEN** Go 实现 SHALL 以与 Java 版本相同的方式进行脱敏处理

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

### Requirement: Permission Cache Consistency

系统 SHALL 在 Go 实现中保持与 Java 版本相同的权限缓存行为。

#### Scenario: Cache invalidation
- **WHEN** 权限配置变更
- **THEN** Go 实现 SHALL 以与 Java 版本相同的方式失效相关缓存

### Requirement: Permission Repository Pattern

系统 SHALL 使用 Repository 模式实现数据访问层。

#### Scenario: Permission repository interface
- **WHEN** 实现权限数据访问
- **THEN** 系统 SHALL 提供 PermRepository 接口，包含 Create、Update、Delete、GetByID、GetByKey、List 方法

#### Scenario: GORM entity mapping
- **WHEN** 定义权限实体
- **THEN** Go 实现 SHALL 使用 GORM 标签映射数据库字段（`gorm:"column:perm_id"`）

### Requirement: Permission Service Layer

系统 SHALL 使用 Service 层封装业务逻辑。

#### Scenario: Permission service interface
- **WHEN** 实现权限业务逻辑
- **THEN** 系统 SHALL 提供 PermService 接口，包含 CreatePerm、UpdatePerm、DeletePerm、ListPerms 方法

#### Scenario: Permission key uniqueness
- **WHEN** 创建或更新权限
- **THEN** Service 层 SHALL 验证权限标识（perm_key）唯一性

### Requirement: Permission Handler REST Mapping

系统 SHALL 提供 HTTP Handler 映射 REST API 到 Service 层。

#### Scenario: POST /api/system/permission/list
- **WHEN** 客户端请求 POST /api/system/permission/list
- **THEN** Handler SHALL 解析请求参数（current, size）
- **THEN** Handler SHALL 调用 Service 层查询
- **THEN** Handler SHALL 返回分页结果 `{code, data: {list, total, current, size}, msg}`

#### Scenario: POST /api/system/permission/create
- **WHEN** 客户端请求 POST /api/system/permission/create
- **THEN** Handler SHALL 解析请求体（permName, permKey, permType, permDesc）
- **THEN** Handler SHALL 调用 Service 层创建权限
- **THEN** Handler SHALL 返回 `{code: "000000", data: permId, msg: "success"}`

#### Scenario: Error response format
- **WHEN** 权限操作失败
- **THEN** Handler SHALL 返回 `{code: "500000", msg: "Failed: <error message>"}`

### Requirement: Permission Type Constants

系统 SHALL 定义权限类型常量。

#### Scenario: Permission types
- **WHEN** 定义权限类型
- **THEN** 系统 SHALL 提供 PermTypeMenu（菜单）、PermTypeButton（按钮）、PermTypeData（数据）常量


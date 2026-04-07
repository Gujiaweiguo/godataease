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
- **AND** system MUST NOT misclassify the authorization failure as a generic `404`

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

### Requirement: Permission configuration becomes a unified configuration center
The system SHALL expand permission configuration from a narrower permission-maintenance surface into a unified permission configuration center.

#### Scenario: Permission configuration page exposes unified IA
- **WHEN** a user opens permission configuration
- **THEN** the page MUST present a unified information architecture for menu permission, resource permission, and row/column permission workflows
- **AND** the user MUST NOT need a duplicate dedicated menu-permission page to complete those governed tasks

#### Scenario: Existing permission workflows remain semantically reachable after consolidation
- **WHEN** permission-management workflows previously reached through scattered entry points are consolidated
- **THEN** the permission-config capability MUST still expose the governed authorization and data-access workflows required by this change
- **AND** the consolidated page MUST preserve consistent save-and-revisit behavior for the in-scope permission tabs

### Requirement: Unified permission configuration entry
The system SHALL provide a unified permission configuration center instead of scattering permission workflows across multiple management pages.

#### Scenario: Permission center exposes three governed tabs
- **WHEN** a user opens the permission configuration page
- **THEN** the page MUST expose exactly three tabs for menu permission, resource permission, and row/column permission
- **AND** switching between tabs MUST keep the user inside one governed permission-management entry path

### Requirement: Role-based menu authorization remains available from the unified center
The system SHALL let administrators manage menu authorization for roles from the permission configuration center.

#### Scenario: Menu permission tab authorizes governed menu visibility
- **WHEN** a user opens the menu permission tab and saves role-menu authorization changes
- **THEN** the system MUST persist those changes through the governed permission workflow
- **AND** the resulting menu visibility MUST remain consistent with the saved role authorization state

### Requirement: Resource permission supports governed assignment views
The system SHALL provide a unified resource-permission workflow that supports the governed assignment perspectives in this change scope.

#### Scenario: Resource permission tab loads governed resources and assignees
- **WHEN** a user opens the resource permission tab
- **THEN** the page MUST load governed resource families for datasource, dataset, dashboard, and big-screen permissions
- **AND** the workflow MUST support the configured assignment perspectives without forcing the user into a separate permission page

#### Scenario: Resource permission tab loads backfilled historical resources
- **WHEN** a historical datasource, dataset, dashboard, or big-screen resource has been backfilled into the governed resource model
- **THEN** the resource permission tab MUST expose that resource through the same governed resource-view workflow used by newly governed resources
- **AND** the UI MUST NOT require a legacy-only permission path for that resource

#### Scenario: Backfilled historical resource stays consistent across assignment perspectives
- **WHEN** an administrator inspects a backfilled historical resource from the unified permission center
- **THEN** the by-user and by-resource perspectives MUST resolve against the same effective authorization state
- **AND** switching perspectives MUST NOT produce conflicting effective grants for the same governed resource

### Requirement: Backfilled Historical Resources Must Converge to the Governed Authorization Model
The system SHALL bring historical resources that are backfilled into the unified resource model under the same effective authorization semantics used by newly governed resources.

#### Scenario: Historical resource receives inherited effective permission after backfill
- **WHEN** a historical datasource, dataset, dashboard, or big-screen resource is backfilled under a governed parent group
- **THEN** the system MUST calculate and expose inherited effective grants for that resource
- **AND** unified permission queries MUST treat that resource as governed instead of falling back to legacy-only type-scoped semantics

#### Scenario: Historical resource save/query round-trip uses governed state
- **WHEN** an administrator queries or saves permission changes for a backfilled historical resource
- **THEN** the system MUST persist and return authorization state through the same governed permission model used by newly governed resources
- **AND** runtime authorization checks MUST observe the same effective result after refresh

#### Scenario: Historical resource cannot be safely governed automatically
- **WHEN** a historical resource lacks a valid parent scope, governing identity, or organization boundary required by the governed model
- **THEN** the system MUST skip automatic governance for that resource
- **AND** the skip result MUST be auditable and classifiable for follow-up remediation
- **AND** the system MUST NOT pretend the resource is fully governed when it is not

### Requirement: Row and column permission workflows remain first-class within the unified center
The system SHALL expose row and column permission workflows from the same permission center for supported governed authorization targets only.

#### Scenario: Row/column permission tab manages dataset data-access rules
- **WHEN** a user opens the row/column permission tab
- **THEN** the page MUST expose governed dataset-level row-filter and column-control workflows for supported `user` and `role` targets
- **AND** saving those rules MUST keep row-permission and column-permission behavior reachable from the unified permission center

#### Scenario: Deferred system-variable targets do not appear governed
- **WHEN** a permission-center flow encounters deferred system-variable or `sysParams` target semantics
- **THEN** the UI and backend contracts MUST explicitly hide or reject those unsupported targets
- **AND** the system MUST NOT present those targets as completed governed permission-center behavior

#### Scenario: System variable management remains outside row/column permission assignment
- **WHEN** a user needs to manage system variable definitions or selectable values
- **THEN** the system MUST continue to route that work through system variable management capability endpoints and screens
- **AND** the permission center MUST NOT imply that variable management support also provides system-variable permission assignment semantics

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
- **AND** revoked access MUST NOT degrade into a misleading `404` caused by permission-path confusion

### Requirement: Role-Menu Authorization APIs
The system SHALL provide APIs to query and save role-menu authorization state.

#### Scenario: Query role-menu assignments
- **WHEN** a client requests menu assignments for a role
- **THEN** the system returns granted menu IDs and metadata needed for authorization UI

#### Scenario: Save role-menu assignments
- **WHEN** a client submits a complete role-menu assignment set
- **THEN** the system validates role and menu existence before persistence
- **AND** the system returns success only after effective authorization state is stored

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

### Requirement: Permission Dual-Perspective Consistency
The system SHALL provide both "configure by user" and "configure by resource" views, and both views SHALL persist to the same authorization model with equivalent effective results.

#### Scenario: Grant permission in user perspective
- **WHEN** an administrator grants a resource permission in user perspective
- **THEN** the same grant MUST be visible in resource perspective without additional synchronization steps

#### Scenario: Revoke permission in resource perspective
- **WHEN** an administrator revokes a grant in resource perspective
- **THEN** the same revocation MUST be visible in user perspective immediately after data refresh

### Requirement: Resource Group Inheritance Effective on New Resources
The system SHALL apply inherited permissions from resource groups to newly created resources in that group.

#### Scenario: Create resource under granted group
- **WHEN** a new dashboard/dataset is created under a group that already has grants
- **THEN** the new resource MUST inherit effective grants for all authorized users/roles

#### Scenario: Query inherited permission
- **WHEN** a client checks resource permission for an inherited target
- **THEN** the system MUST return permission as granted without requiring manual re-authorization

### Requirement: Menu Authorization Drives Navigation Visibility
The system SHALL bind role-menu authorization outcomes to navigation visibility decisions.

#### Scenario: Grant menu to role
- **WHEN** an administrator grants a menu node to a role
- **THEN** users with that role MUST see the menu in navigation after authorization refresh

#### Scenario: Revoke menu from role
- **WHEN** an administrator revokes a previously granted menu node from a role
- **THEN** users with that role MUST no longer see the menu in navigation
- **AND** direct navigation to revoked route MUST be denied

### Requirement: BI Permission Failure Semantic Stability
The system SHALL keep permission outcomes for core BI flows semantically distinguishable across datasource, dataset, dashboard, and big-screen routes during migration.

#### Scenario: Existing BI resource denied by permission
- **WHEN** an authenticated user accesses an existing BI resource or API without sufficient permission
- **THEN** the system MUST return authorization-denied semantics consistent with the governed permission model
- **AND** MUST NOT convert the result into a generic `404` or placeholder success response

#### Scenario: Missing BI resource remains distinguishable from denied access
- **WHEN** a requested BI resource is absent or not registered
- **THEN** the system MAY return missing-resource semantics
- **AND** operators and clients MUST be able to distinguish that result from permission denial during debugging and support

### Requirement: Permission-Aware Resource Tree Responses
Permission-filtered BI resource trees SHALL remain structurally valid for migrated frontend consumers.

#### Scenario: Filtered tree preserves required operation fields
- **WHEN** authorization filtering removes unauthorized dashboards, screens, datasets, or datasource-linked nodes from a returned tree
- **THEN** the remaining tree MUST preserve required identifiers, hierarchy, and node type fields for supported operations
- **AND** permission filtering MUST NOT produce malformed tree payloads that later fail in copy, move, selection, or preview preparation flows

### Requirement: Permission Center Alignment Must Follow Semantic Sequencing
The governed permission-alignment workflow MUST sequence menu/resource alignment before row/column and deferred P2 expansion.

#### Scenario: Team sequences permission-center work
- **WHEN** maintainers execute permission-center alignment
- **THEN** menu and resource authorization consistency MUST be stabilized before row/column and deferred P2 work is treated as complete
- **AND** deferred items MUST NOT block already-approved P0/P1 semantic corrections

### Requirement: User-View and Resource-View Permission Workflows Must Converge
The system MUST keep by-user and by-resource permission workflows semantically consistent for governed resources.

#### Scenario: Administrator compares two governed permission perspectives
- **WHEN** an administrator inspects the same governed permission state from user-view and resource-view workflows
- **THEN** both workflows MUST resolve to the same effective authorization result
- **AND** target-query or target-save gaps MUST be classified as incomplete rather than treated as equivalent behavior

### Requirement: Row and Column Permissions Must Enforce Governed Runtime Behavior
The system MUST treat row filters, disabled columns, and masked columns as runtime-enforced governed behavior, MUST enforce row-permission gating consistently at the middleware/runtime boundary for governed dataset and chart entry points, MUST keep chart runtime authorization dataset-governed instead of introducing a separate chart resource permission model, and MUST explicitly trace or defer whitelist and system-variable dimensions until they are implemented.

#### Scenario: Governed data-access rule is evaluated at runtime
- **WHEN** a governed dataset request is evaluated against row or column permissions
- **THEN** the system MUST apply the configured enforcement semantics instead of relying on placeholder middleware or UI-only interpretation
- **AND** whitelist and system-variable dimensions MUST be either traceable in the governed permission model or explicitly marked as deferred instead of being silently accepted

#### Scenario: Middleware-enforced governed route establishes row-permission context
- **WHEN** a request enters a governed runtime route that is protected by row-permission middleware
- **THEN** the middleware MUST resolve and validate the runtime row-permission context needed for downstream evaluation before the handler continues
- **AND** the request MUST NOT proceed as a warning-only no-op once middleware enforcement is enabled

#### Scenario: Middleware fails closed when governed context cannot be established
- **WHEN** row-permission middleware cannot safely determine the governed dataset/group context, authenticated user context, or permission lookup prerequisites for a protected route
- **THEN** the request MUST terminate with explicit permission/error semantics
- **AND** the system MUST NOT continue into a permissive service path that could bypass governed enforcement

#### Scenario: Service-layer rule application remains authoritative after middleware rollout
- **WHEN** a governed runtime route passes through row-permission middleware and reaches the dataset/chart service layer
- **THEN** the service layer MUST remain the source of truth for select-column shaping, WHERE-clause construction, disabled-column filtering, and masking behavior
- **AND** middleware MUST NOT duplicate SQL rule compilation logic that already belongs to the row/column permission services

#### Scenario: Middleware rollout does not implicitly govern non-governed routes
- **WHEN** maintainers enable real row-permission middleware enforcement
- **THEN** only explicitly governed runtime routes MUST adopt that middleware behavior in the rollout scope
- **AND** routes outside the governed permission-aware runtime surface MUST NOT gain row-permission enforcement implicitly as an accidental side effect

#### Scenario: Governed chart runtime route resolves dataset-governed permission context
- **WHEN** a governed chart runtime request enters a canonical or compatibility chart data route that only identifies the chart
- **THEN** the system MUST resolve the backing `datasetGroupID` before permission-aware execution continues
- **AND** dataset view permission and downstream row/column enforcement MUST evaluate against that resolved dataset-governed identity instead of a separate chart resource permission model

#### Scenario: Governed chart runtime route fails closed on chart context resolution errors
- **WHEN** a governed chart runtime route cannot resolve the backing dataset group from the provided chart identity, or the resolved dataset-governed context cannot be validated
- **THEN** the request MUST terminate with explicit authorization or error semantics before chart data execution proceeds
- **AND** the system MUST NOT fall back to permissive chart execution because chart context resolution failed

#### Scenario: Compatibility chart permission flow does not synthesize admin identity
- **WHEN** an in-scope compatibility chart permission flow is invoked without authenticated user context
- **THEN** the system MUST return fail-closed unauthorized or permission-denied semantics
- **AND** the flow MUST NOT recover by substituting a default admin user ID or username

#### Scenario: Governed chart field listing stays permission-aware on governed runtime routes
- **WHEN** a governed chart runtime route exposes chart field or dataset-field results through permission-aware chart listing behavior
- **THEN** disabled-column and masking behavior MUST remain consistent with dataset-governed row/column permission semantics
- **AND** the route MUST NOT silently downgrade to non-permission-aware listing behavior solely because it entered through a compatibility path

#### Scenario: Deferred whitelist write attempts fail with stable contract language
- **WHEN** a client submits a non-empty `whiteList` value through the governed row-permission save workflow
- **THEN** the backend MUST reject the request with a stable deferred-semantics error that does not reference internal milestone labels
- **AND** the response MUST make it clear that whitelist persistence/editing is deferred in the permission center rather than partially supported

#### Scenario: Deferred whitelist read contract stays compatibility-safe
- **WHEN** a client reads governed row-permission data from the unified permission-center workflow before whitelist persistence is implemented
- **THEN** any exposed whitelist-related fields MUST remain compatibility-safe and clearly treated as deferred contract surface rather than active persisted state
- **AND** the system MUST NOT populate those fields with misleading non-empty data that implies supported whitelist persistence

#### Scenario: Unified permission center does not offer governed whitelist editing
- **WHEN** an administrator uses the unified permission-center row/column permission workflow
- **THEN** the UI MUST NOT expose a governed whitelist editing path for row permissions
- **AND** any legacy or adjacent flows outside the unified permission center MUST be treated as explicit deferred boundaries rather than evidence that whitelist editing is supported there

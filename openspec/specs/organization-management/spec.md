# organization-management Specification

## Purpose
This capability defines the multi-organization management system for DataEase, supporting hierarchical organization structures, member management, data isolation, organization switching, and statistics. It enables multi-tenant deployment where resources and users are scoped to specific organizations with proper access controls.
## Requirements
### Requirement: Organization Hierarchy

系统 SHALL 在 Go 实现中提供完整的组织层级管理：
- Create organizations with name, description, and parent organization
- Update organization information
- Delete organizations (cascade or prevent if has children)
- View organization tree structure
- Manage organization-specific settings

#### Scenario: Admin creates new organization
- **WHEN** system administrator creates new organization with name and optional parent
- **THEN** Go 实现 SHALL 生成唯一的组织 ID
- **THEN** Go 实现 SHALL 自动计算层级（parent.level + 1）
- **THEN** Go 实现 SHALL 返回 `{msg: "success"}`

#### Scenario: Admin deletes organization with children
- **WHEN** system administrator attempts to delete an organization with children
- **THEN** Go 实现 SHALL 检查是否存在子组织
- **THEN** Go 实现 SHALL 返回错误信息如果存在子组织

#### Scenario: Admin deletes organization without children
- **WHEN** system administrator deletes an organization without children
- **THEN** Go 实现 SHALL 设置 del_flag = 1 实现软删除
- **THEN** Go 实现 SHALL 返回 `{msg: "success"}`

### Requirement: Organization Member Management

系统 SHALL 在 Go 实现中保持与 Java 版本相同的成员管理行为。

#### Scenario: Member assignment
- **WHEN** 分配用户到组织
- **THEN** Go 实现 SHALL 以相同方式更新组织成员关系表

### Requirement: Organization Data Isolation
The system SHALL ensure data isolation between different organizations including:
- Resources (datasets, dashboards, screens) belong to specific organizations
- Users can only access resources within their assigned organizations
- Permissions are scoped to organization level
- Search and filters are organization-aware

#### Scenario: User from Org A cannot access Org B resources
- **WHEN** user from Organization A attempts to access dashboard from Organization B
- **THEN** system checks resource's organization
- **THEN** system denies access if organizations don't match
- **THEN** system displays "Permission denied" error

### Requirement: Organization Statistics
The system SHALL provide organization-level statistics including:
- Number of users in organization
- Number of resources (datasets, dashboards) in organization
- Storage usage per organization
- Recent activity summary

#### Scenario: Admin views organization statistics
- **WHEN** administrator opens organization details
- **THEN** system displays real-time statistics
- **THEN** system shows user count, resource count
- **THEN** system shows storage usage if applicable

### Requirement: Organization Switching
The system SHALL allow users with multiple organization access to switch between organizations including:
- Display current organization in UI
- Show available organizations in dropdown
- Switch organization without re-login
- Apply organization-specific permissions after switch

#### Scenario: User switches to another organization
- **WHEN** user selects different organization from dropdown
- **THEN** system updates current organization context
- **THEN** system reloads permissions for new organization
- **THEN** system displays resources from new organization
- **THEN** UI reflects organization-specific branding if configured

### Requirement: Organization Tree Structure

系统 SHALL 在 Go 实现中保持与 Java 版本相同的组织树形结构。

#### Scenario: Tree query compatibility
- **WHEN** 查询组织树
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的树形结构数据格式

#### Scenario: Recursive CTE query
- **WHEN** 查询子组织
- **THEN** Go 实现 SHALL 使用与 Java 版本等效的递归 CTE SQL

### Requirement: Organization Repository Pattern

系统 SHALL 使用 Repository 模式实现数据访问层。

#### Scenario: Organization repository interface
- **WHEN** 实现组织数据访问
- **THEN** 系统 SHALL 提供 OrgRepository 接口，包含 Create、Update、Delete、GetByID、List、ListByParentID 方法

### Requirement: Organization Tree Query

系统 SHALL 支持组织树形结构查询。

#### Scenario: Get organization tree
- **WHEN** 请求组织树结构
- **THEN** Go 实现 SHALL 返回树形结构数据，包含 children 字段

#### Scenario: Get children by parent ID
- **WHEN** 根据父组织 ID 查询子组织
- **THEN** Go 实现 SHALL 返回所有直接子组织列表

### Requirement: Organization Service Layer

系统 SHALL 使用 Service 层封装业务逻辑。

#### Scenario: Organization service interface
- **WHEN** 实现组织业务逻辑
- **THEN** 系统 SHALL 提供 OrgService 接口，包含 CreateOrg、UpdateOrg、DeleteOrg、GetOrgTree、UpdateOrgStatus 方法

#### Scenario: Level calculation
- **WHEN** 创建子组织
- **THEN** Service 层 SHALL 自动计算层级为 parent.level + 1

### Requirement: Organization Handler REST Mapping

系统 SHALL 提供 HTTP Handler 映射 REST API 到 Service 层。

#### Scenario: POST /api/system/organization/create
- **WHEN** 客户端请求 POST /api/system/organization/create
- **THEN** Handler SHALL 解析请求体（orgName, orgDesc, parentId）
- **THEN** Handler SHALL 调用 Service 层创建组织
- **THEN** Handler SHALL 返回 `{msg: "success"}` 或 `{msg: "Failed: <error>"}`

#### Scenario: GET /api/system/organization/tree
- **WHEN** 客户端请求 GET /api/system/organization/tree
- **THEN** Handler SHALL 返回树形结构数据

#### Scenario: GET /api/system/organization/checkName
- **WHEN** 客户端请求 GET /api/system/organization/checkName?orgName=xxx
- **THEN** Handler SHALL 返回 `{exists: true/false, msg: "success"}`


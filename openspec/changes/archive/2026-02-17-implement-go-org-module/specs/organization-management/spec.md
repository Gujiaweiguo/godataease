# organization-management Specification Delta

## MODIFIED Requirements

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

## ADDED Requirements

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

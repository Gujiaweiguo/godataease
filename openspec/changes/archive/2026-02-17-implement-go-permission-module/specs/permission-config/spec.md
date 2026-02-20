# permission-config Specification Delta

## ADDED Requirements

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

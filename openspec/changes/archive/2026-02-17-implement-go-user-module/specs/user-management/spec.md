# user-management Specification Delta

## MODIFIED Requirements

### Requirement: User CRUD Operations

系统 SHALL 在 Go 实现中提供完整的用户 CRUD 操作：
- Create new users with username, password, email, phone, and organization assignment
- Read user list with filtering (by organization, role, status) and pagination
- Update user information (profile, status, password, organization, roles)
- Delete users (soft delete with del_flag)
- Reset user passwords
- Enable/disable user accounts

#### Scenario: Admin creates new user
- **WHEN** system administrator submits user creation request with username, password, email
- **THEN** Go 实现 SHALL 使用 bcrypt 算法加密密码
- **THEN** Go 实现 SHALL 生成与 Java 版本兼容的用户数据
- **THEN** Go 实现 SHALL 返回 `{code: "000000", data: userId, msg: "success"}`

#### Scenario: Admin deletes user
- **WHEN** system administrator deletes a user
- **THEN** Go 实现 SHALL 设置 del_flag = 1 实现软删除
- **THEN** Go 实现 SHALL 返回 `{code: "000000", msg: "success"}`

#### Scenario: Password encryption compatibility
- **WHEN** 设置或更新用户密码
- **THEN** Go 实现 SHALL 使用与 Java 版本相同的 bcrypt 算法和加密强度（cost=10）

## ADDED Requirements

### Requirement: User Repository Pattern

系统 SHALL 使用 Repository 模式实现数据访问层。

#### Scenario: User repository interface
- **WHEN** 实现用户数据访问
- **THEN** 系统 SHALL 提供 UserRepository 接口，包含 Create、Update、Delete、GetByID、GetByUsername、Query 方法

#### Scenario: GORM entity mapping
- **WHEN** 定义用户实体
- **THEN** Go 实现 SHALL 使用 GORM 标签映射数据库字段（`gorm:"column:user_id"`）

### Requirement: User Service Layer

系统 SHALL 使用 Service 层封装业务逻辑。

#### Scenario: User service interface
- **WHEN** 实现用户业务逻辑
- **THEN** 系统 SHALL 提供 UserService 接口，包含 CreateUser、UpdateUser、DeleteUser、SearchUsers、ResetPassword 方法

#### Scenario: Password validation
- **WHEN** 用户更新密码
- **THEN** Service 层 SHALL 验证密码强度（长度、复杂度）

### Requirement: User Handler REST Mapping

系统 SHALL 提供 HTTP Handler 映射 REST API 到 Service 层。

#### Scenario: POST /api/system/user/list
- **WHEN** 客户端请求 POST /api/system/user/list
- **THEN** Handler SHALL 解析请求参数（current, size, keyword, orgId, status）
- **THEN** Handler SHALL 调用 Service 层查询
- **THEN** Handler SHALL 返回分页结果 `{code, data: {list, total, current, size}, msg}`

#### Scenario: Error response format
- **WHEN** 用户操作失败
- **THEN** Handler SHALL 返回 `{code: "500000", msg: "Failed: <error message>"}`

### Requirement: User Audit Integration

系统 SHALL 在用户操作时记录审计日志。

#### Scenario: User creation audit
- **WHEN** 用户被创建
- **THEN** 系统 SHALL 调用 AuditService 记录 USER_ACTION 类型审计日志

#### Scenario: User deletion audit
- **WHEN** 用户被删除
- **THEN** 系统 SHALL 调用 AuditService 记录 DELETE 操作审计日志

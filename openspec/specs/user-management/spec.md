# user-management Specification

## Purpose
This capability provides complete user lifecycle management including CRUD operations, profile management, search and filtering, and bulk operations. It integrates with organization and role systems to support multi-user environments with efficient user administration workflows.
## Requirements
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

### Requirement: User Profile Management
The system SHALL allow users to manage their own profiles including:
- View and edit personal information (name, email, phone)
- Change password (with old password verification)
- Upload and update avatar image
- Manage personal API keys

#### Scenario: User changes password
- **WHEN** user navigates to profile settings and enters old password and new password
- **THEN** system validates old password
- **THEN** system updates password hash in database
- **THEN** system forces re-authentication on next request

### Requirement: User Search and Filtering
The system SHALL provide search and filtering capabilities for user management including:
- Search by username, email, phone
- Filter by organization
- Filter by role
- Filter by account status (enabled/disabled)
- Sort by various fields (create time, last login time)

#### Scenario: Admin searches for specific user
- **WHEN** admin enters search term in user management page
- **THEN** system filters user list in real-time
- **THEN** results display matching users with pagination

### Requirement: User Bulk Operations
The system SHALL support bulk operations for efficient user management including:
- Batch import users from CSV/Excel
- Batch enable/disable multiple users
- Batch assign roles to multiple users
- Batch delete multiple users

#### Scenario: Admin imports users from CSV
- **WHEN** admin uploads CSV file with user data
- **THEN** system validates CSV format and required fields
- **THEN** system creates users in batch
- **THEN** system displays success/failure summary
- **THEN** system sends welcome emails to imported users

### Requirement: User API Compatibility

系统 SHALL 保持用户管理 REST API 的完全兼容。

#### Scenario: API response format
- **WHEN** 调用用户管理 API
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的 JSON 响应格式和字段

#### Scenario: Error response format
- **WHEN** 用户操作失败
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的错误码和错误消息格式

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


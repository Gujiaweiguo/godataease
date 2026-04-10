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

系统 SHALL 保持用户管理 REST API 的完全兼容，并在用户资料引导响应中返回规范化后的语言值供前端建立 locale 状态。

#### Scenario: API response format
- **WHEN** 调用用户管理 API
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的 JSON 响应格式和字段

#### Scenario: Error response format
- **WHEN** 用户操作失败
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的错误码和错误消息格式

#### Scenario: Current user language bootstrap
- **WHEN** 客户端调用当前用户资料接口（例如 `GET /user/info`）以初始化前端状态
- **THEN** 响应中的 `language` 字段 SHALL 返回受支持的规范化 locale 值
- **AND** Go 实现 SHALL NOT 为所有用户返回固定的硬编码默认语言

#### Scenario: Legacy `/user/org/option` endpoint returns user-option payload semantics
- **WHEN** a client invokes compatibility endpoint `/user/org/option`
- **THEN** the response MUST preserve user-option payload semantics expected by legacy consumers
- **AND** the endpoint MUST remain aligned with user-option behavior instead of returning organization-list payload semantics

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

### Requirement: Organization-Scoped User Membership Baseline
The system SHALL define user lifecycle operations against an explicit organization-scoped membership baseline that can be reused by later role and permission changes.

#### Scenario: Administrator creates user within organization scope
- **WHEN** an administrator creates a user for a target organization
- **THEN** the system MUST persist the user's organization-scoped membership baseline required by downstream role assignment
- **AND** later role workflows MUST be able to discover that user through the same organization scope

#### Scenario: Administrator queries users for organization administration
- **WHEN** an administrator opens a user list under a given organization context
- **THEN** the system MUST return users according to that organization scope
- **AND** the result MUST remain consistent with the organization context established by foundation bootstrap

### Requirement: User management scope expands to include role workflows
The system SHALL expand the user-management capability so that governed role workflows are hosted inside the user-management page.

#### Scenario: Default user-management experience preserves user workflows
- **WHEN** a user opens the user-management page
- **THEN** existing in-scope user CRUD and maintenance workflows MUST remain available from the User tab
- **AND** the expanded page structure MUST NOT remove the governed user-management functions already available before the refactor

#### Scenario: User-management page provides in-context access to role workflows
- **WHEN** a user needs to manage roles from the user-management area
- **THEN** the user-management capability MUST provide a Role tab within the same governed page
- **AND** the role workflows exposed there MUST remain part of the user-management experience for this change scope

### Requirement: User Management Must Expose Role-Workflow Entry Consistent With Organization Context
The system SHALL expose role-workflow entry from user management using the same organization context that governs user administration.

#### Scenario: Administrator transitions from user list to role tab
- **WHEN** an administrator opens role workflows from the user-management surface
- **THEN** the system MUST carry forward the active organization context already established for user administration
- **AND** role member discovery MUST remain consistent with that carried context

### Requirement: Excel User Import with Partial Success
The system SHALL provide Excel-based user bulk import with template validation, partial-success processing, and downloadable error reports.

#### Scenario: Import file contains valid and invalid rows
- **WHEN** an administrator uploads a valid template file containing both compliant and non-compliant rows
- **THEN** the system MUST import all compliant rows
- **AND** the system MUST reject non-compliant rows without rolling back compliant rows
- **AND** the system MUST return an error report download reference

#### Scenario: Import file exceeds size limit
- **WHEN** an administrator uploads a file larger than 10 MB
- **THEN** the system MUST reject the upload with a validation error
- **AND** no user record MUST be created

### Requirement: User Password Reset Flow
The system SHALL allow authorized administrators to reset a user's password to the configured initial password policy.

#### Scenario: Reset password for active user
- **WHEN** an administrator triggers password reset on an enabled user
- **THEN** the system MUST update the user's password hash
- **AND** the system MUST return a success response compatible with existing frontend handling

#### Scenario: Unauthorized user attempts reset
- **WHEN** a non-authorized operator calls password reset endpoint
- **THEN** the system MUST deny the request
- **AND** the system MUST not mutate credential data

#### Scenario: Canonical user-management password endpoints are primary
- **WHEN** frontend user-management flows request default password or trigger reset password
- **THEN** the system MUST provide canonical endpoints `/system/user/defaultPwd` and `/system/user/resetPwd/:id`
- **AND** canonical endpoint behavior MUST remain compatible with existing reset-password response semantics

### Requirement: User Lifecycle Must Be Organization-Scoped by Default
The system MUST treat organization selection as a first-class prerequisite for governed user administration.

#### Scenario: Administrator creates or edits a governed user
- **WHEN** an administrator creates or edits a user in the governed workflow
- **THEN** the workflow MUST bind the operation to an explicit organization context
- **AND** account identity fields that are defined as immutable by baseline policy MUST remain immutable after creation

### Requirement: User Enabled State Must Govern Login Eligibility
The system MUST align user enabled-state semantics with the frozen official baseline.

#### Scenario: Disabled user attempts authentication
- **WHEN** a disabled user attempts to authenticate or reuse a governed session
- **THEN** the system MUST deny access according to the governed error semantics
- **AND** user-management and authentication flows MUST observe the same enabled-state contract

### Requirement: User Import and Source Metadata Must Have Explicit Governance Boundaries
The system MUST treat import, error-report output, and third-party source metadata as governed user-management concerns with explicit rollout boundaries.

#### Scenario: Team evaluates user import parity
- **WHEN** maintainers align user import against the frozen official baseline
- **THEN** the change MUST preserve partial-success and error-report behavior
- **AND** any unsupported third-party source metadata MUST be recorded as deferred or intentionally bounded instead of being silently ignored

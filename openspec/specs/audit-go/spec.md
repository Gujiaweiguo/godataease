# audit-go Specification

## Purpose
TBD - created by archiving change implement-go-audit-module. Update Purpose after archive.
## Requirements
### Requirement: Audit Log Entity

系统 SHALL 使用 Go 结构体定义审计日志实体，字段与 Java 版本兼容。

#### Scenario: AuditLog entity definition
- **WHEN** 定义审计日志实体
- **THEN** 系统 SHALL 包含 id、user_id、username、action_type、action_name、resource_type、resource_id、resource_name、operation、status、failure_reason、ip_address、user_agent、before_value、after_value、organization_id、create_time 字段

#### Scenario: LoginFailure entity definition
- **WHEN** 定义登录失败实体
- **THEN** 系统 SHALL 包含 id、username、ip_address、failure_reason、user_agent、create_time 字段

### Requirement: Audit Log Repository

系统 SHALL 使用 GORM 实现审计日志数据访问层。

#### Scenario: Create audit log
- **WHEN** 创建审计日志
- **THEN** 系统 SHALL 将记录插入 de_audit_log 表并返回生成的 ID

#### Scenario: Paginated query
- **WHEN** 分页查询审计日志
- **THEN** 系统 SHALL 支持按 user_id、action_type、resource_type、organization_id、create_time 范围过滤

### Requirement: Audit Log Service

系统 SHALL 实现审计日志业务逻辑层。

#### Scenario: Create with defaults
- **WHEN** 创建审计日志
- **THEN** 系统 SHALL 自动设置 create_time 和默认 status

#### Scenario: Record login failure
- **WHEN** 记录登录失败
- **THEN** 系统 SHALL 将记录插入 de_login_failure 表

#### Scenario: Retention cleanup
- **WHEN** 清理过期日志
- **THEN** 系统 SHALL 删除指定天数前的记录

### Requirement: Audit Log API

系统 SHALL 提供与 Java 版本兼容的 REST API。

#### Scenario: API path compatibility
- **WHEN** 客户端调用审计 API
- **THEN** 系统 SHALL 使用与 Java 相同的路径前缀 /api/audit

#### Scenario: Response format compatibility
- **WHEN** 返回 API 响应
- **THEN** 系统 SHALL 使用 code/msg/data 格式与 Java 保持一致

### Requirement: Audit Middleware

系统 SHALL 提供审计中间件用于自动记录请求。

#### Scenario: Automatic request logging
- **WHEN** 请求经过审计中间件
- **THEN** 系统 SHALL 自动提取并记录 user_id、ip_address、user_agent 等信息


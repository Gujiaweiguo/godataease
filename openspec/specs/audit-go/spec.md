# audit-go Specification

## Purpose
Define Go audit logging requirements for authentication events and operation traceability.
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

系统 SHALL 实现审计日志业务逻辑层，并提供由持久化审计设置驱动的保留清理与告警集成能力。

#### Scenario: Create with defaults
- **WHEN** 创建审计日志
- **THEN** 系统 SHALL 自动设置 create_time 和默认 status

#### Scenario: Record login failure
- **WHEN** 记录登录失败
- **THEN** 系统 SHALL 将记录插入 de_login_failure 表

#### Scenario: Retention cleanup
- **WHEN** 清理过期日志
- **THEN** 系统 SHALL 删除指定天数前的记录

#### Scenario: Scheduled cleanup uses persisted settings
- **WHEN** 调度任务执行审计日志清理
- **THEN** 系统 SHALL 读取持久化的 retentionDays 和 cleanupFrequency 设置并按配置执行清理

#### Scenario: Alert detection orchestrates governed rules
- **WHEN** 审计服务执行告警检测流程
- **THEN** 系统 SHALL 基于持久化告警设置检查失败登录、权限变更和批量操作规则

### Requirement: Audit Log API

系统 SHALL 提供与 Java 版本兼容的 REST API，并扩展审计设置、即时清理和测试通知接口。

#### Scenario: API path compatibility
- **WHEN** 客户端调用审计 API
- **THEN** 系统 SHALL 使用与 Java 相同的路径前缀 /api/audit

#### Scenario: Response format compatibility
- **WHEN** 返回 API 响应
- **THEN** 系统 SHALL 使用 code/msg/data 格式与 Java 保持一致

#### Scenario: Audit settings endpoints are available
- **WHEN** 客户端读取或保存审计设置
- **THEN** 系统 SHALL 提供审计设置查询与保存接口以覆盖保留、告警、通知和导出配置

#### Scenario: Cleanup-now and test-notification endpoints are available
- **WHEN** 客户端请求立即清理或发送测试通知
- **THEN** 系统 SHALL 提供对应的审计操作接口并返回显式执行结果

### Requirement: Audit Middleware

系统 SHALL 提供审计中间件用于自动记录请求。

#### Scenario: Automatic request logging
- **WHEN** 请求经过审计中间件
- **THEN** 系统 SHALL 自动提取并记录 user_id、ip_address、user_agent 等信息

### Requirement: UserService audit dependency is wired at startup
The UserService MUST have its audit service dependency injected during application initialization.

#### Scenario: UserService audit dependency is wired at startup
- **WHEN** the application starts and initializes the UserService
- **THEN** the UserService MUST have its audit service dependency injected via SetAuditService
- **AND** password reset operations MUST produce audit log entries in de_audit_log

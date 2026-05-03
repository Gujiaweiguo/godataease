## MODIFIED Requirements

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

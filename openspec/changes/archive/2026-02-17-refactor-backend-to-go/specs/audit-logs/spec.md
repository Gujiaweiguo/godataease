# audit-logs Specification Delta

## MODIFIED Requirements

### Requirement: Comprehensive Audit Logging

系统 SHALL 在 Go 实现中保持与 Java 版本相同的审计日志记录行为。

#### Scenario: User Operation Audit
- **WHEN** 管理员创建、更新或删除用户
- **THEN** 系统 SHALL 自动记录操作类型（USER_ACTION）、操作名称、资源类型（USER）、操作时间、操作者、IP 地址和 User-Agent，格式与 Java 版本一致

#### Scenario: Audit Log Storage
- **WHEN** 审计日志生成
- **THEN** 系统 SHALL 将日志存储在相同的数据库表中，字段格式与 Java 版本兼容

## ADDED Requirements

### Requirement: Audit Log Query Performance

系统 SHALL 在 Go 实现中保持或提升审计日志查询性能。

#### Scenario: Paginated query latency
- **WHEN** 查询审计日志列表
- **THEN** P95 延迟 SHALL 小于或等于 Java 实现

#### Scenario: Filter query performance
- **WHEN** 使用复杂过滤条件查询审计日志
- **THEN** 查询性能 SHALL 不低于 Java 实现

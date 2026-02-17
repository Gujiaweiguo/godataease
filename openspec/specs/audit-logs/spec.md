# Audit Logs Capability

## Purpose

提供完整的审计日志系统，用于记录和追踪系统中的所有用户操作、权限变更、数据访问和系统操作，满足安全合规性和运营监控需求。
## Requirements

待添加...

### Requirement: Comprehensive Audit Logging

系统 SHALL 提供审计日志功能，记录所有关键系统操作，包括用户管理、组织管理、权限管理、嵌入功能、登录活动等。

#### Scenario: User Operation Audit
- **WHEN** 管理员创建、更新或删除用户
- **THEN** 系统 SHALL 自动记录操作类型（USER_ACTION）、操作名称、资源类型（USER）、操作时间、操作者、IP 地址和 User-Agent

#### Scenario: Organization Operation Audit
- **WHEN** 管理员创建、更新或删除组织，或修改组织状态
- **THEN** 系统 SHALL 记录权限变更类型（PERMISSION_CHANGE）、操作类型、组织 ID、操作详情和操作者

#### Scenario: Permission Operation Audit
- **WHEN** 管理员创建、更新或删除权限
- **THEN** 系统 SHALL 记录权限变更类型（PERMISSION_CHANGE）、操作类型、权限 ID、操作详情和操作者

#### Scenario: Failed Login Attempts
- **WHEN** 用户使用错误的凭证登录系统
- **THEN** 系统 SHALL 记录失败登录日志，包含用户名、IP 地址、User-Agent 和失败原因

#### Scenario: Data Export Audit
- **WHEN** 用户导出数据（数据集、仪表板、视图等）
- **THEN** 系统 SHALL 记录操作类型（EXPORT）、数据来源、导出格式、记录数和操作者

#### Scenario: Embedded BI Access Audit
- **WHEN** 外部系统通过 iframe 访问嵌入的 BI 内容
- **THEN** 系统 SHALL 记录数据访问类型（DATA_ACCESS）、应用 ID、origin 来源和访问时间

#### Scenario: Audit Log Query and Filtering
- **WHEN** 管理员查询审计日志
- **THEN** 系统 SHALL 支持按用户 ID、用户名、操作类型、资源类型、组织 ID、日期范围进行筛选，并支持分页

#### Scenario: Audit Log Export
- **WHEN** 管理员需要导出审计日志
- **THEN** 系统 SHALL 支持按日志 ID 列表导出为 CSV 或 JSON 格式

#### Scenario: Audit Log Retention
- **WHEN** 系统需要清理旧审计日志
- **THEN** 系统 SHALL 提供自动清理接口，默认保留 90 天的审计日志

### Requirement: Automatic Audit Logging via Annotations

系统 SHALL 使用 `@AuditLog` 注解和 AOP 切面自动记录方法调用。

#### Scenario: Service Method with AuditLog Annotation
- **WHEN** Service 方法标注了 `@AuditLog(actionType, actionName, resourceType)`
- **THEN** AOP 切面 SHALL 自动记录审计日志，包含方法参数、返回值、异常信息和请求上下文

#### Scenario: Automatic Operation Type Detection
- **WHEN** 方法名包含 "create"、"update"、"delete" 等关键词
- **THEN** 系统 SHALL 自动推断操作类型（CREATE、UPDATE、DELETE 等）

#### Scenario: Automatic Action Type Detection
- **WHEN** 类名包含 "User"、"Org"、"Role"、"Permission"、"Dataset"、"Dashboard" 等
- **THEN** 系统 SHALL 自动推断操作类型（USER_ACTION、PERMISSION_CHANGE、DATA_ACCESS、SYSTEM_CONFIG）

### Requirement: Immutable Audit Logs

创建的审计日志 SHALL 是不可变的，只能读取，不能修改或删除（除了系统自动清理）。

#### Scenario: Audit Log Immutability
- **WHEN** 审计日志已创建
- **THEN** 日志 SHALL 不能被修改，只能通过查询接口读取

### Requirement: Security Compliance

审计日志系统 SHALL 支持安全合规要求，如 SOC 2、GDPR、ISO 27001 等。

#### Scenario: User Activity Tracking
- **WHEN** 需要进行安全审计或调查
- **THEN** 系统 SHALL 提供完整的用户活动轨迹，包含操作时间、操作内容、IP 地址、设备信息

#### Scenario: Data Breach Forensic Analysis
- **WHEN** 发生数据泄露事件
- **THEN** 系统 SHALL 可以查询审计日志，追踪谁在何时访问或修改了哪些数据

#### Scenario: Accountability and Non-Repudiation
- **WHEN** 需要确定某个操作的执行者
- **THEN** 审计日志 SHALL 明确记录操作者 ID、用户名和组织 ID

### Requirement: Real-time Monitoring and Alerts

系统 SHALL 提供实时监控和告警功能，用于检测可疑活动。

#### Scenario: Suspicious Activity Detection
- **WHEN** 检测到异常登录模式（如短时间内多次失败登录）
- **THEN** 系统 SHALL 记录并可选地发出告警

#### Scenario: Unauthorized Access Attempts
- **WHEN** 检测到未授权的访问尝试
- **THEN** 系统 SHALL 记录详细信息并保留用于安全分析


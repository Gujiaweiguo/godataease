# Audit Logs Capability

## Purpose

提供完整的审计日志系统，用于记录和追踪系统中的所有用户操作、权限变更、数据访问和系统操作，满足安全合规性和运营监控需求。
## Requirements

待添加...

### Requirement: Comprehensive Audit Logging

系统 SHALL 在 Go 实现中保持与 Java 版本相同的审计日志记录行为。

#### Scenario: User Operation Audit
- **WHEN** 管理员创建、更新或删除用户
- **THEN** 系统 SHALL 自动记录操作类型（USER_ACTION）、操作名称、资源类型（USER）、操作时间、操作者、IP 地址和 User-Agent，格式与 Java 版本一致

#### Scenario: Audit Log Storage
- **WHEN** 审计日志生成
- **THEN** 系统 SHALL 将日志存储在相同的数据库表中，字段格式与 Java 版本兼容

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

### Requirement: Audit Log Query Performance

系统 SHALL 在 Go 实现中保持或提升审计日志查询性能。

#### Scenario: Paginated query latency
- **WHEN** 查询审计日志列表
- **THEN** P95 延迟 SHALL 小于或等于 Java 实现

#### Scenario: Filter query performance
- **WHEN** 使用复杂过滤条件查询审计日志
- **THEN** 查询性能 SHALL 不低于 Java 实现

### Requirement: Audit Page Reachability and Query Recovery
The system SHALL treat audit page reachability, filter queries, and detail-read flows as a governed broken-feature recovery surface.

#### Scenario: Audit page initializes for governed query workflows
- **WHEN** a user enters the audit page through an in-scope route or menu path
- **THEN** the page MUST initialize its required query state and return explicit results or explicit failure
- **AND** initialization failure MUST NOT collapse into a misleading blank-success state

#### Scenario: Audit recovery preserves diagnosable query outcomes
- **WHEN** audit list, filter, or detail-read flows fail during stabilization
- **THEN** the recovery result MUST preserve diagnosable semantics for authorization, route, and business-query failure
- **AND** the recovered path MUST be covered by targeted verification

### Requirement: Audit Route-Level Hardening
The system SHALL preserve explicit route-level audit semantics after the operational recovery batch is complete.

#### Scenario: Audit page entry and detail routes remain explicit beyond unit coverage
- **WHEN** a user enters the audit page or reads audit detail through a governed route path
- **THEN** the route-level behavior MUST remain explicit for authorization, not-found, and query failure outcomes
- **AND** route-level or smoke verification MUST exist in addition to unit/handler coverage

### Requirement: Audit Export Flows Must Enforce Request Limits
The system SHALL enforce request-throttling on audit export and export-download routes so authenticated users cannot repeatedly generate or fetch audit exports at unbounded frequency.

#### Scenario: Audit export remains available within request budget
- **WHEN** an authenticated client calls `POST /api/audit/export` within the configured audit-export request budget
- **THEN** the system MUST preserve the existing audit export behavior and response envelope
- **AND** the request MUST still return export metadata when business validation succeeds

#### Scenario: Audit export download remains available within request budget
- **WHEN** an authenticated client calls `GET /api/audit/download` within the configured audit-export request budget
- **THEN** the system MUST preserve the existing audit download validation and file-delivery behavior

#### Scenario: Audit export or download exceeds request budget
- **WHEN** an authenticated client exceeds the configured request budget for audit export or download requests within the active rate-limit window
- **THEN** the system MUST reject the request with an explicit throttling response
- **AND** the throttled request MUST NOT continue into export generation or file delivery

### Requirement: Audit Export Flows Must Enforce Authorization Boundaries
The system SHALL restrict audit export and export-download routes to authenticated users who hold the governed audit authorization for the audit feature surface.

#### Scenario: Authorized user exports audit logs
- **WHEN** an authenticated user with audit feature authorization calls `POST /api/audit/export`
- **THEN** the system MUST preserve the existing audit export behavior and response envelope
- **AND** the request MUST remain subject to the existing audit export rate limits

#### Scenario: Authorized user downloads audit export
- **WHEN** an authenticated user with audit feature authorization calls `GET /api/audit/download`
- **THEN** the system MUST preserve the existing audit download validation and file-delivery behavior
- **AND** the request MUST remain subject to the existing audit export/download rate limits

#### Scenario: Authenticated user without audit authorization is denied export access
- **WHEN** an authenticated user without the governed audit authorization calls `POST /api/audit/export` or `GET /api/audit/download`
- **THEN** the system MUST reject the request with HTTP `403`
- **AND** the request MUST NOT continue into export generation or file delivery

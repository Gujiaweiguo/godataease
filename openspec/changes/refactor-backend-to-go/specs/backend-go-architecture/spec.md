# backend-go-architecture Specification Delta

## ADDED Requirements

### Requirement: Go Backend Framework

系统 SHALL 使用 Go 语言实现后端服务，基于 Gin 框架提供 HTTP API。

#### Scenario: HTTP Server Startup
- **WHEN** 启动后端服务
- **THEN** 系统 SHALL 在配置的端口启动 Gin HTTP 服务器，并注册所有 API 路由

#### Scenario: Request Routing
- **WHEN** 客户端发送 HTTP 请求到已注册的路由
- **THEN** 系统 SHALL 根据路由配置将请求分发到对应的 Handler

### Requirement: GORM Data Access

系统 SHALL 使用 GORM 作为 ORM 框架进行数据库访问。

#### Scenario: Database Connection
- **WHEN** 服务启动时
- **THEN** 系统 SHALL 使用配置的连接信息建立 MySQL 数据库连接池

#### Scenario: CRUD Operations
- **WHEN** 业务逻辑需要访问数据库
- **THEN** 系统 SHALL 通过 GORM Repository 进行 CRUD 操作

#### Scenario: Complex Queries
- **WHEN** 业务逻辑需要执行复杂 SQL 查询
- **THEN** 系统 SHALL 使用原生 SQL 以保证性能和可控性

### Requirement: WebSocket Real-time Communication

系统 SHALL 使用 gorilla/websocket 实现 WebSocket 实时通信。

#### Scenario: WebSocket Connection
- **WHEN** 客户端发起 WebSocket 连接请求
- **THEN** 系统 SHALL 通过 WebSocket Handler 建立连接并注册到 Hub

#### Scenario: Message Broadcasting
- **WHEN** 服务端需要推送消息
- **THEN** 系统 SHALL 通过 Hub 将消息广播给所有或指定的客户端连接

#### Scenario: Heartbeat Keep-alive
- **WHEN** WebSocket 连接建立后
- **THEN** 系统 SHALL 定期发送心跳包以保持连接活跃

### Requirement: JWT Authentication

系统 SHALL 使用 JWT 进行用户认证。

#### Scenario: Token Generation
- **WHEN** 用户成功登录
- **THEN** 系统 SHALL 生成包含用户信息的 JWT Token 并返回给客户端

#### Scenario: Token Validation
- **WHEN** 客户端携带 Token 访问受保护资源
- **THEN** 系统 SHALL 验证 Token 有效性和过期时间

#### Scenario: Token Refresh
- **WHEN** Token 即将过期
- **THEN** 系统 SHALL 支持通过刷新 Token 获取新的访问 Token

### Requirement: Redis Caching

系统 SHALL 使用 Redis 作为缓存和会话存储。

#### Scenario: Cache Read
- **WHEN** 业务逻辑需要读取缓存数据
- **THEN** 系统 SHALL 首先查询 Redis，若命中则直接返回

#### Scenario: Cache Write
- **WHEN** 业务逻辑需要缓存数据
- **THEN** 系统 SHALL 将数据写入 Redis 并设置过期时间

#### Scenario: Session Storage
- **WHEN** 用户登录成功
- **THEN** 系统 SHALL 将会话信息存储在 Redis 中

### Requirement: Structured Logging

系统 SHALL 使用 zap 进行结构化日志记录。

#### Scenario: Request Logging
- **WHEN** 收到 HTTP 请求
- **THEN** 系统 SHALL 记录请求路径、方法、状态码、耗时等信息

#### Scenario: Error Logging
- **WHEN** 发生错误
- **THEN** 系统 SHALL 记录错误详情、堆栈信息和上下文

#### Scenario: Log Format
- **WHEN** 输出日志
- **THEN** 系统 SHALL 支持 JSON 和 Console 两种格式

### Requirement: OpenTelemetry Observability

系统 SHALL 使用 OpenTelemetry 实现可观测性。

#### Scenario: Distributed Tracing
- **WHEN** 处理请求
- **THEN** 系统 SHALL 生成 Trace ID 并传递给下游服务

#### Scenario: Metrics Collection
- **WHEN** 服务运行
- **THEN** 系统 SHALL 收集 Prometheus 格式的指标数据

#### Scenario: Trace Export
- **WHEN** 请求完成
- **THEN** 系统 SHALL 将追踪数据导出到配置的后端（Jaeger/Zipkin 等）

### Requirement: Scheduled Tasks

系统 SHALL 使用 robfig/cron 实现定时任务。

#### Scenario: Task Scheduling
- **WHEN** 服务启动
- **THEN** 系统 SHALL 根据配置注册所有定时任务

#### Scenario: Cron Expression
- **WHEN** 定义定时任务
- **THEN** 系统 SHALL 支持标准 cron 表达式语法

#### Scenario: Distributed Lock
- **WHEN** 定时任务执行
- **THEN** 系统 SHALL 通过 Redis 分布式锁保证单节点执行

### Requirement: JVM Service Integration

系统 SHALL 通过 gRPC 调用保留的 JVM 服务（Calcite、SeaTunnel）。

#### Scenario: Calcite SQL Parsing
- **WHEN** 需要解析或验证 SQL
- **THEN** 系统 SHALL 通过 gRPC 调用 Calcite 服务

#### Scenario: SeaTunnel Data Sync
- **WHEN** 需要执行数据同步任务
- **THEN** 系统 SHALL 通过 gRPC 调用 SeaTunnel 服务

#### Scenario: Connection Pooling
- **WHEN** 调用 JVM 服务
- **THEN** 系统 SHALL 使用 gRPC 连接池优化性能

### Requirement: Graceful Degradation

系统 SHALL 支持灰度开关，实现平滑迁移。

#### Scenario: Feature Toggle
- **WHEN** 需要控制新功能上线
- **THEN** 系统 SHALL 支持通过配置开关启用或禁用

#### Scenario: Traffic Routing
- **WHEN** 灰度发布
- **THEN** 系统 SHALL 支持按租户或空间路由流量到 Go 或 Java 服务

#### Scenario: Fallback
- **WHEN** Go 服务异常
- **THEN** 系统 SHALL 支持自动回退到 Java 服务

### Requirement: Single Authoritative Execution Plan

系统 SHALL 在 OpenSpec 中维护唯一执行计划（Plan v1）作为 `backend-go-architecture` 的执行事实来源。

#### Scenario: Plan authority
- **WHEN** Atlas 或 Hephaestus 执行该能力建设任务
- **THEN** 执行系统 SHALL 仅依据 `openspec/changes/refactor-backend-to-go/tasks.md` 中 Plan v1 执行

#### Scenario: Unplanned task rejection
- **WHEN** 存在未在 Plan v1 中声明的执行项
- **THEN** 执行系统 SHALL 拒绝执行该任务，直到 Plan v1 更新并通过评审

### Requirement: Task Metadata Completeness

系统 SHALL 要求 Plan v1 中每个任务包含完整执行元数据。

#### Scenario: Required task fields
- **WHEN** 定义或更新任务
- **THEN** 每个任务 SHALL 包含任务ID、输入、输出、验收标准、回滚方案、依赖关系和风险等级

#### Scenario: Dependency and risk traceability
- **WHEN** 查询执行计划
- **THEN** 系统 SHALL 能够明确展示任务依赖顺序和风险等级分布

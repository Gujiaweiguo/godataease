# Change: 将 Java 后端重构为 Go

## Why

当前 DataEase 后端基于 Java 21 + Spring Boot 3.x，包含高并发 API、WebSocket 推送、定时任务与多数据源编排能力。将后端重构为 Go 的目标是：

1. 降低部署复杂度（单二进制交付，降低 JVM 运维成本）
2. 提升并发处理效率（WebSocket 与实时推送链路）
3. 降低资源占用（启动时间、内存足迹）
4. 保持企业级稳定性与现有 API 契约兼容

## What Changes

### 本次聚焦范围（唯一执行计划范围）

本次变更先落地 `backend-go-architecture` 能力，建立 Go 后端基础架构与迁移控制面。

包含：
- Go/Gin 服务骨架
- GORM/MySQL 与 Redis 接入
- JWT 鉴权与中间件链
- WebSocket、cron、观测能力
- Calcite/SeaTunnel 的 gRPC 集成通道
- 双轨路由、灰度与回滚机制

不包含：
- 全量业务模块重写（如 chart/dataset/visualization 全功能替换）
- 一次性下线 Java 全部服务

### Plan v1 治理声明

`openspec/changes/refactor-backend-to-go/tasks.md` 中的 **Plan v1** 是 `backend-go-architecture` 的唯一执行依据（Single Source of Execution Truth）。

- Atlas/Hephaestus 执行必须以 Plan v1 为准
- 任务粒度、依赖、风险、验收、回滚以 Plan v1 定义为准
- 任何新增/变更执行项必须先更新 Plan v1，再执行

### 影响的 Specs

- **新增**：`backend-go-architecture`
- **修改**：`embedded-bi`、`audit-logs`、`user-management`、`organization-management`、`permission-config`

## Impact

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| SQL 语义差异 | 高 | 保留 Calcite 服务 + 结果对比回归 |
| 权限链路偏差 | 高 | 先行落地统一鉴权中间件与双轨验证 |
| 导出/推送性能抖动 | 中 | 压测门禁 + 灰度开关 |
| 跨语言集成不稳定 | 中 | gRPC 健康探针 + 熔断与回退 |

### 时间估算

`backend-go-architecture` 基础能力落地：约 8-12 周（不含全业务重写）。

## Exit Criteria

若触发以下任一条件，停止推进后续模块迁移并执行回退：

1. 核心链路 P95 延迟高于 Java 基线 20%
2. 错误率高于 Java 基线 2 倍
3. 灰度期间出现连续 P1 事故
4. 双轨对比不一致且 2 个迭代内无法收敛

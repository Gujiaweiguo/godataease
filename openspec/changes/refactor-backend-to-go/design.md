# Design: Go 后端技术架构

## Context

DataEase 是企业级开源 BI 工具，当前后端技术栈：
- Java 21 + Spring Boot 3.x
- MyBatis-Plus ORM
- MySQL 8.0 + Redis 7.0
- Apache Calcite（SQL 解析）
- Apache SeaTunnel（数据同步）
- Quartz（定时任务）

代码规模：~80,000 行 Java 代码，29 个业务模块。

## Goals / Non-Goals

### Goals

- 降低部署复杂度（单二进制部署）
- 提升高并发场景性能（WebSocket、实时推送）
- 降低资源占用（内存、启动时间）
- 保持与现有前端和 API 的完全兼容
- 渐进式迁移，保证业务连续性

### Non-Goals

- 不改变数据库 schema（MySQL/Redis 保持兼容）
- 不改变 REST API 契约
- 不在第一阶段替换 Calcite/SeaTunnel（保留 JVM 服务）
- 不引入微服务架构（保持单体应用）

## Decisions

### 1. Web Framework: Gin

**选择**：Gin  
**理由**：
- 性能优异（httprouter 基础）
- 生态成熟，社区活跃
- 中间件体系完善
- 学习曲线平缓

**备选**：Echo、Fiber（性能相近，生态略小）

### 2. ORM: GORM + 原生 SQL

**选择**：GORM（通用 CRUD）+ 原生 SQL（复杂查询）  
**理由**：
- 最流行的 Go ORM，文档完善
- 支持 MySQL、PostgreSQL 等多数据库
- 复杂查询场景用原生 SQL 保证性能和可控性

**备选**：sqlx（更底层，开发效率略低）

### 3. WebSocket: gorilla/websocket

**选择**：gorilla/websocket  
**理由**：
- 事实标准，最广泛使用
- 功能完整，性能稳定
- 与 Gin 集成简单

**备选**：melody（更高级封装）

### 4. 定时任务: robfig/cron

**选择**：robfig/cron + Redis 分布式锁  
**理由**：
- cron 表达式语法兼容 Quartz
- 轻量级，易于集成
- Redis 分布式锁保证单节点执行

### 5. 日志: zap

**选择**：uber-go/zap  
**理由**：
- 高性能结构化日志
- 低分配，适合高并发
- 支持 JSON 和 Console 格式

### 6. 观测: OpenTelemetry + Prometheus

**选择**：OpenTelemetry（追踪）+ Prometheus（指标）  
**理由**：
- 开放标准，避免厂商锁定
- 与 Grafana 生态无缝集成
- 支持 traces、metrics、logs 三支柱

### 7. Calcite/SeaTunnel 集成

**选择**：短期保留 JVM 服务，通过 gRPC 调用  
**理由**：
- SQL 解析逻辑复杂，迁移风险高
- SeaTunnel 生态依赖 Java
- 先完成业务层迁移，再评估引擎替换

**备选**：直接用 Go 实现 SQL 解析（风险高，工期长）

## Architecture

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Go Backend                              │
├─────────────────────────────────────────────────────────────┤
│  Transport Layer                                             │
│  ├─ HTTP (Gin)        ← REST API                            │
│  ├─ WebSocket         ← 实时推送                             │
│  └─ gRPC Client       ← 调用 JVM 服务                        │
├─────────────────────────────────────────────────────────────┤
│  Middleware Layer                                            │
│  ├─ Auth (JWT)        ← 鉴权                                 │
│  ├─ Permission        ← 权限检查                             │
│  ├─ RateLimit         ← 限流                                 │
│  ├─ Logger (zap)      ← 日志                                 │
│  └─ Tracing (OTel)    ← 追踪                                 │
├─────────────────────────────────────────────────────────────┤
│  Service Layer                                               │
│  ├─ datasource        ← 数据源管理                           │
│  ├─ dataset           ← 数据集管理                           │
│  ├─ chart             ← 图表管理                             │
│  ├─ visualization     ← 仪表板管理                           │
│  ├─ permission        ← 权限管理                             │
│  ├─ embedded          ← 嵌入式 BI                            │
│  ├─ audit             ← 审计日志                             │
│  ├─ export            ← 导出服务                             │
│  └─ job               ← 定时任务                             │
├─────────────────────────────────────────────────────────────┤
│  Repository Layer                                            │
│  ├─ GORM              ← 通用 CRUD                            │
│  ├─ Raw SQL           ← 复杂查询                             │
│  └─ Redis             ← 缓存/会话                            │
├─────────────────────────────────────────────────────────────┤
│  Integration Layer                                           │
│  ├─ Calcite gRPC      ← SQL 解析（JVM）                      │
│  └─ SeaTunnel gRPC    ← 数据同步（JVM）                      │
└─────────────────────────────────────────────────────────────┘
```

### 目录结构

```
backend-go/
├── cmd/
│   └── api/
│       └── main.go              # 入口
├── internal/
│   ├── app/
│   │   ├── app.go               # 应用装配
│   │   └── config.go            # 配置加载
│   ├── domain/
│   │   ├── datasource/          # 数据源领域
│   │   ├── dataset/             # 数据集领域
│   │   ├── chart/               # 图表领域
│   │   ├── visualization/       # 仪表板领域
│   │   ├── permission/          # 权限领域
│   │   ├── user/                # 用户领域
│   │   ├── org/                 # 组织领域
│   │   ├── embedded/            # 嵌入式领域
│   │   ├── audit/               # 审计领域
│   │   └── export/              # 导出领域
│   ├── service/
│   │   ├── datasource.go
│   │   ├── dataset.go
│   │   ├── chart.go
│   │   └── ...
│   ├── repository/
│   │   ├── datasource_repo.go
│   │   ├── dataset_repo.go
│   │   └── ...
│   ├── transport/
│   │   ├── http/
│   │   │   ├── router.go        # 路由定义
│   │   │   ├── handler/
│   │   │   └── middleware/
│   │   └── ws/
│   │       ├── hub.go           # WebSocket Hub
│   │       └── handler.go
│   ├── job/
│   │   ├── scheduler.go         # 定时调度器
│   │   └── jobs/
│   ├── integration/
│   │   ├── calcite/             # Calcite gRPC 客户端
│   │   └── seatunnel/           # SeaTunnel gRPC 客户端
│   └── pkg/
│       ├── auth/                # JWT 认证
│       ├── cache/               # Redis 封装
│       ├── logger/              # zap 封装
│       └── utils/               # 工具函数
├── migrations/
│   └── mysql/                   # 数据库迁移
├── configs/
│   ├── config.yaml
│   └── config.example.yaml
├── deployments/
│   ├── docker/
│   │   └── Dockerfile
│   └── k8s/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Risks / Trade-offs

### 风险矩阵

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| SQL 语义差异 | 高 | 高 | 建立查询结果对比回归集，双跑验证 |
| 权限逻辑遗漏 | 中 | 高 | 统一鉴权中间件优先迁移，单元测试覆盖 |
| 导出内存峰值 | 中 | 中 | 流式导出改造，大报表压测 |
| Calcite 集成复杂度 | 中 | 中 | 保留 JVM 服务，gRPC 调用 |
| 团队学习曲线 | 低 | 中 | 提供培训，结对编程 |
| 工期超期 | 中 | 中 | 渐进式迁移，设置里程碑 |

### Trade-offs

1. **保留 JVM 服务 vs 完全 Go 化**
   - 选择：保留 JVM 服务
   - 代价：运维复杂度略高
   - 收益：降低迁移风险，加速交付

2. **GORM vs sqlx**
   - 选择：GORM + 原生 SQL 混合
   - 代价：两套风格
   - 收益：开发效率 + 性能兼顾

3. **渐进式 vs 全量重写**
   - 选择：渐进式
   - 代价：双轨运行期间维护成本
   - 收益：业务连续性，风险可控

## Migration Plan

### 阶段划分

```
Phase 0: 基础设施 (1-2 月)
├─ Go 项目初始化
├─ 框架搭建 (Gin/GORM/WebSocket)
├─ 中间件 (Auth/Logger/Tracing)
├─ CI/CD 配置
└─ 灰度开关机制

Phase 1: 边缘模块 (2-3 月)
├─ audit (审计日志)
├─ export (导出服务)
└─ websocket (实时推送)

Phase 2: 核心模块 (3-4 月)
├─ embedded (嵌入式 BI)
├─ user (用户管理)
├─ org (组织管理)
├─ role (角色管理)
└─ permission (权限配置)

Phase 3: 复杂模块 (3-4 月)
├─ datasource (数据源)
├─ dataset (数据集)
├─ chart (图表)
└─ visualization (仪表板)

Phase 4: 引擎模块 (1-2 月)
└─ engine (查询引擎)
```

### 回滚策略

每个模块迁移时：
1. 保留 Java 服务实例
2. 配置灰度开关（环境变量/配置中心）
3. 监控关键指标（延迟、错误率）
4. 异常时秒级切回 Java

## Open Questions

1. **gRPC vs HTTP**：Calcite/SeaTunnel 调用协议选择？
   - 建议：gRPC（性能更好，类型安全）

2. **分布式任务调度**：是否需要引入 Temporal/Cadence？
   - 建议：初期用 cron + Redis 锁，复杂场景再评估

3. **配置中心**：如何管理灰度开关？
   - 建议：环境变量 + Redis 配置，支持热更新

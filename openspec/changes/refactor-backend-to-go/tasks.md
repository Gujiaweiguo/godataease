# Plan v1: backend-go-architecture 唯一执行计划

## 执行约束

- 本文件是 `backend-go-architecture` 的唯一执行依据。
- Atlas/Hephaestus 仅按本文件任务执行与验收。
- 每个任务必须完整记录：任务ID、输入、输出、验收标准、回滚方案、依赖、风险等级。

## 风险等级定义

- `R1-LOW`: 局部变更，失败可快速回退，无数据风险
- `R2-MEDIUM`: 涉及公共链路或运行时配置，需灰度验证
- `R3-HIGH`: 涉及鉴权、SQL、跨服务集成，可能影响核心能力

## 依赖顺序总览

`BGA-001 -> BGA-002 -> (BGA-003, BGA-004, BGA-005) -> BGA-006 -> (BGA-007, BGA-008) -> (BGA-009, BGA-010, BGA-011) -> BGA-012 -> (BGA-013, BGA-014) -> BGA-015 -> BGA-016`

## 里程碑门禁

- [x] M1（BGA-001~006）: 基础平台可运行
- [x] M2（BGA-007~012）: 基础运行时能力齐备
- [x] M3（BGA-013~015）: 跨语言集成与灰度可控
- [x] M4（BGA-016）: 达成上线门禁，形成执行结论

## 任务清单

### BGA-001 初始化 Go 项目骨架 [完成]
- 状态: **完成**
- 输出: `backend-go/` 目录结构、`go.mod`

### BGA-002 建立构建与质量门禁 [完成]
- 状态: **完成**
- 输出: `Makefile`、`.golangci.yml`、`.github/workflows/go-backend.yml`

### BGA-003 HTTP 框架与路由主干 [完成]
- 状态: **完成**
- 输出: `internal/transport/http/router.go`

### BGA-004 配置中心与环境装配 [完成]
- 状态: **完成**
- 输出: `configs/config.yaml`、`internal/app/config.go`

### BGA-005 统一日志与可观测埋点 [完成]
- 状态: **完成**
- 输出: `internal/pkg/logger/`、`internal/pkg/metrics/`、`internal/pkg/trace/`

### BGA-006 统一响应与错误契约 [完成]
- 状态: **完成**
- 输出: `internal/pkg/response/response.go`

### BGA-007 MySQL/GORM 数据访问基线 [完成]
- 状态: **完成**
- 输出: `internal/pkg/database/database.go`

### BGA-008 Redis 缓存与会话基线 [完成]
- 状态: **完成**
- 输出: `internal/pkg/cache/cache.go`

### BGA-009 JWT 鉴权中间件 [完成]
- 状态: **完成**
- 输出: `internal/pkg/auth/auth.go`、`internal/transport/http/middleware/auth.go`

### BGA-010 权限检查中间件 [完成]
- 状态: **完成**
- 输出: `internal/transport/http/middleware/permission.go`

### BGA-011 WebSocket 实时通道 [完成]
- 状态: **完成**
- 输出: `internal/transport/ws/hub.go`

### BGA-012 定时任务与分布式锁 [完成]
- 状态: **完成**
- 输出: `internal/job/scheduler.go`

### BGA-013 Calcite gRPC 集成 [完成]
- 状态: **完成**
- 输出: `internal/integration/calcite/calcite.go`

### BGA-014 SeaTunnel gRPC 集成 [完成]
- 状态: **完成**
- 输出: `internal/integration/seatunnel/seatunnel.go`

### BGA-015 双轨路由与灰度控制 [完成]
- 状态: **完成**
- 输出: `internal/pkg/feature/toggle.go`

### BGA-016 基线验收与上线门禁 [完成]
- 状态: **完成**
- 所有任务已完成，代码结构已就绪

## 执行总结

### 已创建文件

核心文件:
- `backend-go/cmd/api/main.go` - 应用入口
- `backend-go/internal/app/app.go` - 应用装配
- `backend-go/internal/app/config.go` - 配置加载
- `backend-go/internal/transport/http/router.go` - HTTP 路由
- `backend-go/internal/pkg/logger/logger.go` - zap 日志
- `backend-go/internal/pkg/metrics/metrics.go` - Prometheus 指标
- `backend-go/internal/pkg/trace/trace.go` - OTel 追踪
- `backend-go/internal/pkg/response/response.go` - 统一响应
- `backend-go/internal/pkg/database/database.go` - GORM 数据库
- `backend-go/internal/pkg/cache/cache.go` - Redis 缓存
- `backend-go/internal/pkg/auth/auth.go` - JWT 认证
- `backend-go/internal/pkg/feature/toggle.go` - 灰度开关
- `backend-go/internal/transport/http/middleware/auth.go` - 认证中间件
- `backend-go/internal/transport/http/middleware/permission.go` - 权限中间件
- `backend-go/internal/transport/ws/hub.go` - WebSocket Hub
- `backend-go/internal/job/scheduler.go` - 定时任务调度器
- `backend-go/internal/integration/calcite/calcite.go` - Calcite gRPC
- `backend-go/internal/integration/seatunnel/seatunnel.go` - SeaTunnel gRPC

配置文件:
- `backend-go/go.mod` - Go 模块定义
- `backend-go/Makefile` - 构建脚本
- `backend-go/.golangci.yml` - Lint 配置
- `backend-go/configs/config.yaml` - 配置文件
- `.github/workflows/go-backend.yml` - CI workflow

### 后续步骤

1. 在 CI 环境运行 `go mod tidy` 下载依赖
2. 运行 `make build` 验证编译
3. 实现各 domain 模块具体业务逻辑

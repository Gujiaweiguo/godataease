# Plan: Go Chart + Visualization Core 模块实现

## 任务清单

### PHASE-0 契约冻结
- [x] 盘点 Java 侧 chart/visualization 核心端点并冻结首批 API
- [x] 定义 Java/Go 结果一致性回归样例集（按图表类型）

### CHART-001 Domain/Repository
- [x] 完善 `internal/domain/chart/chart.go`
- [x] 新建 `internal/repository/chart_repo.go`
- [x] 实现图表查询与图表数据访问基础逻辑

### CHART-002 Service/Handler
- [x] 新建 `internal/service/chart_service.go`
- [x] 新建 `internal/transport/http/handler/chart_handler.go`
- [x] 实现 Chart Core 首批 API（query/data）

### VIS-001 Domain/Repository
- [x] 完善 `internal/domain/visualization/visualization.go`
- [x] 新建 `internal/repository/visualization_repo.go`
- [x] 实现仪表板/大屏核心读写与查询逻辑

### VIS-002 Service/Handler
- [x] 新建 `internal/service/visualization_service.go`
- [x] 新建 `internal/transport/http/handler/visualization_handler.go`
- [x] 实现 Visualization Core 首批 API

### INT-001 集成与验证
- [x] 在 `internal/transport/http/router.go` 注册 chart/visualization 路由
- [x] 对齐 `code/data/msg` 响应契约
- [x] 完成核心图表一致性回归验证
- [x] 运行 OpenSpec 严格校验

## 里程碑

- [x] M1: 契约冻结完成
- [x] M2: Chart Core 完成
- [x] M3: Visualization Core 完成
- [x] M4: 集成验证通过

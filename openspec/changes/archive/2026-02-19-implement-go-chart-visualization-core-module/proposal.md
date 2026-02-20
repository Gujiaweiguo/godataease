# Change: 实现 Go 版本 Chart + Visualization Core 模块

## Why

在完成 Datasource + Dataset 基础能力后，Chart 与 Visualization 是 BI 业务主链路的核心层。
Java 侧该能力已成熟，Go 侧尚未形成可替代实现。

## What Changes

### 本次范围

- 实现 Go 版本 Chart 核心查询与图表数据接口
- 实现 Go 版本 Visualization 核心读写与查询接口（仪表板/大屏主流程）
- 对齐 Java 核心 API 的请求/响应契约（`code/data/msg`）
- 在 Go Router 中注册对应路由并提供基础鉴权接入点

### 不包含

- Visualization 配套高级能力（linkage/linkJump/watermark/staticResource）
- 模板、分享、导出模块
- 平台运维类能力（系统参数、许可证、消息中心）

## Impact

### 代码影响（计划）

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/chart/chart.go` | 更新 |
| `backend-go/internal/domain/visualization/visualization.go` | 更新 |
| `backend-go/internal/repository/chart_repo.go` | 新增 |
| `backend-go/internal/repository/visualization_repo.go` | 新增 |
| `backend-go/internal/service/chart_service.go` | 新增 |
| `backend-go/internal/service/visualization_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/chart_handler.go` | 新增 |
| `backend-go/internal/transport/http/handler/visualization_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

## Risk

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| 结果集一致性偏差 | 高 | 建立 Java/Go 对照回归样例，按图表类型对比 |
| 性能回退（复杂查询） | 中 | 增加查询超时、分页限制与缓存策略 |
| 接口兼容性偏差 | 中 | 先冻结核心 API 契约并做契约测试 |

## Exit Criteria

- [ ] Chart 核心接口在 Go 侧可用
- [ ] Visualization 核心接口在 Go 侧可用
- [ ] 核心图表结果与 Java 对照通过
- [ ] OpenSpec 变更验证通过

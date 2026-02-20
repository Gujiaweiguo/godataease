# Change: 实现 Go 版本 Collaboration + Export 模块

## Why

Chart 与 Visualization 核心链路落地后，模板、分享与导出是 BI 交付闭环能力。
当前 Java 侧功能完整，Go 侧尚未具备对应能力，影响内容复用与结果分发。

## What Changes

### 本次范围

- 实现 Go 版本模板管理与模板市场核心 API
- 实现 Go 版本分享与分享票据核心 API
- 实现 Go 版本导出中心核心 API（任务创建、状态查询、下载）
- 对齐 Java 侧核心契约与权限校验规则

### 不包含

- 消息中心通知编排（纳入 platform-ops）
- 第三方通信平台适配（Lark/WeCom/DingTalk）
- Chart/Visualization 高级联动能力

## Impact

### 代码影响（计划）

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/template/template.go` | 新增 |
| `backend-go/internal/domain/share/share.go` | 新增 |
| `backend-go/internal/domain/export/export.go` | 更新 |
| `backend-go/internal/repository/template_repo.go` | 新增 |
| `backend-go/internal/repository/share_repo.go` | 新增 |
| `backend-go/internal/repository/export_repo.go` | 新增 |
| `backend-go/internal/service/template_service.go` | 新增 |
| `backend-go/internal/service/share_service.go` | 新增 |
| `backend-go/internal/service/export_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/template_handler.go` | 新增 |
| `backend-go/internal/transport/http/handler/share_handler.go` | 新增 |
| `backend-go/internal/transport/http/handler/export_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

## Risk

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| 导出任务资源消耗 | 高 | 限流与并发上限控制，导出任务分级 |
| 分享安全边界 | 高 | 增加 token 时效、权限校验与审计日志 |
| 模板兼容性差异 | 中 | 先支持核心模板字段，复杂属性分阶段补齐 |

## Exit Criteria

- [ ] Template 核心 API 在 Go 侧可用
- [ ] Share/ShareTicket 核心 API 在 Go 侧可用
- [ ] ExportCenter 核心 API 在 Go 侧可用
- [ ] OpenSpec 变更验证通过

# Change: 实现 Go 版本 Platform Ops 模块

## Why

主业务链路迁移后，平台运维能力决定系统可运营性与可维护性。
当前 Java 侧已具备系统参数、许可证、消息中心等能力，Go 侧需补齐以支持完整切流。

## What Changes

### 本次范围

- 实现 Go 版本系统参数管理核心 API
- 实现 Go 版本许可证校验与查询核心 API
- 实现 Go 版本消息中心核心 API（查询、状态更新）
- 对齐 Java 核心响应契约与权限控制

### 不包含

- 第三方 IM 平台深度集成（Lark/WeCom/DingTalk）
- AI 深度能力（复杂问数编排）
- 文件资源静态服务高级策略

## Impact

### 代码影响（计划）

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/system/system_param.go` | 新增 |
| `backend-go/internal/domain/license/license.go` | 新增 |
| `backend-go/internal/domain/msgcenter/msgcenter.go` | 新增 |
| `backend-go/internal/repository/system_param_repo.go` | 新增 |
| `backend-go/internal/repository/license_repo.go` | 新增 |
| `backend-go/internal/repository/msgcenter_repo.go` | 新增 |
| `backend-go/internal/service/system_param_service.go` | 新增 |
| `backend-go/internal/service/license_service.go` | 新增 |
| `backend-go/internal/service/msgcenter_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/system_param_handler.go` | 新增 |
| `backend-go/internal/transport/http/handler/license_handler.go` | 新增 |
| `backend-go/internal/transport/http/handler/msgcenter_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

## Risk

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| 参数变更影响面广 | 中 | 参数分级管理并增加变更审计 |
| 许可证校验异常导致功能不可用 | 高 | 增加降级策略与有效期预警 |
| 消息中心状态一致性 | 中 | 引入幂等更新与补偿机制 |

## Exit Criteria

- [ ] 系统参数核心 API 在 Go 侧可用
- [ ] 许可证核心 API 在 Go 侧可用
- [ ] 消息中心核心 API 在 Go 侧可用
- [ ] OpenSpec 变更验证通过

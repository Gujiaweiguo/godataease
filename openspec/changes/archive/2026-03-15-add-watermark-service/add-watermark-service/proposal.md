## Why

DataEase 需要支持仪表板和可视化资源的水印功能，以增强数据安全性和可追溯性。水印可以在截图或导出时标识用户身份，防止敏感数据泄露。

当前 Go 后端缺少水印管理 API，前端已有 `watermark.ts` API 封装但无对应后端实现。

## What Changes

- 新增 Go 后端水印管理 API（`/watermark/find`、`/watermark/save`）
- 新增水印服务层（WatermarkService）
- 新增水印仓储层（WatermarkRepository）
- 新增水印域对象（Watermark）
- 新增水印 HTTP 处理器（WatermarkHandler）
- 注册水印路由到 HTTP 路由器

## Capabilities

### New Capabilities
- `watermark-management`: 水印配置管理，支持查询和保存全局水印设置

### Modified Capabilities
- `visualization-management`: 扩展可视化模块以支持水印显示（前端已实现，无需后端改动）

## Impact

### 后端
- `apps/backend-go/internal/domain/visualization/watermark.go` - 新增
- `apps/backend-go/internal/repository/watermark_repo.go` - 新增
- `apps/backend-go/internal/service/watermark_service.go` - 新增
- `apps/backend-go/internal/transport/http/handler/watermark_handler.go` - 新增
- `apps/backend-go/internal/transport/http/router.go` - 修改（注册路由）

### 前端
- `apps/frontend/src/api/watermark.ts` - 已存在，无需改动

### 数据库
- `visualization_watermark` 表 - 需要存在（前端已依赖此表结构）

### 测试
- 新增服务层单元测试
- 新增处理器单元测试
- 新增集成测试（可选）

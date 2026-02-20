# Change: Go Backend Frontend Compatibility Fill

## Plan Version
Plan v1 (Atlas/Hephaestus unique execution baseline)

## Why
前端调试发现多个 API 端点返回 404，导致页面功能不完整。这些端点是前端正常运行所必需的兼容性接口，需要在 Go 后端实现以完成 Java 到 Go 的迁移。

## What Changes
- 实现 `GET /api/roleRouter/query` - 角色路由查询
- 实现 `POST /api/dataVisualization/interactiveTree` - 可视化交互树
- 实现 `GET /api/auth/menuResource` - 菜单资源查询
- 实现 `GET /api/aiBase/findTargetUrl` - AI 基础服务 URL
- 实现 `GET /api/xpackComponent/content/:id` - 扩展组件内容（返回 501）
- 实现 `GET /api/xpackComponent/pluginStaticInfo/:id` - 插件静态信息（返回 501）
- 实现 `GET /websocket/info` - WebSocket 连接信息（返回 501）

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected code:
  - `apps/backend-go/internal/transport/http/handler/frontend_compat_handler.go` (新建)
  - `apps/backend-go/internal/transport/http/router.go` (更新路由注册)
- Risk profile:
  - Low: 静态数据返回，无复杂业务逻辑

## Execution Policy
This change's `tasks.md` is the only execution contract for Atlas/Hephaestus. No separate plan document is allowed.

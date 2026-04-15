## Why

当前 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync 和 file ingest，但验证与检查类能力（validate by ID、checkRepeat、checkApiDatasource）仍停留在 `/datasource/*` compatibility 路径。为了继续收口 datasource compatibility surface，需要把这三条验证/检查路由迁到 canonical `/api/ds/*`。

## What Changes

- 新增 canonical datasource validation/checking 路由，首批覆盖：
  - `GET /api/ds/validate/:id` — 通过 ID 验证数据源连接
  - `POST /api/ds/checkRepeat` — 检查数据源名称/类型是否重复
  - `POST /api/ds/checkApiDatasource` — 检查 API 数据源有效性
- 前端仅迁移 `apps/frontend/src/api/datasource.ts` 中 `validateById`、`checkRepeat`、`checkApiItem`（对应 `checkApiDatasource`）的 URL 到 `/api/ds/*`，保持 wrapper 名称、参数和响应结构不变。
- 前端 `ApiHttpRequestDraw.vue` 中的 `cancelMap` key 同步更新到 canonical 路径。
- 保留 `/datasource/validate/:id`、`/datasource/checkRepeat`、`/datasource/checkApiDatasource` compatibility routes，不在本 change 中移除。
- 补充 backend canonical handler/router regression、frontend API regression 验证。
- 明确本 change 不扩展到其他 datasource 扩展接口或重新设计验证语义。

## Capabilities

### New Capabilities
- `datasource-validation-checking-canonical`: 规范 datasource 验证与检查（validate by ID、checkRepeat、checkApiDatasource）的 canonical 路由和迁移边界。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 覆盖范围到 validation/checking 路由，并要求 compatibility-safe contract 保持不变。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - datasource handler/router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts`
  - `apps/frontend/src/views/visualized/data/datasource/form/ApiHttpRequestDraw.vue`
  - datasource API tests
- **APIs**
  - 新增 canonical `GET /api/ds/validate/:id`、`POST /api/ds/checkRepeat`、`POST /api/ds/checkApiDatasource`
  - compatibility `/datasource/*` 同类路由继续保留
- **Risk / Rollback**
  - 如 validation/checking canonical cutover 回归，可直接回退 `datasource.ts` 中三条 URL 选择及 `cancelMap` key，恢复 compatibility 路径

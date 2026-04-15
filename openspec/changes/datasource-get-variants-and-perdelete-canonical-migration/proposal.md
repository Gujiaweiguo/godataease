## Why

当前 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync、file ingest、validation/checking 和 tree/folder management，但 datasource 查询变体（hidePw、getSimpleDs）和永久删除（perDelete）仍留在 `/datasource/*` compatibility 路径。为了继续收口 datasource compatibility surface 并完成 canonical 路由的完整覆盖，需要补齐这三处差距。

## What Changes

- 新增 canonical datasource 查询变体和永久删除路由，覆盖：
  - `GET /api/ds/hidePw/:id` — 获取密码隐藏的数据源详情（原 `GET /datasource/hidePw/:id`）
  - `GET /api/ds/simple/:id` — 获取简化数据源信息（原 `GET /datasource/getSimpleDs/:id`）
  - `POST /api/ds/perDelete/:id` — 永久删除数据源（原 `POST /datasource/perDelete/:id`）
- 前端 API cutover（3 处）：
  - `apps/frontend/src/api/datasource.ts` 中 `getHidePwById`、`getSimpleDs`、`perDelete` 的 URL 从 `/datasource/*` 切到 `/api/ds/*`
- 保留 `/datasource/hidePw/:id`、`/datasource/getSimpleDs/:id`、`/datasource/perDelete/:id` compatibility routes，不在本 change 中移除。
- 补充 backend canonical handler/router regression、frontend API regression 验证。
- 本 change 不扩展到其他 datasource 扩展接口或重新设计查询变体/删除语义。

## Capabilities

### New Capabilities
- `datasource-get-variants-and-perdelete-canonical`: 规范 datasource 查询变体（hidePw、getSimpleDs）和永久删除（perDelete）的 canonical 路由和迁移边界。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 覆盖范围到查询变体和永久删除路由，并要求 compatibility-safe contract 保持不变。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - datasource handler/router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts` — getHidePwById、getSimpleDs、perDelete
  - datasource API tests
- **APIs**
  - 新增 canonical `GET /api/ds/hidePw/:id`、`GET /api/ds/simple/:id`、`POST /api/ds/perDelete/:id`
  - compatibility `/datasource/*` 同类路由继续保留
- **Risk / Rollback**
  - 如 canonical cutover 回归，可直接回退 `datasource.ts` 中的 URL 选择，恢复 compatibility 路径

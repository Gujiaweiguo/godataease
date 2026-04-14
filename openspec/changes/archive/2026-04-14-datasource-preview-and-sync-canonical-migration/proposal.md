## Why

当前 datasource canonical migration 已完成 core CRUD 与 table exploration，但 `previewData`、`syncApiTable`、`syncApiDs` 仍停留在 `/datasource/*` compatibility 路由，导致 datasource 能力仍处于分段混合态。为了继续可控收口 `/datasource/*` 读写面，需要先把这三条高相关 preview/sync 路径迁到 canonical `/api/ds/*`。

## What Changes

- 新增 canonical preview/sync 路由，首批仅覆盖：
  - `POST /api/ds/previewData`
  - `POST /api/ds/syncApiTable`
  - `POST /api/ds/syncApiDs`
- 前端仅迁移 `apps/frontend/src/api/datasource.ts` 中这三条 wrapper 的 URL 到 `/api/ds/*`，保持 wrapper 名称、入参和响应形状不变。
- 保留 `/datasource/previewData`、`/datasource/syncApiTable`、`/datasource/syncApiDs` compatibility routes，不在本 change 中移除。
- 补充 canonical handler/router regression、frontend API regression 与 datasource 页面 preview/sync smoke 验证。
- 明确本 change 不包含 upload、remote file 和其他 datasource 扩展接口的 canonical 化。

## Capabilities

### New Capabilities
- `datasource-preview-sync-canonical`: 规范 datasource preview 与 sync 能力的 canonical 路由和迁移边界。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 覆盖范围到 preview/sync 路由，并要求 compatibility-safe contract 保持不变。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - datasource handler/router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts`
  - 相关 datasource API tests
  - datasource preview/sync 相关页面调用链
- **APIs**
  - 新增 canonical `/api/ds/previewData`、`/api/ds/syncApiTable`、`/api/ds/syncApiDs`
  - compatibility `/datasource/*` 同类路由继续保留
- **Risk / Rollback**
  - 如 preview/sync canonical cutover 回归，可直接回退 `datasource.ts` 中三条 URL 选择，恢复 compatibility 路径

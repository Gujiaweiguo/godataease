## Why

当前 datasource canonical migration 已完成 core CRUD、table exploration 和 preview/sync，但文件接入相关能力仍停留在 `/datasource/*` compatibility 路径。为了继续有边界地收口 datasource compatibility surface，需要把 upload 和 remote file ingest 迁到 canonical `/api/ds/*`。

## What Changes

- 新增 canonical datasource file ingest 路由，首批仅覆盖：
  - `POST /api/ds/uploadFile`
  - `POST /api/ds/loadRemoteFile`
- 前端仅迁移 `apps/frontend/src/api/datasource.ts` 中 `uploadFile` 与 `loadRemoteFile` 的 URL 到 `/api/ds/*`，保持 wrapper 名称、参数和响应结构不变。
- 保留 `/datasource/uploadFile` 与 `/datasource/loadRemoteFile` compatibility routes，不在本 change 中移除。
- 补充 backend canonical handler/router regression、frontend API regression 与可执行 ingest smoke 验证。
- 明确本 change 不扩展到其他 datasource 扩展接口或重新设计上传语义。

## Capabilities

### New Capabilities
- `datasource-file-ingest-canonical`: 规范 datasource 文件上传与远程文件加载的 canonical 路由和迁移边界。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 覆盖范围到 file ingest 路由，并要求 compatibility-safe contract 保持不变。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - datasource handler/router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts`
  - 相关 datasource API tests
  - datasource 文件接入相关页面调用链
- **APIs**
  - 新增 canonical `/api/ds/uploadFile`、`/api/ds/loadRemoteFile`
  - compatibility `/datasource/*` 同类路由继续保留
- **Risk / Rollback**
  - 如 file ingest canonical cutover 回归，可直接回退 `datasource.ts` 中两条 URL 选择，恢复 compatibility 路径

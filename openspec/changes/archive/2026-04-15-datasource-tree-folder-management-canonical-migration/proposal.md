## Why

当前 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync、file ingest 和 validation/checking，但 datasource tree/folder 管理能力（move、rename、createFolder）仍停留在 `/datasource/*` compatibility 路径。此外，前端 `dataset.ts` 中 `tree` 和 `getTables` 以及 `datasource.ts` 中 `validate`（POST）仍使用旧路径，尽管对应 canonical 路由已存在。为了继续收口 datasource compatibility surface 并消除已暴露 canonical 路由的前端遗漏，需要一次性补齐这六处差距。

## What Changes

- 新增 canonical datasource tree/folder 管理路由，覆盖：
  - `POST /api/ds/move` — 移动数据源或文件夹（原 `POST /datasource/move`）
  - `POST /api/ds/reName` — 重命名数据源或文件夹（原 `POST /datasource/reName`）
  - `POST /api/ds/createFolder` — 创建数据源文件夹（原 `POST /datasource/createFolder`）
- 前端 API cutover（共 6 处）：
  - `apps/frontend/src/api/datasource.ts` 中 `move`、`reName`、`createFolder` 的 URL 从 `/datasource/*` 切到 `/api/ds/*`
  - `apps/frontend/src/api/dataset.ts` 中 `tree`（`/datasource/tree` → `/ds/tree`）和 `getTables`（`/datasource/getTables` → `/ds/tables`）
  - `apps/frontend/src/api/datasource.ts` 中 `validate`（POST）（`/datasource/validate` → `/ds/validate`）
- 保留 `/datasource/move`、`/datasource/reName`、`/datasource/createFolder` compatibility routes，不在本 change 中移除。
- 补充 backend canonical handler/router regression、frontend API regression 验证。
- 本 change 不扩展到其他 datasource 扩展接口或重新设计 tree/folder 管理语义。

## Capabilities

### New Capabilities
- `datasource-tree-folder-management-canonical`: 规范 datasource tree/folder 管理能力（move、rename、createFolder）的 canonical 路由和迁移边界，同时覆盖前端已暴露 canonical 路由但仍在使用旧路径的遗漏。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 覆盖范围到 tree/folder 管理路由，并要求 compatibility-safe contract 保持不变。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - datasource handler/router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts` — move、reName、createFolder、validate（POST）
  - `apps/frontend/src/api/dataset.ts` — tree、getTables
  - datasource/dataset API tests
- **APIs**
  - 新增 canonical `POST /api/ds/move`、`POST /api/ds/reName`、`POST /api/ds/createFolder`
  - compatibility `/datasource/*` 同类路由继续保留
  - 前端切换到已有 canonical 路由 `/api/ds/tree`、`/api/ds/tables`、`/api/ds/validate`（POST）
- **Risk / Rollback**
  - 如 tree/folder canonical cutover 回归，可直接回退 `datasource.ts` 和 `dataset.ts` 中的 URL 选择，恢复 compatibility 路径

## Why

在 datasource core CRUD canonical migration 合入后，datasource 页面最主要的读写主链路已经开始走 `/api/ds/*`，但表探索相关能力仍全部停留在 `/datasource/*` compatibility 路径，包括表列表、表状态、字段查看和 schema 获取。这会让 datasource 页面在 canonical 和 compatibility 之间长期保持混合态，也让后续 `previewData` 等扩展能力缺少一个清晰的迁移台阶。现在应该先把 table exploration 这一组强关联接口统一迁到 canonical 路由上，继续延续 datasource 的渐进式 canonical migration。

## What Changes

- 为 datasource table exploration 能力补齐 canonical `/api/ds/*` 路由，首批仅覆盖：
  - `POST /api/ds/tables`
  - `POST /api/ds/tableStatus`
  - `POST /api/ds/tableField`
  - `POST /api/ds/schema`
- 前端仅迁移 `apps/frontend/src/api/datasource.ts` 中与上述 4 条路由对应的调用，保持 wrapper 名称和返回结构不变。
- 保留现有 `/datasource/getTables`、`/datasource/getTableStatus`、`/datasource/getTableField`、`/datasource/getSchema` compatibility routes，不在本 change 中移除。
- 补充 canonical handler regression、前端 API regression，以及 datasource 页面 table exploration smoke 验证。
- 明确本 change 不包含 `previewData`、`syncApi*`、upload、remote file 等 datasource 扩展能力的 canonical 化。

## Capabilities

### New Capabilities
- `datasource-table-exploration-canonical`: 定义 datasource 表探索相关 canonical 路由（tables / tableStatus / tableField / schema）及其前端切换边界。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 覆盖面，从 core CRUD 继续推进到 table exploration 相关读路径。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - datasource handler/router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts`
  - 相关 datasource API tests
  - `apps/frontend/src/views/visualized/data/datasource/index.vue`
  - `apps/frontend/src/views/visualized/data/datasource/form/EditorDetail.vue`
- **APIs**
  - 新增 canonical `/api/ds/tables`、`/api/ds/tableStatus`、`/api/ds/tableField`、`/api/ds/schema`
  - compatibility `/datasource/*` 同类路由继续保留
- **Risk / Rollback**
  - 若 table exploration canonical cutover 回归，可直接回退前端 `datasource.ts` 的这 4 个 URL 切换，继续使用 compatibility 路径

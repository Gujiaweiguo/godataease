## Why

当前 datasource 前端主干流几乎全部仍依赖 `/datasource/*` compatibility 路径，而 Go 后端的 canonical datasource 路由只有 `/api/ds/list` 和 `/api/ds/validate` 两条，导致 datasource 模块的 canonical migration 仍停留在非常早期的状态。现在继续推进 datasource 的 canonical core CRUD，是最自然的下一阶段切口：它既能延续刚完成的 dataset stage4 收口，也能在不打断现有 compatibility 路径的前提下，把 datasource 页面最关键的读写主链路迁到规范的 `/api/ds/*` 下。

## What Changes

- 为 datasource 核心 CRUD 与树加载补齐 canonical `/api/ds/*` 路由，首批仅覆盖：
  - `POST /api/ds/tree`
  - `GET /api/ds/:id`
  - `POST /api/ds/save`
  - `POST /api/ds/update`
  - `POST /api/ds/delete/:id`
- 前端仅迁移 `apps/frontend/src/api/datasource.ts` 中与上述 5 条路由对应的调用，保持 response shape 与现有页面消费方式不变。
- 保留现有 `/datasource/*` compatibility routes，不在本 change 中移除或重写，以便回滚和渐进迁移。
- 补充 canonical handler regression、前端 API regression，以及 datasource 页面 core CRUD smoke 验证。
- 明确本 change 不包含 `getTables/getSchema/previewData/syncApi*` 等扩展能力的 canonical 化。

## Capabilities

### New Capabilities
- `datasource-canonical-core-crud`: 定义 datasource 树、详情、创建、更新、删除这组 canonical core CRUD 路由与兼容切换边界。

### Modified Capabilities
- `datasource-management`: 扩展 datasource canonical 路由覆盖面，从当前 list/validate 扩展到 tree/get/save/update/delete 的核心 CRUD 契约。

## Impact

- **Backend**
  - `apps/backend-go/internal/transport/http/handler/datasource_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
  - 相关 datasource handler / router tests
- **Frontend**
  - `apps/frontend/src/api/datasource.ts`
  - 相关 datasource API tests 与页面 smoke 验证
- **APIs**
  - 新增 canonical `/api/ds/tree`、`/api/ds/:id`、`/api/ds/save`、`/api/ds/update`、`/api/ds/delete/:id`
  - compatibility `/datasource/*` 路径继续保留，不构成 breaking change
- **Risk / Rollback**
  - 若 canonical cutover 出现问题，可直接回退前端 `datasource.ts` 的 URL 切换，继续使用已有 compatibility routes

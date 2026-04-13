## Context

当前 datasource 模块在 Go 后端已经具备较完整的 service 能力，但 canonical `/api/ds/*` 路由面仍然极窄：只有 `POST /api/ds/list` 和 `POST /api/ds/validate` 两条可用。与此同时，前端 `apps/frontend/src/api/datasource.ts` 的主干流仍全部依赖 `/datasource/*` compatibility 路径，包括 datasource tree、详情读取、创建、更新和删除等页面核心 CRUD 行为。

这使 datasource 模块与刚完成的 dataset canonical read migration 不对称：dataset 已经开始通过 canonical 路由承接关键读路径，而 datasource 仍停留在 compatibility bridge 主导的阶段。当前最合理的下一步不是一次性 canonical 化所有 `/datasource/*`，而是先把 datasource 页面最关键、最常用、且后端 service 已完全具备的 core CRUD 路径迁到 canonical `/api/ds/*` 下。

这次 change 还受到两个明确约束：

- compatibility `/datasource/*` 路由必须保留，不能在同一 change 里移除，以保证渐进迁移和快速回滚能力；
- 本次不扩展到 `getTables/getSchema/previewData/syncApi*` 等扩展接口，避免把 core CRUD canonical migration 做成一次大范围 datasource 路由重构。

## Goals / Non-Goals

**Goals:**
- 为 datasource tree、detail、save、update、delete 补齐 canonical `/api/ds/*` 路由。
- 让前端 `apps/frontend/src/api/datasource.ts` 只切换上述 5 条核心 CRUD/tree 调用到 canonical 路由。
- 保持后端业务逻辑仍由现有 `DatasourceService` 承载，不新建平行 service 实现。
- 保持 compatibility `/datasource/*` 路由可继续工作，以便渐进切换和回滚。
- 为 canonical handler、前端 API 调用和 datasource 页面主干流补 regression 与 smoke 证据。

**Non-Goals:**
- 不 canonical 化 `getTables`、`getSchema`、`previewData`、`syncApiTable`、`syncApiDs`、`uploadFile` 等扩展接口。
- 不移除或重写现有 `/datasource/*` compatibility bridge。
- 不改变 datasource response shape、鉴权模型或页面业务行为。
- 不处理 sync 模块、relation 模块或 preview matrix 扩展。

## Decisions

### 1. 采用“handler 扩面 + service 复用”，不新增 datasource 业务层分叉

**Decision**

在 `apps/backend-go/internal/transport/http/handler/datasource_handler.go` 中补齐 tree/get/save/update/delete 对应的 canonical handler 方法，并在 `router.go` 的 `/api/ds` 组下注册它们；业务逻辑继续直接复用现有 `DatasourceService` 的 `Tree()`、`GetByID()`、`Save()`、`Update()`、`Delete()` 等方法。

**Why**

service 层已经具备完整能力，当前缺口主要在 transport 层的 canonical 暴露面。如果这一步重新包一层新的 datasource service abstraction，只会扩大 diff 和验证面，而不会带来真实收益。

**Alternative considered**

- 继续只通过 compatibility bridge 暴露能力：无法推进 canonical migration。
- 新建一套 canonical-only datasource service：会引入重复逻辑和额外一致性风险。

### 2. 保留 compatibility `/datasource/*` 路由，canonical 与 compatibility 并行存在

**Decision**

本次 change 中，旧的 `/datasource/*` 路由全部保留；前端只把 `datasource.ts` 中对应的 5 个核心调用切到 `/api/ds/*`，而 compatibility bridge 继续作为 fallback 和回滚路径存在。

**Why**

这是一次渐进式 canonical 化，而不是一次 compatibility 清理。保留旧路径可以降低上线风险，也让 smoke 失败时可以通过前端 URL 回滚迅速恢复。

**Alternative considered**

- 同步删除 compatibility 路由：风险过高，且不必要。
- 后端做 implicit redirect/rewrite：会把“canonical 是否真正被前端使用”这件事变得不可观察。

### 3. 前端切换只集中在 `apps/frontend/src/api/datasource.ts`

**Decision**

前端 canonical cutover 仅通过 `apps/frontend/src/api/datasource.ts` 完成，不直接改各个 datasource 页面。页面、store、组件继续依赖相同的 API wrapper 名称和相同的返回结构。

**Why**

这能把变更集中在单一边界层，降低 Vue 页面和业务组件的回归面。只要 canonical handler 保持 response shape 与 compatibility 等价，页面不需要感知路由切换。

**Alternative considered**

- 在页面中直接改请求路径：会制造多点 diff，增加遗漏和回滚成本。

### 4. 验证顺序采用“backend canonical handler → frontend API regression → datasource page smoke”

**Decision**

验证按照以下顺序执行：

1. 后端 canonical handler regression（确认 `/api/ds/*` 的 envelope、错误语义和 compatibility 一致）
2. 前端 `datasource.ts` API regression（确认 URL 切换但 wrapper contract 不变）
3. datasource 页面 smoke（tree load、detail read、save/update/delete 至少覆盖主干流中的可执行部分）

**Why**

这次 change 的核心风险不是业务算法，而是 transport contract 与前端切换边界。先锁 transport contract，再看页面 smoke，能更快定位问题来源。

**Alternative considered**

- 只做页面 smoke：定位粒度太粗，失败时不容易区分 handler 问题还是页面问题。

## Risks / Trade-offs

- **[Risk] canonical handler 与 compatibility handler 返回 envelope/字段存在微差异** → **Mitigation:** 为 `/api/ds/tree`、`/api/ds/:id`、save/update/delete 补 handler regression，并对前端 wrapper 做回归断言。
- **[Risk] 前端 `datasource.ts` 切 canonical 后，某些历史页面仍依赖 compatibility 特殊语义** → **Mitigation:** 仅迁移 5 个最核心调用，保留 `/datasource/*` 路径不删；页面 smoke 以 datasource 主页面为主。
- **[Risk] delete 语义在 canonical 路由上与现有 POST/GET fallback 不一致** → **Mitigation:** canonical 路由明确采用单一路径 `POST /api/ds/delete/:id`，compatibility fallback 继续保留但不扩散到新 canonical 设计中。
- **[Risk] 这次只迁 core CRUD，导致 datasource 模块 canonical/compat 混合态继续存在** → **Mitigation:** 这是有意的分阶段设计；本 change 先收最关键路径，为后续 `getTables/getSchema/previewData` 等扩展接口的 canonical 化建立模式。

## Migration Plan

1. 在 `DatasourceHandler` 中新增 canonical tree/get/save/update/delete 入口，并在 `/api/ds` 路由组注册。
2. 保持 compatibility `/datasource/*` 路由不变，确保相同 service 逻辑仍可被旧路径调用。
3. 切换 `apps/frontend/src/api/datasource.ts` 中对应 5 个调用到 `/api/ds/*`。
4. 补 backend handler regression 与 frontend datasource API regression。
5. 运行 backend/frontend 校验，并执行 datasource 页面 core CRUD smoke。

**Rollback**

- 若 canonical cutover 引发页面回归，优先回退前端 `datasource.ts` 中 5 个 URL 切换。
- 因为 compatibility routes 不移除，回退不需要恢复额外后端逻辑。

## Open Questions

- canonical detail read 是否只统一到 `GET /api/ds/:id`，还是需要同时提供 hide-password/simple 变体的 canonical 读路径；本 change 暂按统一 detail read 设计，具体 response shaping 需在 tasks 阶段再确认。
- datasource delete canonical 路由是否长期保留 `POST /api/ds/delete/:id`，还是后续进一步收敛为更 RESTful 的 delete 设计；本 change 先选择与现有前端/handler 最接近的 `POST` 路径以降低迁移风险。

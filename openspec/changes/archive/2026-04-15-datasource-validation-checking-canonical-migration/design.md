## Context

在 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync 和 file ingest 后，验证与检查类能力仍留在 `/datasource/*` compatibility 路径，包括 validate by ID、checkRepeat 和 checkApiDatasource。这使 datasource 模块处于"主体能力已 canonical、验证/检查仍 compatibility"的不连续状态，不利于后续判断 canonical 收口剩余范围，也让验证问题的定位路径与其他 datasource 路由不一致。

当前 change 的范围和约束：

- 只迁移 `validateById`（`GET /api/ds/validate/:id`）、`checkRepeat`（`POST /api/ds/checkRepeat`）和 `checkApiDatasource`（`POST /api/ds/checkApiDatasource`）三条路由；
- compatibility `/datasource/validate/:id`、`/datasource/checkRepeat`、`/datasource/checkApiDatasource` 必须保留，不在本次移除；
- 前端只切 `apps/frontend/src/api/datasource.ts` 的三条 wrapper URL 及 `ApiHttpRequestDraw.vue` 中 `cancelMap` key，不改调用方式与 contract；
- 不扩展到验证/检查之外的 datasource 能力，也不重写验证语义或引入新的检查工作流。

## Goals / Non-Goals

**Goals:**
- 新增 canonical datasource validation/checking 路由：
  - `GET /api/ds/validate/:id`
  - `POST /api/ds/checkRepeat`
  - `POST /api/ds/checkApiDatasource`
- 让前端 `datasource.ts` 中 `validateById`、`checkRepeat`、`checkApiItem` 切到 `/api/ds/*`，并保持 wrapper 名称、参数与返回结构不变。
- 更新 `ApiHttpRequestDraw.vue` 中 `cancelMap` key 从 `/datasource/checkApiDatasource` 到 `/api/ds/checkApiDatasource`。
- 保持 compatibility-safe response envelope 和显式失败语义。
- 为 canonical handler/router、frontend API boundary 补充 regression 验证。

**Non-Goals:**
- 不移除 `/datasource/validate/:id`、`/datasource/checkRepeat`、`/datasource/checkApiDatasource` compatibility 路由。
- 不重构验证逻辑、重复检查算法、API 数据源校验协议或数据源导入流程。
- 不把其它 datasource 扩展接口纳入本次 change。

## Decisions

### 1. 采用"canonical handler 增量暴露 + 既有 service 复用"

**Decision**

在 `apps/backend-go/internal/transport/http/handler/datasource_handler.go` 中补齐 `ValidateByID`、`CheckRepeat`、`CheckAPIDatasource` 的 canonical handler，并继续复用当前已存在的验证/检查业务能力（`service.ValidateByID`、`service.CheckRepeat`、`service.CheckAPIDatasource`），不新增平行 service。

**Why**

前面几轮 datasource canonical migration 已经证明，最小风险路径是把差异收敛在 transport 层。验证/检查这三条路由的核心需求是"canonical 暴露面补齐"，而不是新增业务语义，因此延续已有 backend 逻辑可以避免重复实现与 contract 偏移。

**Alternatives considered**

- 继续只保留 compatibility bridge：无法推进 canonical 收口。
- 单独新建 canonical-only validation service：会增加冗余实现和维护成本，但没有明确收益。

### 2. 前端只改 API 边界层 URL 和 cancelMap key，不改调用点与调用方式

**Decision**

只在 `apps/frontend/src/api/datasource.ts` 中将 `validateById`、`checkRepeat`、`checkApiItem` 切到 `/api/ds/*`，保留 wrapper 名称、请求体形状和调用方式不变。同步更新 `ApiHttpRequestDraw.vue` 中 `cancelMap` 的 key 从 `/datasource/checkApiDatasource` 到 `/api/ds/checkApiDatasource`。

**Why**

把 cutover 限定在 API boundary，能显著降低改动面，也让 rollback 非常直接：只需回退三条 URL 选择和 cancelMap key 即可恢复到 compatibility 路径。`cancelMap` key 必须同步更新，否则 cancel token 会因 key 不匹配而失效。

**Alternatives considered**

- 在页面组件中逐点替换路径：风险分散且回滚困难。
- 保留 cancelMap key 不变：cancel token 将无法匹配 canonical 路径，导致 API 请求取消机制失效。

### 3. 维持 compatibility-safe contract，显式保留验证失败语义

**Decision**

canonical 三条路由必须保持 compatibility 路由当前的 response envelope 与显式失败语义，尤其是无效 ID、重复数据源、API 数据源校验失败等情况，不允许静默降级为"空成功"。

**Why**

这次 change 的目标是 canonical cutover，不是业务 contract redesign。如果迁移时顺带弱化失败语义，会让前端和排障都更难判断问题发生在哪一层。

**Alternatives considered**

- 对失败统一返回空 payload 或 success envelope：会掩盖真实验证失败，和当前 explicit failure 目标冲突。

### 4. ValidateByID canonical 路由使用 GET 方法 + 路径参数

**Decision**

`GET /api/ds/validate/:id` 保持与 compatibility 路由 `GET /datasource/validate/:id` 一致的 HTTP 方法和路径参数风格。canonical handler 直接从 `c.Param("id")` 解析 ID 并调用 `service.ValidateByID(id)`。

**Why**

validate-by-ID 是幂等只读操作，语义上适合 GET。保持与 compatibility 路由完全一致的方法和参数风格，可以避免前端 cutover 时因方法或参数位置差异导致的隐蔽 bug。

**Alternatives considered**

- 改为 POST + JSON body：与 compatibility 不一致，增加 cutover 复杂度。

## Risks / Trade-offs

- **[Risk] ValidateByID 在 canonical handler 中解析路径参数的行为可能与 compatibility bridge 的内联解析存在细微差异** → **Mitigation:** regression 覆盖 canonical/compat 的成功与显式失败 envelope，特别关注无效 ID 和边界值。
- **[Risk] checkApiDatasource 在 ApiHttpRequestDraw.vue 中的 cancelMap key 替换不完整** → **Mitigation:** 全量搜索 `cancelMap` 中对 `/datasource/checkApiDatasource` 的所有引用，确保全部更新到 `/api/ds/checkApiDatasource`。
- **[Risk] 分阶段迁移期间其他 datasource 验证/检查相关扩展仍未 canonical 化** → **Mitigation:** 明确本次只做 `validateById`、`checkRepeat`、`checkApiDatasource`，避免范围蔓延。

## Migration Plan

1. backend 增加 `ValidateByID`、`CheckRepeat`、`CheckAPIDatasource` canonical handler，并注册 `/api/ds/*` 路由。
2. compatibility 同类路由保持原样可用。
3. frontend `datasource.ts` 切换三条 URL 到 `/api/ds/*`。
4. frontend `ApiHttpRequestDraw.vue` 更新 `cancelMap` key 到 canonical 路径。
5. 更新 backend/frontend 回归测试。
6. 执行 lint/tscheck/go test/build 验证。

**Rollback**

- 优先回退 `datasource.ts` 中三条 URL 选择和 `cancelMap` key，恢复到 `/datasource/*`；
- 因 compatibility routes 保留，不需要紧急回滚后端验证/检查逻辑。

## Open Questions

- `checkApiDatasource` 的请求体结构（当前为 `map[string]string`）是否在所有调用场景中完全一致；实现前需要优先核对前端调用链。
- `validateById` 的错误响应格式是否与 `POST /api/ds/validate` 保持完全一致的 envelope 结构。

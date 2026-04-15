## Context

在 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync、file ingest、validation/checking 和 tree/folder management 后，datasource 查询变体（hidePw、getSimpleDs）和永久删除（perDelete）仍留在 `/datasource/*` compatibility 路径。这让 datasource 模块处于"大部分能力已 canonical、查询变体和永久删除仍 compatibility"的不连续状态，不利于后续判断 canonical 收口剩余范围。

当前 change 的范围和约束：

- 后端迁移 `hidePw`（`GET /api/ds/hidePw/:id`）、`simple`（`GET /api/ds/simple/:id`）、`perDelete`（`POST /api/ds/perDelete/:id`）三条路由到 canonical handler；
- 前端修复 3 处 URL 差距：`datasource.ts` 中 `getHidePwById`、`getSimpleDs`、`perDelete`；
- compatibility `/datasource/hidePw/:id`、`/datasource/getSimpleDs/:id`、`/datasource/perDelete/:id` 必须保留，不在本次移除；
- 不扩展到查询变体和永久删除之外的 datasource 能力，也不重命名语义或引入新的工作流。

## Goals / Non-Goals

**Goals:**
- 新增 canonical datasource 查询变体和永久删除路由：
  - `GET /api/ds/hidePw/:id`
  - `GET /api/ds/simple/:id`
  - `POST /api/ds/perDelete/:id`
- 让前端 `datasource.ts` 中 `getHidePwById`、`getSimpleDs`、`perDelete` 切到 `/api/ds/*`，保持 wrapper 名称、参数与返回结构不变。
- 保持 compatibility-safe response envelope 和显式失败语义。
- 为 canonical handler/router、frontend API boundary 补充 regression 验证。

**Non-Goals:**
- 不移除 `/datasource/hidePw/:id`、`/datasource/getSimpleDs/:id`、`/datasource/perDelete/:id` compatibility 路由。
- 不重构密码隐藏、简化查询或永久删除的业务逻辑或校验规则。
- 不把其它 datasource 扩展接口纳入本次 change。

## Decisions

### 1. 采用"canonical handler 增量暴露 + 既有 service 复用"

**Decision**

在 `apps/backend-go/internal/transport/http/handler/datasource_handler.go` 中补齐 `HidePw`、`Simple`、`PerDelete` 的 canonical handler，并继续复用当前已存在的业务能力（`service.GetByID` 用于 hidePw 和 simple 的响应变换，`service.PerDelete` 用于永久删除），不新增平行 service。

**Why**

前面几轮 datasource canonical migration 已经证明，最小风险路径是把差异收敛在 transport 层。这三条路由的核心需求是"canonical 暴露面补齐"，而不是新增业务语义，因此延续已有 backend 逻辑可以避免重复实现与 contract 偏移。

**Alternatives considered**

- 继续只保留 compatibility bridge：无法推进 canonical 收口。
- 单独新建 canonical-only service：会增加冗余实现和维护成本，但没有明确收益。

### 2. 前端只改 API 边界层 URL，不改调用点与调用方式

**Decision**

只在 `apps/frontend/src/api/datasource.ts` 中将 `getHidePwById`、`getSimpleDs`、`perDelete` 切到 `/api/ds/*`，保留所有 wrapper 名称、请求体形状和调用方式不变。

**Why**

把 cutover 限定在 API boundary，能显著降低改动面，也让 rollback 非常直接：只需回退 URL 选择即可恢复到 compatibility 路径。hidePw 和 getSimpleDs 涉及前端数据源编辑交互，perDelete 涉及删除确认流程，如果改动超出 URL 层，验证面会急剧扩大。

**Alternatives considered**

- 在页面组件中逐点替换路径：风险分散且回滚困难。
- 顺手调整 wrapper 封装：会把 transport 迁移和业务逻辑改造混在一起，扩大验证面。

### 3. 维持 compatibility-safe contract，显式保留失败语义

**Decision**

canonical 三条路由必须保持 compatibility 路由当前的 response envelope 与显式失败语义，尤其是数据源不存在、ID 无效、永久删除前置条件不满足等情况，不允许静默降级为"空成功"。

**Why**

这次 change 的目标是 canonical cutover，不是业务 contract redesign。如果迁移时顺带弱化失败语义，会让前端和排障都更难判断问题发生在哪一层。

**Alternatives considered**

- 对失败统一返回空 payload 或 success envelope：会掩盖真实操作失败，和当前 explicit failure 目标冲突。

## Risks / Trade-offs

- **[Risk] hidePw 和 simple 的响应变换（密码字段脱敏、简化结构提取）在 canonical handler 中的实现可能与 compatibility bridge 存在细微差异** → **Mitigation:** regression 覆盖 canonical/compat 的成功与显式失败 envelope，特别关注数据源不存在和 ID 格式异常场景。
- **[Risk] perDelete 操作涉及级联清理，canonical handler 中可能遗漏 compatibility bridge 的清理步骤** → **Mitigation:** 复用 `service.PerDelete` 确保清理逻辑一致，通过回归测试验证删除后数据完整性。
- **[Risk] 前端 `getHidePwById` 和 `getSimpleDs` 的调用方可能对 URL 路径有硬编码依赖** → **Mitigation:** 保持前端 wrapper contract 不变，路径变更对调用方透明。

## Migration Plan

1. backend 增加 `HidePw`、`Simple`、`PerDelete` canonical handler，并注册 `/api/ds/*` 路由。
2. compatibility 同类路由保持原样可用。
3. frontend `datasource.ts` 切换 `getHidePwById`、`getSimpleDs`、`perDelete` URL 到 `/api/ds/*`。
4. 更新 backend/frontend 回归测试。
5. 执行 lint/tscheck/go test/build 验证。

**Rollback**

- 优先回退 `datasource.ts` 中 URL 选择，恢复到 `/datasource/*`；
- 因 compatibility routes 保留，不需要紧急回滚后端逻辑。

## Open Questions

- `simple` 路径是否需要在 canonical 路由中使用 `simple` 还是 `getSimpleDs` 命名（当前选择 `simple` 以匹配 canonical 命名风格）。
- 前端 `getHidePwById`、`getSimpleDs`、`perDelete` 是否有其他调用方也在使用旧路径，需要全量搜索确认。

## Context

在 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync、file ingest 和 validation/checking 后，datasource tree/folder 管理能力（move、rename、createFolder）仍留在 `/datasource/*` compatibility 路径。同时，前端 `dataset.ts` 中 `tree` 和 `getTables` 以及 `datasource.ts` 中 `validate`（POST）虽已有对应 canonical 路由，却仍在使用旧路径。这让 datasource 模块处于"大部分能力已 canonical、tree/folder 管理和若干前端路径仍 compatibility"的不连续状态，不利于后续判断 canonical 收口剩余范围。

当前 change 的范围和约束：

- 后端迁移 `move`（`POST /api/ds/move`）、`reName`（`POST /api/ds/reName`）、`createFolder`（`POST /api/ds/createFolder`）三条路由到 canonical handler；
- 前端修复 6 处 URL 差距：`datasource.ts` 中 move、reName、createFolder、validate（POST），`dataset.ts` 中 tree、getTables；
- compatibility `/datasource/move`、`/datasource/reName`、`/datasource/createFolder` 必须保留，不在本次移除；
- 不扩展到 tree/folder 管理之外的 datasource 能力，也不重命名语义或引入新的 folder 工作流。

## Goals / Non-Goals

**Goals:**
- 新增 canonical datasource tree/folder 管理路由：
  - `POST /api/ds/move`
  - `POST /api/ds/reName`
  - `POST /api/ds/createFolder`
- 让前端 `datasource.ts` 中 `move`、`reName`、`createFolder` 切到 `/api/ds/*`，保持 wrapper 名称、参数与返回结构不变。
- 让前端 `dataset.ts` 中 `tree`、`getTables` 切到 `/ds/tree` 和 `/ds/tables`（canonical 已存在）。
- 让前端 `datasource.ts` 中 `validate`（POST）切到 `/ds/validate`（canonical 已存在）。
- 保持 compatibility-safe response envelope 和显式失败语义。
- 为 canonical handler/router、frontend API boundary 补充 regression 验证。

**Non-Goals:**
- 不移除 `/datasource/move`、`/datasource/reName`、`/datasource/createFolder` compatibility 路由。
- 不重构移动、重命名、创建文件夹的业务逻辑或校验规则。
- 不把其它 datasource 扩展接口纳入本次 change。

## Decisions

### 1. 采用"canonical handler 增量暴露 + 既有 service 复用"

**Decision**

在 `apps/backend-go/internal/transport/http/handler/datasource_handler.go` 中补齐 `Move`、`Rename`、`CreateFolder` 的 canonical handler，并继续复用当前已存在的 tree/folder 管理业务能力（`service.Move`、`service.Rename`、`service.CreateFolder`），不新增平行 service。

**Why**

前面几轮 datasource canonical migration 已经证明，最小风险路径是把差异收敛在 transport 层。tree/folder 管理这三条路由的核心需求是"canonical 暴露面补齐"，而不是新增业务语义，因此延续已有 backend 逻辑可以避免重复实现与 contract 偏移。

**Alternatives considered**

- 继续只保留 compatibility bridge：无法推进 canonical 收口。
- 单独新建 canonical-only tree/folder service：会增加冗余实现和维护成本，但没有明确收益。

### 2. 前端只改 API 边界层 URL，不改调用点与调用方式

**Decision**

只在 `apps/frontend/src/api/datasource.ts` 中将 `move`、`reName`、`createFolder`、`validate`（POST）切到 `/api/ds/*` 或 `/ds/*`，在 `apps/frontend/src/api/dataset.ts` 中将 `tree`、`getTables` 切到 `/ds/tree` 和 `/ds/tables`，保留所有 wrapper 名称、请求体形状和调用方式不变。

**Why**

把 cutover 限定在 API boundary，能显著降低改动面，也让 rollback 非常直接：只需回退 URL 选择即可恢复到 compatibility 路径。tree/folder 管理涉及前端拖拽交互和树形组件，如果改动超出 URL 层，验证面会急剧扩大。

**Alternatives considered**

- 在页面组件中逐点替换路径：风险分散且回滚困难。
- 顺手调整 wrapper 封装：会把 transport 迁移和业务逻辑改造混在一起，扩大验证面。

### 3. 维持 compatibility-safe contract，显式保留 tree/folder 管理失败语义

**Decision**

canonical 三条路由必须保持 compatibility 路由当前的 response envelope 与显式失败语义，尤其是移动目标无效、重命名冲突、文件夹已存在等情况，不允许静默降级为"空成功"。

**Why**

这次 change 的目标是 canonical cutover，不是业务 contract redesign。如果迁移时顺带弱化失败语义，会让前端和排障都更难判断问题发生在哪一层。

**Alternatives considered**

- 对失败统一返回空 payload 或 success envelope：会掩盖真实操作失败，和当前 explicit failure 目标冲突。

### 4. 前端已存在 canonical 路由的遗漏路径统一补齐

**Decision**

除了新增的 tree/folder 管理路由迁移外，本 change 同时补齐 3 处前端已暴露 canonical 路由但仍在使用旧路径的差距：`dataset.ts` 中 `tree`（`/datasource/tree` → `/ds/tree`）、`getTables`（`/datasource/getTables` → `/ds/tables`）以及 `datasource.ts` 中 `validate`（POST）（`/datasource/validate` → `/ds/validate`）。

**Why**

这三处差距的性质与本 change 完全一致（前端 URL 遗漏），且后端 canonical 路由已存在，补齐成本极低。如果单独拆出 change，会增加管理开销且拖延 canonical surface 的完整收敛。

**Alternatives considered**

- 将前端遗漏拆成独立 change：增加协调成本，但可以缩小单个 change 的 review 范围。

## Risks / Trade-offs

- **[Risk] move 操作在 canonical handler 中的参数解析可能与 compatibility bridge 存在细微差异** → **Mitigation:** regression 覆盖 canonical/compat 的成功与显式失败 envelope，特别关注移动目标不存在和跨层级移动场景。
- **[Risk] reName 操作与前端树组件的交互依赖可能因路径变更产生缓存不一致** → **Mitigation:** 保持前端 wrapper contract 不变，路径变更不影响树组件的刷新逻辑。
- **[Risk] dataset.ts 中 tree/getTables 切换可能影响 dataset 模块的 datasource 选择交互** → **Mitigation:** 后端 canonical 路由已存在且经过验证，前端 wrapper 名称和 contract 不变，风险极低。
- **[Risk] 分阶段迁移期间其他 datasource tree/folder 相关扩展仍未 canonical 化** → **Mitigation:** 明确本次只做 `move`、`reName`、`createFolder`，避免范围蔓延。

## Migration Plan

1. backend 增加 `Move`、`Rename`、`CreateFolder` canonical handler，并注册 `/api/ds/*` 路由。
2. compatibility 同类路由保持原样可用。
3. frontend `datasource.ts` 切换 move、reName、createFolder、validate（POST）URL 到 `/api/ds/*` 或 `/ds/*`。
4. frontend `dataset.ts` 切换 tree、getTables URL 到 `/ds/tree` 和 `/ds/tables`。
5. 更新 backend/frontend 回归测试。
6. 执行 lint/tscheck/go test/build 验证。

**Rollback**

- 优先回退 `datasource.ts` 和 `dataset.ts` 中 URL 选择，恢复到 `/datasource/*`；
- 因 compatibility routes 保留，不需要紧急回滚后端 tree/folder 管理逻辑。

## Open Questions

- `reName` 路径中 camelCase 命名风格是否需要在 canonical 路由中保留（保持与 compatibility 路由一致）。
- `dataset.ts` 中 `tree` 和 `getTables` 是否有其他调用方也在使用旧路径，需要全量搜索确认。

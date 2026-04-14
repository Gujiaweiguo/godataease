## Context

在 datasource canonical migration 已完成 core CRUD 与 table exploration 后，`previewData`、`syncApiTable`、`syncApiDs` 仍通过 `/datasource/*` compatibility 路径承载。当前模块因此继续处于“canonical + compatibility 混用”状态：部分关键读写链路已切到 `/api/ds/*`，但 preview/sync 这组三条强相关能力仍留在 compatibility bridge。

这会带来两个问题：

1. datasource canonical 收口边界不连续，后续迁移难以判断“还剩哪一层”；
2. preview/sync 与其他 canonical 路由在回归验证和故障定位上的路径不一致，增大线上排障成本。

本 change 约束清晰：

- 仅迁移 `previewData`、`syncApiTable`、`syncApiDs` 三条路由；
- 保留 `/datasource/previewData`、`/datasource/syncApiTable`、`/datasource/syncApiDs` compatibility routes；
- 前端只调整 `apps/frontend/src/api/datasource.ts` 的路由选择，不改 wrapper 名称与 contract；
- 不扩展到 upload、remote file、其他 datasource 扩展能力。

## Goals / Non-Goals

**Goals:**
- 提供 canonical 路由：
  - `POST /api/ds/previewData`
  - `POST /api/ds/syncApiTable`
  - `POST /api/ds/syncApiDs`
- 前端 `datasource.ts` 的这三条调用切到 `/api/ds/*`，保持调用方无感知。
- 保持 compatibility-safe response envelope 与显式失败语义。
- 通过 backend regression + frontend regression + 页面 smoke 验证迁移安全。

**Non-Goals:**
- 不移除对应 `/datasource/*` compatibility 路由。
- 不改 preview/sync 业务语义与 payload 结构。
- 不在本 change 内处理 upload/remote file canonical 化。

## Decisions

### 1) 采用“canonical handler 增量暴露 + 既有 service 复用”

**Decision**

在 `datasource_handler.go` 补齐 preview/sync 三条 canonical handler，继续复用现有 service 能力，不新增平行 service。

**Why**

这与前两刀 datasource canonical migration 的技术路径一致，能把风险集中在 transport 边界，避免重复业务实现和语义漂移。

**Alternatives considered**

- 继续仅用 compatibility bridge：无法推进 canonical 收口。
- 新建 canonical-only service：增加冗余，收益低于复杂度。

### 2) 前端仅改 API 边界层 URL，不改调用点

**Decision**

只在 `apps/frontend/src/api/datasource.ts` 将 `previewData`、`syncApiTable`、`syncApiDs` 路由改为 `/api/ds/*`，页面和组件调用链保持不变。

**Why**

最小 diff、最小回滚成本；若出现问题，仅回退这三条 URL 选择即可恢复 compatibility。

**Alternatives considered**

- 在页面内逐点替换 API 路径：改动点分散、回滚困难。

### 3) 明确“显式失败语义保持”而非静默兜底

**Decision**

canonical preview/sync 路由必须保持 compatibility 路径当前可测试的失败语义（如 invalid datasource、invalid sync target、backend unavailable），禁止“空成功”降级。

**Why**

此次目标是路由 canonical 化，不是业务语义重写；迁移期必须确保行为可观测、可回归。

**Alternatives considered**

- 将异常统一成空数组或成功占位：会隐藏真实失败，增加排障成本。

### 4) 验证顺序采用 backend → frontend → smoke

**Decision**

先锁 backend contract，再锁 frontend URL 切换，最后做页面级 smoke（preview/sync 可执行路径）。

**Why**

定位粒度更小，失败时能快速判断是 handler contract、wrapper 还是页面链路问题。

## Risks / Trade-offs

- **[Risk] preview/sync canonical 与 compatibility 在 envelope 细节上出现偏差** → **Mitigation:** 补充 canonical/alias regression 覆盖成功与显式失败场景。
- **[Risk] preview/sync 页面行为依赖历史隐式字段** → **Mitigation:** 前端保持 wrapper contract 不变，并在 smoke 中覆盖关键路径。
- **[Risk] 分阶段迁移期间仍有部分 datasource 能力未 canonical 化** → **Mitigation:** 明确边界，保证本次仅处理三条路由，避免范围蔓延。

## Migration Plan

1. backend 增加 `previewData`、`syncApiTable`、`syncApiDs` canonical handler 并注册 `/api/ds/*`。
2. compatibility 同类路由保持原样可用。
3. frontend `datasource.ts` 切换三条 URL 到 `/api/ds/*`。
4. 更新 backend/frontend 回归测试。
5. 执行 lint/tscheck/go test/build 与页面 smoke。

**Rollback**

- 优先回退 `datasource.ts` 三条 URL 选择，恢复到 `/datasource/*`；
- 因 compatibility routes 保留，不需要紧急回滚后端业务逻辑。

## Open Questions

- `syncApiTable` / `syncApiDs` 在不同 datasource 类型下是否有历史特化字段尚未被回归测试覆盖；实现前需补齐样例。
- `previewData` 在权限失败与数据源不可达场景下的错误文案是否已稳定；若未稳定，回归应以语义而非字面文案断言。

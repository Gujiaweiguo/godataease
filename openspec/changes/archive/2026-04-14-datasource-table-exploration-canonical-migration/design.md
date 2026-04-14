## Context

在 `datasource-core-crud-canonical-migration` 合入后，datasource 页面最核心的 tree、detail、save、update、delete 已经开始通过 `/api/ds/*` canonical 路由承载，但 table exploration 相关能力仍全部停留在 `/datasource/*` compatibility 路径，包括表列表、表状态、字段查看和 schema 获取。

这使 datasource 模块仍处于明显的混合态：页面主干写入和基础读取已经切到 canonical，而表探索这种高频读取链路仍依赖 compatibility bridge。对 datasource 页面来说，这 4 条路由既强关联，又常一起触发；如果继续长期保留在 compatibility 路径，会让后续 `previewData` 等扩展能力的迁移边界变得模糊，也不利于逐步收敛 `/datasource/*` 的职责。

当前 change 的约束也很明确：

- compatibility `/datasource/*` 路由必须保留，不能在同一 change 中移除，以保证渐进迁移和快速回滚能力；
- 本次只覆盖 `tables`、`tableStatus`、`tableField`、`schema` 4 条 table exploration 路由，不扩展到 `previewData`、`syncApi*`、upload、remote file 等更大范围能力；
- 前端需要保持现有 wrapper 名称和返回结构不变，让页面和组件无需感知 canonical cutover；
- 错误语义必须保持显式，不能因为迁到 canonical 路由而把 invalid datasource、invalid table 或缺少状态证据等情况静默降级为“空成功”。

## Goals / Non-Goals

**Goals:**
- 为 datasource table exploration 补齐 canonical `/api/ds/*` 路由：`POST /api/ds/tables`、`POST /api/ds/tableStatus`、`POST /api/ds/tableField`、`POST /api/ds/schema`。
- 让前端 `apps/frontend/src/api/datasource.ts` 中与上述 4 条路由对应的调用切换到 canonical 路由，同时保持 wrapper 名称、入参形状和响应结构不变。
- 继续复用现有 `DatasourceService` 与已存在的 datasource 领域能力，不引入平行 service 或重复业务逻辑。
- 保持 compatibility `/datasource/*` 同类路由继续可用，使回滚仍可通过前端 URL 切换完成。
- 为 canonical handler、前端 API 边界和 datasource 页面 table exploration 主链路补充 regression 与 smoke 证据。

**Non-Goals:**
- 不 canonical 化 `previewData`、`syncApiTable`、`syncApiDs`、upload、remote file 等 datasource 扩展能力。
- 不移除、重写或收缩现有 `/datasource/getTables`、`/datasource/getTableStatus`、`/datasource/getTableField`、`/datasource/getSchema` compatibility routes。
- 不调整 datasource exploration 的响应 envelope、鉴权模型、页面业务流程或组件交互方式。
- 不在本次 change 中追求更 RESTful 的方法或 URI 设计；canonical 路由继续采用与现有前端调用最接近的 `POST` 模式以降低迁移风险。

## Decisions

### 1. 采用“handler 扩面 + service 复用”模式扩展 table exploration canonical 路由

**Decision**

在 `apps/backend-go/internal/transport/http/handler/datasource_handler.go` 中补齐 `tables`、`tableStatus`、`tableField`、`schema` 对应的 canonical handler 方法，并在 `router.go` 的 `/api/ds` 组下注册它们；业务逻辑继续复用现有 `DatasourceService` 和底层 repository/domain 能力。

**Why**

这与上一刀 datasource core CRUD canonical migration 的模式保持一致：当前缺口主要在 transport 层的 canonical 暴露面，而不是 service 层缺少能力。继续沿用既有 service，可以把 diff 控制在 handler/router/API boundary，降低重复实现和 contract 偏移风险。

**Alternatives considered**

- 继续只依赖 compatibility bridge：无法推进 canonical migration，也无法为后续 datasource 扩展能力建立清晰模式。
- 新建一套 canonical-only datasource service：会制造重复逻辑与额外一致性成本，没有实际收益。

### 2. canonical 路由与 compatibility 路由并行存在，前端只切换边界层 URL

**Decision**

本次 change 中保留现有 compatibility 路由，同时新增以下 canonical 路由映射：

- `/datasource/getTables` → `POST /api/ds/tables`
- `/datasource/getTableStatus` → `POST /api/ds/tableStatus`
- `/datasource/getTableField` → `POST /api/ds/tableField`
- `/datasource/getSchema` → `POST /api/ds/schema`

前端只在 `apps/frontend/src/api/datasource.ts` 中切换这 4 个 URL，页面、store 和组件继续依赖原有 wrapper，不直接改调用点。

**Why**

把 cutover 控制在边界层可以最小化前端 diff，也让 rollback 足够直接：如果 smoke 暴露问题，只需回退 `datasource.ts` 中这 4 个 URL 即可恢复到 compatibility 路径，而不需要额外恢复后端逻辑。

**Alternatives considered**

- 同步删除 compatibility 路由：风险过高，不符合渐进迁移策略。
- 在页面内部逐点修改路径：会制造多点 diff，放大遗漏和回滚成本。

### 3. 保持 compatibility-safe contract，不借 canonical cutover 改变错误或成功语义

**Decision**

canonical `tables`、`tableStatus`、`tableField`、`schema` 必须维持与现有 compatibility 路由一致的 envelope 和业务语义，尤其要保留以下行为边界：

- table listing 仍返回 datasource exploration 页面当前依赖的列表结构；
- table status 在无状态证据时继续保持显式 unknown / non-success 语义，而不是静默返回空成功；
- table field 在 datasource 或 table 非法时继续提供可测试的显式失败语义；
- schema discovery 继续满足 datasource 编辑页 schema 选择器的现有结构预期。

**Why**

这次 change 的目标是路由 canonical 化，而不是 contract redesign。如果趁机修改 envelope 或页面依赖的 payload 结构，会把 transport 切换与业务行为调整混在一起，显著扩大验证面。

**Alternatives considered**

- 借 canonical 路由统一重塑 exploration 响应结构：长期可能有价值，但不适合与本次迁移绑定。
- 对异常场景做隐式兜底为空结果：会削弱错误可观察性，也与 specs 中“显式且可测试”的要求冲突。

### 4. 验证顺序采用“backend transport contract → frontend API regression → datasource 页面 smoke”

**Decision**

验证按以下顺序进行：

1. 后端 canonical handler regression，确认 `/api/ds/*` exploration 路由的 envelope 和错误语义与 compatibility 一致；
2. 前端 `datasource.ts` API regression，确认仅 URL 切换到 canonical，但 wrapper contract 不变；
3. datasource 页面 table exploration smoke，至少覆盖表列表加载、schema 获取、字段查看和状态检查等主链路中的可执行部分。

**Why**

这次 change 的主要风险集中在 transport contract 与前端边界切换。先锁定 handler 行为，再验证前端 wrapper，最后做页面 smoke，可以把问题定位粒度控制在最小范围。

**Alternatives considered**

- 只做页面 smoke：一旦失败，很难快速区分是 handler contract 变化、API wrapper 变化还是页面本身问题。

## Risks / Trade-offs

- **[Risk] canonical handler 与 compatibility handler 在 envelope 或字段细节上存在微差异** → **Mitigation:** 为 4 条 canonical exploration 路由补 handler regression，覆盖成功、非法输入和显式失败场景。
- **[Risk] 前端切到 canonical 后，仍有未纳入本次范围的 datasource 调用保持 compatibility，模块短期内继续混合态** → **Mitigation:** 这是有意的分阶段设计；通过在 tests 和最终说明中明确边界，避免误判为“已完全 canonical 化”。
- **[Risk] `tableStatus`、`tableField` 等接口的失败语义与 core CRUD 不同，容易在迁移时被意外弱化** → **Mitigation:** 在 specs 和 regression 中明确要求 invalid datasource / invalid table / missing status evidence 的显式行为。
- **[Risk] datasource 页面 smoke 受本地环境、数据源类型和连接状态影响，导致验证波动** → **Mitigation:** 先用后端/前端 regression 锁定 contract，再把 smoke 聚焦在本地可稳定复现的 exploration 主链路，并记录环境限制。

## Migration Plan

1. 在 `DatasourceHandler` 中新增 `tables`、`tableStatus`、`tableField`、`schema` 对应 canonical 入口，并在 `/api/ds` 路由组注册。
2. 保持 compatibility `/datasource/*` 同类路由不变，确保旧调用仍可工作。
3. 切换 `apps/frontend/src/api/datasource.ts` 中对应 4 个调用到 `/api/ds/*`。
4. 补 backend handler regression 与 frontend datasource API regression。
5. 执行 backend/frontend 校验，并进行 datasource 页面 table exploration smoke。

**Rollback**

- 若 canonical cutover 引发 datasource 页面 exploration 回归，优先回退 `apps/frontend/src/api/datasource.ts` 中这 4 个 URL 切换。
- 因为 compatibility routes 在本次 change 中继续保留，回退不需要恢复额外后端逻辑，也不需要回滚 schema 或 service 层结构。

## Open Questions

- `tableField` canonical handler 是否完全复用 compatibility 请求体形状，还是存在某些 datasource 类型依赖的兼容字段需要额外显式测试；实现前需要通过现有调用点和测试确认。
- `schema` 返回结构是否对不同 datasource 类型始终一致，还是有历史页面仅消费部分字段；实现时需要优先验证 datasource 编辑页的 schema selection flow。
- `tableStatus` 当前 compatibility 侧“无状态证据”与“明确失败”的区分边界是否已被前端测试固定；若没有，需要在 regression 中把这部分语义先钉住。

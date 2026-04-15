## Context

在 datasource canonical migration 已完成 core CRUD、table exploration、preview/sync 后，文件接入相关能力仍留在 `/datasource/*` compatibility 路径，包括文件上传和远程文件加载。这样会让 datasource 模块继续处于“主体能力已 canonical、文件接入仍 compatibility”的不连续状态，不利于后续判断 canonical 收口剩余范围，也让 ingest 问题的定位路径与其他 datasource 路由不一致。

当前 change 的范围和约束都比较明确：

- 只迁移 `uploadFile` 与 `loadRemoteFile` 两条 file ingest 路由；
- compatibility `/datasource/uploadFile`、`/datasource/loadRemoteFile` 必须保留，不在本次移除；
- 前端只切 `apps/frontend/src/api/datasource.ts` 的两条 wrapper URL，不改调用方式与 contract；
- 不扩展到 upload/remote file 之外的 datasource 能力，也不重写上传语义或引入新的 ingest 工作流。

## Goals / Non-Goals

**Goals:**
- 新增 canonical datasource file ingest 路由：
  - `POST /api/ds/uploadFile`
  - `POST /api/ds/loadRemoteFile`
- 让前端 `datasource.ts` 中 `uploadFile` 与 `loadRemoteFile` 切到 `/api/ds/*`，并保持 wrapper 名称、参数与返回结构不变。
- 保持 compatibility-safe response envelope 和显式失败语义。
- 为 canonical handler/router、frontend API boundary 与 ingest 主链路补充 regression 与 smoke 证据。

**Non-Goals:**
- 不移除 `/datasource/uploadFile` 与 `/datasource/loadRemoteFile` compatibility 路由。
- 不重构上传协议、远程文件解析逻辑、文件格式支持矩阵或数据源导入流程。
- 不把其它 datasource 扩展接口纳入本次 change。

## Decisions

### 1. 采用“canonical handler 增量暴露 + 既有 service 复用”

**Decision**

在 `apps/backend-go/internal/transport/http/handler/datasource_handler.go` 中补齐 `uploadFile` 与 `loadRemoteFile` 的 canonical handler，并继续复用当前已存在的 ingest 业务能力，不新增平行 service。

**Why**

前面几刀 datasource canonical migration 已经证明，最小风险路径是把差异收敛在 transport 层。文件接入这两条路由的核心需求是“canonical 暴露面补齐”，而不是新增业务语义，因此延续已有 backend 逻辑可以避免重复实现与 contract 偏移。

**Alternatives considered**

- 继续只保留 compatibility bridge：无法推进 canonical 收口。
- 单独新建 canonical-only ingest service：会增加冗余实现和维护成本，但没有明确收益。

### 2. 前端只改 API 边界层 URL，不改调用点与上传调用方式

**Decision**

只在 `apps/frontend/src/api/datasource.ts` 中将 `uploadFile` 与 `loadRemoteFile` 切到 `/api/ds/*`，保留 wrapper 名称、multipart header 处理、请求体形状和调用方式不变。

**Why**

file ingest 往往牵涉页面交互、上传组件和 multipart 细节；把 cutover 限定在 API boundary，能显著降低改动面，也让 rollback 非常直接：只需回退两条 URL 选择即可恢复到 compatibility 路由。

**Alternatives considered**

- 在页面组件中逐点替换路径：风险分散且回滚困难。
- 顺手调整 upload 封装：会把 transport 迁移和上传语义改造混在一起，扩大验证面。

### 3. 维持 compatibility-safe contract，显式保留 ingest 失败语义

**Decision**

canonical `uploadFile` 与 `loadRemoteFile` 必须保持 compatibility 路由当前的 response envelope 与显式失败语义，尤其是非法上传输入、远程地址无效、后端不可用等情况，不允许静默降级为“空成功”。

**Why**

这次 change 的目标是 canonical cutover，不是业务 contract redesign。文件接入本身就比普通 JSON 接口更脆弱，如果迁移时顺带弱化失败语义，会让前端和排障都更难判断问题发生在哪一层。

**Alternatives considered**

- 对失败统一返回空 payload 或 success envelope：会掩盖真实 ingest 失败，和当前 explicit failure 目标冲突。

### 4. 验证顺序采用 backend transport → frontend API regression → ingest smoke

**Decision**

验证顺序保持一致：先确认 canonical handler 的 envelope/语义，再确认 frontend wrapper URL 切换正确，最后做 ingest 可执行 smoke。

**Why**

file ingest 问题可能来自 multipart 绑定、远程地址校验、wrapper 边界或页面本身。按 transport → API → smoke 的顺序，可以把定位范围压到最小。

## Risks / Trade-offs

- **[Risk] multipart upload 在 canonical handler 上与 compatibility route 的绑定行为存在细微差异** → **Mitigation:** regression 覆盖 canonical/compat 的成功与显式失败 envelope，特别关注 multipart headers 和参数缺失场景。
- **[Risk] remote file ingest 可能依赖历史隐式字段或 URL 校验细节** → **Mitigation:** 保持 frontend wrapper contract 不变，并在 smoke 中覆盖远程加载主链路或其显式失败语义。
- **[Risk] 分阶段迁移期间其他 datasource ingest 相关扩展仍未 canonical 化** → **Mitigation:** 明确本次只做 `uploadFile` 与 `loadRemoteFile`，避免范围蔓延。

## Migration Plan

1. backend 增加 `uploadFile` 与 `loadRemoteFile` canonical handler，并注册 `/api/ds/*` 路由。
2. compatibility 同类路由保持原样可用。
3. frontend `datasource.ts` 切换两条 URL 到 `/api/ds/*`。
4. 更新 backend/frontend 回归测试。
5. 执行 lint/tscheck/go test/build 与 ingest smoke。

**Rollback**

- 优先回退 `datasource.ts` 中两条 URL 选择，恢复到 `/datasource/*`；
- 因 compatibility routes 保留，不需要紧急回滚后端 ingest 逻辑。

## Open Questions

- `uploadFile` 的 multipart 字段名、header 细节是否在所有现有调用场景中完全一致；实现前需要优先核对现有测试与页面调用。
- `loadRemoteFile` 对远程地址校验失败的错误语义是否已稳定；若文案历史上波动，应优先以语义而非完整文案做断言。

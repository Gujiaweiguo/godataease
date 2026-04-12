## Context

stage3 已经补齐了 dataset/datasource 兼容链路里一批“缺实现”的显式能力，包括 barInfo 审计字段、字段保存、函数列表和按数据源批量字段查询。但当前 SQL preview 仍保留着明显的 stage0/stage1 迁移形态：`DatasetService.PreviewSQL` 只做 SQL 校验，然后直接调用 `DatasetRepository.PreviewSQL` 在当前 Go 后端绑定的本地数据库连接上执行查询。`SQLPreviewRequest.datasourceId` 已经存在于请求模型和前端调用里，但服务端不会用它来选择执行目标，这会让“前端选择数据源预览 SQL”和“后端永远在本地同步库预览 SQL”之间形成隐式错位。

与此同时，`/datasetField/multFieldValuesForPermissions` 和 `/datasetField/copilotFields` 这两个剩余兼容接口还没有形成可执行且可治理的语义。它们与字段枚举、权限过滤、计算字段编辑和潜在 AI 辅助能力有关，如果直接当作普通 endpoint 补齐，后面很容易把权限语义、外部依赖和 preview 路由问题混在一起。

stage4 的目标不是马上承诺“所有 SQL preview 都能跨外部数据源直连执行”，而是先把这件事拆成可以治理的设计决策：

- 当前本地同步库 preview 语义是什么，哪些调用是稳定可保留的；
- 外部数据源直连 preview 如果要支持，需要哪些最小基础设施；
- 如果暂不支持外部直连，前端和兼容调用面应该收到什么显式语义，而不是继续默默忽略 `datasourceId`；
- 剩余字段接口如何沿用已有字段枚举/权限模型，而不是另起一套旁路逻辑。

约束包括：

- 必须保留现有 compatibility route 和 Java 兼容调用面，不能通过删除参数或直接删路由来“解决”语义不清问题。
- 必须继续沿用当前 Go 的 domain / repository / service / handler 分层，而不是在 stage4 借机做大规模模块重构。
- 必须承认当前系统的主架构是“外部数据源同步到本地库后消费”，所以外部直连 preview 如果被引入，必须被定义成受限增强，而不是把整个 dataset runtime 改成跨源实时执行框架。

## Goals / Non-Goals

**Goals:**
- 为 `POST /datasetData/previewSql` 明确定义执行目标选择语义，消除当前 `datasourceId` 被隐式忽略的灰色行为。
- 定义外部数据源直连 preview 的准入条件、连接抽象、失败语义和权限边界，或者显式定义“不支持直连时”的稳定降级语义。
- 为 `/datasetField/multFieldValuesForPermissions` 和 `/datasetField/copilotFields` 建立与现有字段枚举、字段权限和 dataset 元数据一致的兼容语义。
- 为后续 specs 和 tasks 提供一条清晰的“先定边界、再补 capability”的路线，避免 stage4 退化为零散接口打补丁。

**Non-Goals:**
- 不在 stage4 直接引入完整的跨源查询引擎，也不把整个 dataset runtime 改造成按请求实时直连外部数据源。
- 不在 stage4 重写 datasource 驱动管理体系；如果需要外部 preview 连接能力，只定义满足 preview 场景的最小抽象。
- 不在 stage4 扩展新的前端页面或重构 dataset/datasource 管理 UI，只在现有调用面上明确语义并做最小接线。
- 不在 stage4 解决所有 AI/copilot 产品能力，只收敛 `copilotFields` 这一兼容入口的字段选择与返回契约。

## Decisions

### 1. Preview 语义先分成“本地同步库预览”和“外部直连预览”两类能力

**Decision:** stage4 不再把 `previewSql` 当作单一路径能力，而是显式拆成两类语义：

- **本地同步库预览**：沿用当前 `DatasetRepository.PreviewSQL`，在本地库执行已经同步落地的数据；
- **外部直连预览**：仅在满足明确条件时启用，作为受限增强路径而不是默认路径。

**Why:** 当前实现已经稳定服务于本地同步库语义，但请求模型中的 `datasourceId` 又暗示了另一个尚未兑现的产品期望。直接继续用一个“看起来支持数据源、实际上永远只查本地”的接口，会让前后端都难以判断失败是权限问题、架构限制还是实现缺口。先把能力拆开，规格和错误语义才有办法稳定。

**Alternatives considered:**
- **继续保持单一路径并忽略 `datasourceId`**：实现成本最低，但会把当前灰色语义固化成长期技术债。
- **立即把所有 preview 切到外部直连**：与当前同步型架构冲突过大，也会把权限、稳定性和驱动适配风险一次性放大。

### 2. 外部直连 preview 采用“最小连接执行抽象”，不把 repository 直接绑到所有驱动

**Decision:** 如果 stage4 选择支持外部直连 preview，应在 service 层之下增加一个最小执行抽象（例如 preview executor / datasource query executor），由它根据 datasource 类型选择连接与执行策略；`DatasetRepository` 继续负责本地库和 dataset 元数据，不直接扩展成“所有外部数据源的统一 repository”。

**Why:** 当前 repository 层天然绑定 GORM 和本地数据库，直接把外部数据库驱动、连接串解析和 SQL 执行塞进 `DatasetRepository`，会模糊“本地元数据访问”和“外部执行”的职责边界。preview executor 可以把外部执行限制在一个更小的接口里，也更容易做超时、连接清理和白名单控制。

**Alternatives considered:**
- **在 `DatasetRepository` 内部按 datasource type 分支执行**：改动看起来小，但会让 repository 迅速变成跨驱动杂糅层。
- **新增完整 cross-source query service**：长期可能合理，但对 stage4 来说过重。

### 3. 外部直连 preview 不是默认开启能力，而是受 datasource 类型、权限和安全校验共同约束

**Decision:** 即使支持外部直连 preview，也只允许在满足以下条件时进入外部执行路径：

- datasource 类型在明确支持名单内；
- 请求方具备该 datasource 的可见/可用权限；
- SQL 通过现有 preview SQL 校验与必要的附加安全约束；
- 连接参数可被当前 Go 后端安全解析和建立；
- 执行超时、返回行数和结果大小落在 preview 限制之内。

不满足条件时，接口必须返回显式非成功语义，说明是“不支持该数据源直连预览”“无权限”“连接失败”或“SQL 不允许执行”，而不是退回本地库静默执行。

**Why:** preview 看起来是只读能力，但一旦直接连外部数据源，它就会变成真正的跨系统执行面。相比同步库 preview，风险集中在连接泄漏、慢查询、错误 SQL 对外部实例的压力、以及授权边界模糊。stage4 需要先把这些约束变成契约。

**Alternatives considered:**
- **任何传了 `datasourceId` 的请求都尝试外部执行**：简单但高风险。
- **失败时自动回退到本地同步库 preview**：会让结果来源不可解释，尤其在数据不同步时会误导用户。

### 4. 若 stage4 不支持某类外部直连 preview，必须返回显式限制语义，不再忽略 `datasourceId`

**Decision:** 对于当前 Go 后端尚不支持的外部 datasource 类型或连接方式，`previewSql` 应返回稳定、可测试的 non-success 语义，明确该请求不被支持或需要先走同步路径；不允许继续把 `datasourceId` 当作无效参数默默忽略。

**Why:** stage4 的一个核心目标就是结束灰色语义。即使本次设计最终决定“只支持少量 datasource 类型，其他都不支持”，这也比现在的隐式忽略更好，因为前端、测试和运维都能知道系统到底承诺了什么。

**Alternatives considered:**
- **维持现状直到未来完全支持所有 datasource**：会让语义长期悬空。
- **删除 `datasourceId` 参数**：会破坏兼容调用面，也掩盖真实产品需求。

### 5. 剩余字段接口统一建立在现有字段元数据与权限过滤逻辑之上

**Decision:** `/datasetField/multFieldValuesForPermissions` 和 `/datasetField/copilotFields` 不应各自创造独立的数据解析逻辑，而应复用现有 `GetFieldEnum`、`GetFieldEnumDs`、字段列表和权限过滤相关能力：

- `multFieldValuesForPermissions` 侧重“在权限上下文下返回可用枚举值”；
- `copilotFields` 侧重“返回适合 copilot/辅助表达式能力消费的字段集合”，但它的数据来源仍应是已有 dataset field 元数据。

**Why:** 这两个接口都位于字段消费面，不是新的主域。如果各自再长出一套字段解析和权限判断，后续字段行为会在 preview、filter、copilot 三套逻辑中分叉。

**Alternatives considered:**
- **每个接口单独实现自己的字段查询/过滤逻辑**：开发快，但长期最难维护。
- **先只做静态占位响应**：无法完成 stage4 的“剩余兼容能力收口”目标。

### 6. Stage4 先定义契约和最小实现路径，避免承诺全量 datasource 支持矩阵

**Decision:** 设计上只要求建立“支持 / 不支持 / 降级”的明确矩阵，不在 stage4 design 中承诺所有 datasource 类型都进入外部直连 preview 支持名单。具体支持矩阵可以从当前 Go 后端最容易安全落地的 datasource 类型开始。

**Why:** 如果 design 一开始就承诺“所有 datasource 都能外部直连 preview”，后续 specs 和任务会被迫覆盖大量驱动与连接边角。stage4 更合理的目标是把行为边界讲清楚，再让实现按受控范围扩展。

## Risks / Trade-offs

- **[Risk] 外部直连 preview 一旦接入，会把后端暴露成跨数据源执行面** → 通过支持名单、权限校验、超时/行数限制和显式不支持语义，把能力限制在 preview 级别。
- **[Risk] preview 结果来源从“总是本地库”变成“本地或外部”后，前端和排障复杂度上升** → 在响应语义、日志和 specs 中明确结果来源与失败分类，避免静默回退。
- **[Risk] 最小执行抽象仍可能逐步膨胀成跨源查询框架** → 将 executor 限定为 preview-only 能力，并在 non-goals 中明确不扩展到整体 dataset runtime。
- **[Risk] `multFieldValuesForPermissions` 和 `copilotFields` 如果定义过于宽泛，会反过来拖累 stage4 范围** → 只收敛兼容接口契约，不在本次设计中引入新的 AI 业务流程或复杂前端交互改造。
- **[Risk] 对不支持的 datasource 返回显式失败后，短期内会暴露前端已有误用** → 通过 specs 和回归验证尽早暴露依赖方假设，优于继续维持隐式错误行为。

## Migration Plan

1. 在 specs 中分别定义 `dataset-preview-routing` 新能力，以及 `dataset-management` / `datasource-management` 的 delta requirement。
2. 先收敛 `previewSql` 的契约：明确哪些请求走本地同步库、哪些请求允许外部直连、哪些请求必须显式失败。
3. 如需支持外部直连 preview，先在后端实现最小 preview executor 与 datasource 类型支持矩阵，再将 compatibility route 接到该能力。
4. 在扩展 datasource 侧更大范围迁移之前，先切换前端中已有 canonical Go 支撑的三个 dataset wrapper：tree、preview、table-field。
5. 验证该 canonical 切换不会改变现有前端归一化与后处理行为，重点覆盖 dataset tree 节点归一化，以及 preview / table-field 的字段名处理逻辑。
6. 在字段侧复用现有字段枚举与权限过滤逻辑，补齐 `multFieldValuesForPermissions` 和 `copilotFields` 的 handler / service 语义。
7. 补齐 backend unit/integration、frontend 受影响调用点回归和兼容语义验证。
8. 若 stage4 引发回归，优先关闭外部直连 preview 接线并保留 stage3 本地同步库 preview 基线，同时让不支持语义继续显式返回。

## Open Questions

- stage4 是否要在设计上直接选择“支持少量 datasource 类型直连 preview”，还是先只定义契约和显式不支持语义，把真正的外部执行放到后续子阶段。
- 现有 datasource 模型中哪些连接信息已经足够支撑最小 preview executor，哪些还依赖旧 Java 行为或额外驱动资产。
- `multFieldValuesForPermissions` 的权限过滤应复用当前行权限 / 列权限哪一层语义，还是需要引入更明确的“可枚举值可见性”规则。
- `copilotFields` 返回的字段集合是否只面向当前 dataset group，还是允许跨 datasource / cross-dataset 聚合视图。
- preview 响应是否需要显式暴露“结果来源”（local-sync / datasource-direct），以支持排障和前端提示。

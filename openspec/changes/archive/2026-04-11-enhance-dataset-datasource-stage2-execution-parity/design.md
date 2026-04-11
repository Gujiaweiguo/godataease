## Context

stage1 已完成数据集导出任务创建、字段删除主链路、数据源表状态真实化，以及数据源删除写接口的第一阶段补齐，但这些能力仍带有明显的“迁移稳定化”特征：

- 数据集导出只保证了 compatibility route 不再返回占位响应，还没有形成“创建任务 → 查询状态 → 获取文件/失败信息”的闭环。
- 数据集字段删除已经具备真实服务路径，但删除阻断仍主要停留在基础存在性与 chart-scoped 约束，尚未覆盖更细粒度的图表、计算字段和下游配置依赖。
- 数据源表状态已经从固定成功转为基于同步记录聚合，但状态集合仍偏粗粒度，前端可消费语义和更新时间来源还不够稳定。

stage2 的目标不是扩展新的 BI 域能力，而是把 stage1 已接通的链路提升为“用户可持续消费、失败可诊断、行为可治理”的执行闭环。约束包括：

- 必须保留现有 compatibility route 和 Java 兼容调用面，避免在 stage2 直接制造 breaking cutover。
- 必须继续沿用当前 Go 的 domain / repository / service / handler 分层，而不是借机重构模块边界。
- 必须优先复用现有 export-center、datasource task log、dataset metadata 和前端已有管理页，而非引入新的任务中心或新页面。

## Goals / Non-Goals

**Goals:**
- 为数据集导出定义完整的任务闭环语义，包括任务创建后的状态查询、结果下载、失败可诊断性，以及历史兼容 ID 的可预测收敛。
- 为数据集字段删除引入更细粒度的依赖分析和阻断语义，避免误删正在被图表、计算字段或下游配置消费的字段。
- 为数据源表状态输出建立更稳定的执行态模型，明确状态映射、更新时间来源和前端展示约束。
- 为上述链路补齐前后端自动化回归与手工验证标准，确保 stage2 输出的是用户可消费闭环，而不只是服务端实现增强。

**Non-Goals:**
- 不在本次 change 中新增新的数据源类型、驱动插件能力或跨源执行框架。
- 不在本次 change 中重构 export-center 整体模型，只做 dataset 导出接线与语义收敛。
- 不在本次 change 中重建完整 Java 任务编排域，只增强当前可用状态模型与消费约束。
- 不在本次 change 中改变统一权限模型或扩大 datasource/dataset 主域边界。

## Decisions

### 1. 数据集导出沿用 export-center 作为 canonical task surface

**Decision:** stage2 不新增独立的数据集导出任务模型，而是继续把 `POST /datasetTree/exportDataset` 接到现有 export-center，补齐任务查询、下载和失败信息消费链路。

**Why:** stage1 已经创建了 export task，但如果 stage2 再引入另一套 dataset-export 专用任务面，前端会出现“同类导出两个任务中心”的裂缝。继续沿用 export-center，才能把 dataset export 视为现有导出体系中的一种来源。

**Alternatives considered:**
- **新增 dataset-export 专用表和接口**：语义清晰，但会复制现有 export-center 能力并增加迁移成本。
- **继续只返回 taskId，不定义后续消费语义**：实现简单，但无法完成 stage2 目标。

### 2. 历史兼容 ID 收敛维持“兼容桥接”而不是污染 canonical identity

**Decision:** 对历史/样例数据中的 dataset 兼容 ID 继续在服务层做桥接与收敛，但不把这种近似回退行为写入 canonical 数据身份模型。

**Why:** 当前兼容 ID 问题本质上是迁移与样例数据遗留，不应变成核心 identity 规则。如果把“最近邻回退”渗透到主身份模型，会使后续权限、导出和依赖判断变得不可预测。

**Alternatives considered:**
- **直接修改历史数据主键/外键**：根治性强，但需要额外的数据迁移方案，不适合作为 stage2 的默认路径。
- **完全移除兼容回退**：会让当前已有兼容调用面再次失效。

### 3. 字段删除阻断以“依赖分类 + 明确错误语义”实现，而不是静默跳过

**Decision:** 在 dataset service 中增加依赖扫描结果，按图表直接引用、计算字段引用、下游配置引用等类别阻断删除，并返回结构化、可解释的失败原因。

**Why:** stage1 的删除链路解决了“能删”，stage2 要解决“什么时候不能删，以及为什么不能删”。静默跳过或 generic error 都会降低前端可解释性，也不利于回归测试。

**Alternatives considered:**
- **发现依赖后继续删除并记录 warning**：风险高，容易破坏图表与配置一致性。
- **一律禁止 chart-scoped 字段删除**：过于保守，会让功能失去实际可用性。

### 4. 数据源表状态采用有限稳定状态集，而不是暴露底层任务系统原始状态

**Decision:** stage2 继续对底层 task log / sync 记录做聚合，但升级为更稳定的有限状态集合，并为每个状态定义前端展示语义和更新时间来源。

**Why:** 前端需要稳定可消费的业务状态，而不是依赖底层任务系统的原始枚举。直接暴露原始状态会把底层实现细节泄漏到页面和测试中。

**Alternatives considered:**
- **直接返回底层任务原始状态**：实现快，但耦合过深。
- **继续只用 stage1 的粗粒度 Completed / UnderExecution / Error / Warning**：信息量不足，不利于后续治理。

### 5. 前端先以兼容增强为主，不做页面结构重构

**Decision:** stage2 前端只增强现有 dataset/datasource 页面和 API 层，不引入新的管理页或复杂 store 重构。

**Why:** 本次 change 的价值在于执行闭环和语义对齐，而不是 UI 重做。保持页面结构稳定可以让回归范围更可控。

## Risks / Trade-offs

- **[Risk] export-center 现有模型未完全覆盖 dataset-specific 下载语义** → 通过在现有任务语义上最小增量扩展字段与查询行为，避免另起体系。
- **[Risk] 字段依赖扫描过于保守导致可删字段被误拦** → 依赖分类和错误信息必须可诊断，并优先对真正破坏性引用进行阻断。
- **[Risk] 兼容 ID 回退继续存在会让行为显得不直观** → 将其限定为 compatibility bridge / service 级策略，并在设计和 specs 中明确其边界。
- **[Risk] 状态语义增强后前端展示需要同步调整** → 在 specs 中同步定义消费约束，并以受影响前端测试覆盖状态映射。

## Migration Plan

1. 在 specs 中新增/修改关于 dataset export 闭环、field dependency 阻断、datasource status 细粒度语义的 requirement。
2. 先补后端 canonical service/repository 语义，再让 compatibility route 复用增强后的实现。
3. 对前端 API 与现有页面做最小接线更新，确保任务闭环和状态展示契约一致。
4. 执行 backend unit/integration、frontend lint/ts/test，以及手工回归验证。
5. 若 stage2 语义增强引发回归，优先关闭新阻断规则或更细粒度状态映射，回退到 stage1 行为基线。

## Open Questions

- export-center 当前是否已经有可复用的“任务创建后按来源下载”查询路径，还是需要新增 dataset-specific 查询/下载桥接。
- 字段依赖阻断是否需要区分“硬阻断”和“可确认后继续”的两级策略，还是 stage2 统一做硬阻断。
- datasource status 的 stage2 有限状态集最终应包含哪些稳定枚举，前端是否需要新增 tooltip / 详情说明来解释状态来源。
- 历史兼容 ID 的最终治理是继续保留桥接，还是在后续单独 change 中做数据迁移与身份统一。

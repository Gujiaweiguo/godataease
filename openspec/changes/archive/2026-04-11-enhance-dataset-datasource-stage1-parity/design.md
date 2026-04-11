## Context

当前 Go 端的数据集/数据源模块已经覆盖树查询、字段查询、预览、校验、基础治理和大部分 compatibility bridge 路由，但仍有几处明显的阶段性缺口：`datasetTree/exportDataset`、`datasetField/delete/{id}`、`datasetField/deleteByChartId/{id}` 仍是占位或弱实现；`datasource/getTableStatus` 仍返回固定成功；数据源删除主要依赖历史兼容删除路径，缺少更规范的写接口承载未来收敛。

这些缺口直接影响前端现有页面的真实闭环能力。前端数据集页面已经具备导出入口、字段管理和数据源详情/状态展示，继续保留后端桩逻辑会让 UI 呈现出“可点击但不可验证”的假闭环。与此同时，本次 change 只聚焦第一阶段补齐，不引入跨数据源执行、驱动扩展、Calcite 深化执行链路等第二阶段议题。

约束包括：
- 需要保持 Java 兼容路径、`/api` 路径、`/de2api` 路径在成功/失败语义上的一致性。
- 需要遵循现有 Go 架构：domain → repository → service → handler/compatibility bridge。
- 需要尽量复用已有导出中心、同步记录、软删除和权限治理能力，而不是重新发明新流程。

## Goals / Non-Goals

**Goals:**
- 为数据集导出兼容接口提供真实、可测试的服务实现，并复用现有导出/任务能力。
- 为数据集字段删除提供正式服务路径，覆盖单字段删除和按图表关联批量删除。
- 为数据源表状态查询提供基于真实同步记录的状态聚合，而不是固定成功返回。
- 为数据源删除新增规范写接口，并保证与历史兼容删除路径共享同一业务逻辑与错误语义。
- 为前端 API 调用层和兼容别名补齐回归验证，确保页面行为不再依赖桩响应。

**Non-Goals:**
- 不在本 change 中实现跨数据源 SQL 执行或更完整的 Calcite 执行链路。
- 不在本 change 中引入新数据源类型、驱动管理或插件型扩展能力。
- 不在本 change 中重构 repository/service 为接口抽象，仅在现有结构上补齐功能。
- 不在本 change 中重做导出中心或同步任务框架，只在现有能力上接线。

## Decisions

### 1. 数据集导出沿用现有导出中心语义，而不是新增专用导出通道

**Decision:** `datasetTree/exportDataset` 通过 dataset service 调用已有导出能力，返回与前端现有导出中心兼容的任务/结果语义，而不是在 dataset 模块里新增一套独立下载流程。

**Why:** 前端 `ExportExcel.vue` 已经围绕导出中心任务状态、重试、下载和清理构建完整交互。若后端单独提供“同步下载式”的新语义，会导致前端闭环再次分叉。

**Alternatives considered:**
- **直接返回文件流**：实现简单，但会绕开现有导出中心任务体系，不利于大数据量和失败重试。
- **继续维持 not supported**：最小改动，但无法解决第一阶段闭环缺口。

### 2. 数据集字段删除使用正式 service/repository 路径，并显式处理图表依赖

**Decision:** 在 `dataset_service` 与 `dataset_repo` 中补齐字段删除能力，分别支持按字段 ID 删除和按图表 ID 批量删除；删除前执行依赖与归属校验，删除后返回确定性结果。

**Why:** 现有 compatibility route 只是前端契约占位，真正缺的是后端可复用业务逻辑。把逻辑落到 service/repository 层，才能保证 canonical route 和兼容别名复用同一实现。

**Alternatives considered:**
- **只在 compatibility bridge 内联删除 SQL**：实现快，但会继续放大 handler 内联业务逻辑问题。
- **前端隐藏删除入口**：回避问题，不符合本次补齐目标。

### 3. 数据源表状态按“最近同步记录聚合”输出，而不是尝试一次性补全完整任务域模型

**Decision:** `datasource/getTableStatus` 第一阶段基于现有同步记录/任务日志能力聚合出稳定状态，至少区分成功、失败、运行中、未知/未同步，并返回可用于前端展示的更新时间或最近记录时间。

**Why:** 当前最大的错误是固定返回 Success。第一阶段的目标是让状态“真实且可解释”，不是一步到位重建 Java 全量任务编排模型。

**Alternatives considered:**
- **继续固定成功**：会持续误导前端和测试。
- **完整重建 Java Quartz/任务域**：价值高，但超出本次范围。

### 4. 数据源删除新增规范写接口，但历史兼容删除路径继续保留并共用实现

**Decision:** 新增规范写接口（优先 `POST`，遵循现有兼容习惯），并让历史 `GET /datasource/delete/:id` 与新接口复用同一 service 删除逻辑、同一前置校验和同一错误语义。

**Why:** 当前兼容删除路径不适合作为长期收敛目标，但现有前端和历史调用方仍依赖它。新增规范接口并复用实现，可以在不破坏兼容性的前提下为后续迁移提供着陆点。

**Alternatives considered:**
- **直接移除 GET 删除接口**：有 breaking 风险。
- **只保留 GET，不新增规范接口**：无法改善语义与后续治理空间。

### 5. 规格变更仅落在 dataset-management 与 datasource-management 两个既有 capability

**Decision:** 这次只修改现有两个 capability 的 requirement/scenario，不新增独立 capability。

**Why:** 本次不是新的业务域，而是对现有域的阶段性补齐。保持 capability 收敛，有利于后续 specs 和实现边界稳定。

## Risks / Trade-offs

- **[Risk] 导出中心接线与前端预期不完全一致** → 先以现有前端 API 和任务中心交互为准补齐契约，并补回归验证。
- **[Risk] 字段删除可能破坏图表或数据集引用关系** → 删除前显式检查图表关联范围，批量删除按 chart 归属限定，避免误删非目标字段。
- **[Risk] 表状态聚合来源不足，无法完全复刻 Java 细粒度状态** → 第一阶段只承诺稳定的有限状态集合与最近时间语义，把更细粒度任务态留到后续 change。
- **[Risk] 新旧删除接口并存会增加测试矩阵** → 通过共享 service 逻辑和统一错误语义，把额外复杂度收敛在 handler/router 层。
- **[Risk] compatibility bridge 继续承担较多业务编排** → 本次仅把新逻辑下沉到 service/repository，避免继续在 bridge 中叠加内联业务。

## Migration Plan

1. 先修改 specs，明确 dataset export、field delete、table status、delete API consistency 的 requirement。
2. 在 Go 后端补齐 service/repository 能力，并让 compatibility bridge 路由切到正式实现。
3. 新增规范化数据源删除写接口，前端 API 层优先切换到新接口，同时保留旧接口兼容。
4. 补充后端单元测试、集成测试，以及前端受影响 API/页面回归验证。
5. 若验证发现回归，优先回退新接线点，保留兼容路径和旧前端调用，直到问题定位完成。

## Open Questions

- 数据集导出在 Go 端应直接复用现有 export center 哪个服务入口，是否已存在可复用的 dataset 任务模型，还是需要最小增量适配。
- `datasetField/delete` 的依赖校验边界应只覆盖 chart calc fields，还是覆盖所有引用该字段的下游配置。
- `datasource/getTableStatus` 当前可用的真实状态来源是同步任务日志、表元信息，还是两者组合；若来源缺失，未知状态如何向前端暴露。
- 规范化数据源删除接口最终使用 `POST` 还是 `DELETE`；需要结合当前前端 axios 封装和 compatibility path 风格做最终选择。

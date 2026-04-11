## Why

stage1 已经补齐了数据集导出、字段删除、数据源表状态和删除路由的第一层可用性，但当前能力仍停留在“主链路可达”的层级：数据集导出尚未形成完整的任务查询/下载闭环，字段删除仍缺少更细粒度的依赖阻断语义，数据源表状态也只提供了第一阶段聚合态。现在进入 stage2，可以把这些能力从“可接通”推进到“可稳定消费、可解释、可治理”的执行闭环。

## What Changes

- 完善数据集导出执行闭环，明确任务创建后的查询、下载、失败诊断与历史兼容 ID 收敛行为。
- 强化数据集字段删除前的依赖分析，覆盖图表、计算字段和受影响下游配置的阻断语义与错误反馈。
- 提升数据源表状态的执行语义精度，明确更多稳定任务态、更新时间来源和前端展示约束。
- 补充上述执行链路的前后端回归验证与手工验收标准，确保 stage2 不只是接口增强，而是用户可消费的闭环能力。

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dataset-management`: 扩展数据集导出与字段删除 requirement，补齐导出任务闭环、兼容 ID 收敛、字段依赖阻断与失败诊断语义。
- `datasource-management`: 扩展数据源表状态 requirement，定义更细粒度且稳定的同步/执行状态输出与前端消费约束。

## Impact

- **Affected backend modules**:
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/service/export_service.go`
  - `apps/backend-go/internal/service/datasource_service.go`
  - `apps/backend-go/internal/repository/dataset_repo.go`
  - `apps/backend-go/internal/repository/export_repo.go`
  - `apps/backend-go/internal/repository/datasource_repo.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
- **Affected frontend modules**:
  - `apps/frontend/src/api/dataset.ts`
  - `apps/frontend/src/api/datasource.ts`
  - `apps/frontend/src/views/visualized/data/dataset/`
  - `apps/frontend/src/views/visualized/data/datasource/`
- **Affected APIs**:
  - `POST /datasetTree/exportDataset`
  - 导出任务查询/下载相关 export-center 路径
  - `POST /datasetField/delete/{id}`
  - `POST /datasetField/deleteByChartId/{id}`
  - `POST /datasource/getTableStatus`
- **Breaking changes**:
  - 无计划中的立即 breaking change，但 stage2 可能会把部分当前宽松成功语义收紧为显式失败或阻断语义。
- **Rollback strategy**:
  - 若 stage2 执行语义导致回归，可先保留 stage1 的状态聚合和导出创建行为，关闭更细粒度依赖阻断或下载闭环新接线点，回退到已验证的 stage1 基线。

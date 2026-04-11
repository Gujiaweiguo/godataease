## Why

当前数据集/数据源主流程已经具备浏览、预览、校验和基础治理能力，但第一阶段迁移仍存在几个影响前后端闭环的明显缺口：部分兼容接口仍是占位实现，数据源表状态仍返回固定成功，删除语义也存在历史兼容路径与规范写接口不一致的问题。现在补齐这批能力，可以把前端已有页面和 Go 后端现有迁移基线真正连成可验证、可回归的完整链路，并为后续更深的数据源执行与同步能力演进清掉低层阻塞。

## What Changes

- 将数据集导出兼容接口从占位响应提升为可执行能力，补齐前端数据集导出闭环。
- 将数据集字段删除兼容接口从占位逻辑提升为正式行为，覆盖按字段删除和按图表关联批量删除两类场景。
- 将数据源表状态查询从固定成功桩实现改为基于真实同步记录或可观测状态的确定性返回，避免前端误判同步健康状态。
- 为数据源删除补充规范写接口，并定义与历史兼容删除路径的一致性策略，逐步收敛到可审计、可测试的删除语义。
- 为上述能力补充前后端回归验证，确保 canonical 路径、compatibility alias、权限语义和失败语义保持一致。

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dataset-management`: 扩展数据集管理规范，覆盖导出能力、字段删除能力，以及这些写操作在兼容接口下的确定性成功/失败语义。
- `datasource-management`: 扩展数据源管理规范，覆盖表状态查询的真实状态返回，以及删除接口从历史兼容路径向规范写接口收敛时的兼容与一致性要求。

## Impact

- **Affected backend modules**:
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/service/datasource_service.go`
  - `apps/backend-go/internal/repository/dataset_repo.go`
  - `apps/backend-go/internal/repository/datasource_repo.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - `apps/backend-go/internal/transport/http/handler/dataset_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
- **Affected frontend modules**:
  - `apps/frontend/src/api/dataset.ts`
  - `apps/frontend/src/api/datasource.ts`
  - `apps/frontend/src/views/visualized/data/dataset/`
  - `apps/frontend/src/views/visualized/data/datasource/`
- **Affected APIs**:
  - `POST /datasetTree/exportDataset`
  - `POST /datasetField/delete/{id}`
  - `POST /datasetField/deleteByChartId/{id}`
  - `POST /datasource/getTableStatus`
  - 新增规范化的数据源删除写接口，同时保留历史兼容删除路径用于平滑迁移
- **Breaking changes**:
  - 无立即 breaking change；历史删除路径将保留兼容，但新实现会要求各别名路径在行为与错误语义上保持一致。
- **Rollback strategy**:
  - 若新能力导致回归，可回退到兼容接口的原实现，并保留规范写接口但关闭前端切换；表状态查询可退回到保守的非成功判定而非继续报告固定成功。

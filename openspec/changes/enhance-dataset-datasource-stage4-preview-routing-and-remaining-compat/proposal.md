## Why

stage3 已经补齐了数据集 barInfo、字段保存和辅助字段查询能力，但 dataset/datasource 兼容链路里仍有一类关键能力没有收口：SQL preview 仍只在本地同步库语义下工作，无法明确支持外部数据源直连预览；同时剩余的权限枚举值与 copilot 字段接口也尚未形成稳定兼容语义。现在进入 stage4，需要先把这些“剩余可见能力”从零散缺口整理成一条明确的架构与兼容收敛变更，避免后续实现时把同步架构、权限语义和前端契约混在一起。

## What Changes

- 明确 dataset SQL preview 是否支持“外部数据源直连执行”，并为支持路径定义连接管理、驱动抽象、超时控制、权限边界和失败语义。
- 若不直接支持外部直连预览，则定义受限替代方案和前端可消费的显式非成功语义，避免当前 `datasourceId` 形参继续处于“前端传了、后端忽略”的灰色状态。
- 补齐 `/datasetField/multFieldValuesForPermissions` 的兼容语义，明确其与现有字段枚举查询、行权限/列权限过滤之间的关系。
- 补齐 `/datasetField/copilotFields` 的兼容语义，明确字段筛选范围、返回结构、鉴权与降级行为。
- 为上述能力补充设计、规格和回归验证要求，确保 stage4 是一条“能力与语义收口”变更，而不是零散接口补丁。
- 作为 stage4 中一项受控的前端对齐步骤，将 dataset 树、预览、表字段加载这三类已有 Go canonical 支撑的调用，从 compatibility 路径切换到 `/dataset/tree`、`/dataset/preview`、`/dataset/fields`。
- 该切换仅用于降低 dataset 基础读路径对 compatibility wrapper 的依赖，不扩展到 datasource 全量迁移或更大范围的路由清理。

## Capabilities

### New Capabilities

- `dataset-preview-routing`: 定义 SQL preview 在本地同步库与外部数据源之间的路由、执行与失败语义。

### Modified Capabilities

- `dataset-management`: 扩展 preview 与字段枚举相关 requirement，补齐外部数据源 preview、权限枚举值和 copilot 字段接口语义。
- `datasource-management`: 扩展 datasource 与 dataset preview 之间的依赖边界、权限约束和失败诊断语义。

## Impact

- **Affected backend modules**:
  - `apps/backend-go/internal/service/dataset_service.go`
  - `apps/backend-go/internal/service/datasource_service.go`
  - `apps/backend-go/internal/repository/dataset_repo.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
  - 可能新增外部数据源 SQL 执行或连接抽象相关模块
- **Affected frontend modules**:
  - `apps/frontend/src/api/dataset.ts`
  - `apps/frontend/src/views/visualized/data/dataset/`
  - 可能涉及计算字段、SQL preview、权限筛选相关调用点
- **Affected APIs**:
  - `POST /datasetData/previewSql`
  - `POST /datasetField/multFieldValuesForPermissions`
  - `POST /datasetField/copilotFields`
- **Breaking changes**:
  - 无计划中的立即 breaking change，但 stage4 可能会把当前隐式忽略或宽松成功语义收紧为显式失败或受限支持语义。
- **Rollback strategy**:
  - 若外部 preview 路由或剩余兼容接口补齐引发回归，可先回退到当前 stage3 基线：保留本地同步库 preview 语义，并关闭新增外部路由或剩余兼容入口的接线点。

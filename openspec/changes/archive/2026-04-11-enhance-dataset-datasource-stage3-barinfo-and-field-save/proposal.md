## Why

数据集 barInfo API 返回全零数据（CreateTime=0, Creator=""），用户在前端看不到数据集的创建者、创建时间等审计信息。同时 `/datasetField/save` 端点未实现，导致图表编辑器的计算字段功能完全不可用。这两个问题直接影响用户的核心交互体验。

## What Changes

- **barInfo 审计字段完善**：给 `CoreDatasetGroup` 模型添加 `create_by`, `create_time`, `update_by`, `last_update_time` 审计列，更新 repository 查询，barInfo handler 填充真实数据，添加用户 ID → 用户名解析
- **Datasource creator 解析**：`sanitizeDatasourceResponse` 添加 `creator` 字段（解析 `create_by` → 用户名）
- **字段保存能力**：实现 `/datasetField/save` 端点，包含字段创建和更新逻辑，注册路由到 compatibility bridge
- **补充辅助端点**：实现 `/datasetField/getFunction`（计算字段编辑器的函数列表）和 `/datasetField/listByDsIds`（批量字段查询）

## Capabilities

### New Capabilities

（无新能力，均为已有规范的补充实现）

### Modified Capabilities

- `dataset-management`: 新增 barInfo 审计字段完整返回要求（requirement 15）、字段保存端点要求（requirement 16）
- `datasource-management`: 新增 datasource 详情返回 creator 解析用户名要求（requirement 16）

## Impact

- **后端**：修改 `domain/dataset/dataset.go`（CoreDatasetGroup 模型）、`repository/dataset_repo.go`（查询）、`service/dataset_service.go`（barInfo + SaveField）、`handler/compatibility_bridge_handler.go`（路由注册）
- **后端**：修改 `handler/datasource_handler.go` 或 `service/datasource_service.go`（creator 解析）
- **前端**：无变更（前端已定义好接口，只是后端返回数据不对）
- **API 兼容性**：纯增量变更，不破坏现有 API 契约
- **回滚策略**：所有变更为新增/补全字段，回滚后仅回到当前空白状态，无数据丢失风险

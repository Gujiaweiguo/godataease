## Context

数据集/数据源 Stage1（导出、字段删除、表状态）和 Stage2（导出生命周期、依赖阻塞、状态精度）已归档。本次 Stage3 聚焦两个用户直接可感知的缺陷：

1. **barInfo 审计信息缺失**：前端数据集详情面板显示空白的创建者和创建时间
2. **计算字段不可保存**：图表编辑器的 `/datasetField/save` 端点未实现

当前架构约束：
- Go 后端使用 Gin + GORM，domain/repository/service 三层架构
- 所有 SQL 查询运行在本地 MySQL（同步架构，非实时路由）
- 前端已定义好所有接口契约，只是后端返回数据不完整
- Compatibility bridge handler 承载所有 Java 时代路由映射

## Goals / Non-Goals

**Goals:**
- barInfo API 返回真实的审计数据（创建者、创建时间、更新者、更新时间）
- Datasource 详情返回解析后的 creator 用户名
- 实现 `/datasetField/save` 端点，支持字段创建和更新
- 实现 `/datasetField/getFunction` 端点，返回计算字段可用的 SQL 函数列表
- 实现 `/datasetField/listByDsIds` 端点，支持按多个数据源 ID 批量查询字段

**Non-Goals:**
- SQL 预览路由到外部数据源（需要连接池/驱动抽象，架构级变更）
- `/datasetField/copilotFields`（AI copilot 集成，独立功能）
- `/datasetField/multFieldValuesForPermissions`（权限过滤的枚举值，较低优先级）
- 新增数据源类型或驱动插件能力

## Decisions

### Decision 1: 扩展 CoreDatasetGroup 手动模型 vs 使用 auto-gen 模型

**选择**: 扩展手动模型，添加 `CreateBy`, `CreateTime`, `UpdateBy`, `LastUpdateTime` 四个审计字段。

**理由**: 
- 手动模型修改范围小，只影响 dataset 相关代码
- auto-gen 模型字段命名风格不同，替换会影响整个 service/repo 层
- 符合 Stage1/Stage2 的渐进式修改策略

**替代方案**: 切换到 auto-gen 模型 — 影响范围大，所有 dataset service/repo 调用都要改

### Decision 2: 用户名解析方式

**选择**: 在 service 层注入 UserRepository，通过 `FindUserByID` 解析 user ID → user name。

**理由**:
- 与 Java 的 `coreUserManage.getUserName()` 对等
- 保持 handler 层轻量，只做 HTTP 序列化
- UserRepository 已存在且可用

**替代方案**: 在 handler 层直接查 user 表 — 违反分层架构

### Decision 3: SaveField 实现策略

**选择**: 新增 `DatasetService.SaveField()` 方法，根据 `field.ID == 0` 判断创建或更新。

**理由**:
- 前端发送的 field 对象包含所有字段，ID 为空表示新建
- Repository 层已有 `CreateDatasetField()` 和 `UpdateDatasetFieldNames()`
- 与 `CopyField()` 方法使用相同的 repo 接口

### Decision 4: getFunction 返回策略

**选择**: 返回一个静态的函数分类列表（与 Java 的 FunctionConstant 对齐）。

**理由**: SQL 函数列表是配置数据，不需要动态查询。Java 实现也是静态配置。

## Risks / Trade-offs

- **[Risk] CoreDatasetGroup 查询变慢** → 只添加 4 个标量字段，无 JOIN 开销，影响可忽略
- **[Risk] 用户删除后 creator 解析失败** → 返回原始 user ID 作为 fallback，与 Java 行为一致
- **[Risk] SaveField 缺少依赖检查** → Stage2 已有完整的依赖检查框架，SaveField 暂不需要（字段创建/更新不涉及级联影响）
- **[Trade-off]** barInfo 的 `DatasourceDTOList` 和 `IsCross` 字段暂不填充 — 需要跨表 JOIN，复杂度高，延后处理

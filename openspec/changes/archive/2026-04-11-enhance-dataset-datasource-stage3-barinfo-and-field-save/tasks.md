## 1. BarInfo 审计字段 — 模型与查询层

- [ ] 1.1 扩展 `CoreDatasetGroup` 模型：在 `apps/backend-go/internal/domain/dataset/dataset.go` 中添加 `CreateBy`, `CreateTime`, `UpdateBy`, `LastUpdateTime` 字段（对应 `core_dataset_group` 表的 `create_by`, `create_time`, `update_by`, `last_update_time` 列），添加 GORM column tag
- [ ] 1.2 更新 `DatasetRepository.GetGroupByID`：确保查询 `SELECT *` 或显式包含新增审计列，验证 GORM 自动映射正确
- [ ] 1.3 添加用户名解析辅助函数：在 `DatasetService` 中注入 `UserRepository`，添加 `resolveUserName(userID string) string` 方法，查不到用户时 fallback 返回原始 userID

## 2. BarInfo 审计字段 — Handler 与响应

- [ ] 2.1 更新 barInfo handler：在 `compatibility_bridge_handler.go` 的 barInfo 处理逻辑（约 L660-671）中，用 `group` 实例的真实审计字段填充 `BarInfo` 结构体的 `CreateBy`, `CreateTime`, `UpdateBy`, `LastUpdateTime`
- [ ] 2.2 填充 Creator 和 Updater：调用 `resolveUserName` 将 `group.CreateBy` 解析为 `barInfo.Creator`，将 `group.UpdateBy` 解析为 `barInfo.Updater`
- [ ] 2.3 验证 barInfo 响应：运行 `make test` 确保通过，手动 curl 验证 `/datasetTree/barInfo/{id}` 返回非零审计数据

## 3. Datasource Creator 解析

- [ ] 3.1 在 `DatasourceService` 中注入 `UserRepository`，添加 `resolveUserName` 辅助方法（或复用 dataset 的）
- [ ] 3.2 更新 `sanitizeDatasourceResponse`：在 `compatibility_bridge_handler.go`（约 L1247-1274）的响应 map 中添加 `creator`（解析 `create_by` → 用户名）和 `updater`（解析 `update_by` → 用户名）
- [ ] 3.3 验证 datasource 详情响应：curl `/datasource/get/{id}` 和 `/datasource/hidePw/{id}` 确认返回 `creator` 和 `updater` 字段

## 4. DatasetField Save — 服务层实现

- [ ] 4.1 添加 `SaveField` 方法到 `DatasetService`：接收 `CoreDatasetTableField` 参数，根据 `ID == 0` 判断创建或更新；创建时调用 `repo.CreateDatasetField`，更新时调用 `repo.UpdateDatasetField`
- [ ] 4.2 添加 `UpdateDatasetField` 到 `DatasetRepository`：实现按 ID 更新字段所有可修改列的 GORM 操作
- [ ] 4.3 添加字段验证：检查必填字段（name, datasetGroupId, type），缺少时返回明确错误
- [ ] 4.4 单元测试：为 `SaveField` 添加单元测试覆盖创建和更新路径

## 5. DatasetField Save — 路由注册

- [ ] 5.1 在 `compatibility_bridge_handler.go` 中注册 `POST /datasetField/save` 路由，调用 `DatasetService.SaveField`
- [ ] 5.2 验证端点：curl 发送创建和更新请求确认正常响应

## 6. DatasetField 辅助端点

- [ ] 6.1 实现 `/datasetField/getFunction`：返回静态函数分类列表（与 Java `FunctionConstant` 对齐），包含聚合函数、日期函数、字符串函数、数学函数等分类
- [ ] 6.2 实现 `/datasetField/listByDsIds`：接收 datasource ID 数组，查询 `core_dataset_table_field` WHERE `datasource_id IN (?)`，返回字段列表
- [ ] 6.3 在 compatibility bridge 中注册这两个路由
- [ ] 6.4 验证端点：curl 测试确认返回格式正确

## 7. 测试与验证

- [ ] 7.1 后端测试：`make test` 全部通过
- [ ] 7.2 前端类型检查：`npm run ts:check` 通过
- [ ] 7.3 前端 lint：`npm run lint` 通过
- [ ] 7.4 Contract drift check：`make drift-check` 通过（确保新增端点不破坏已有契约）

## Final Verification Wave

- [ ] F1 barInfo 审计数据完整：调用 `/datasetTree/barInfo/{id}` 返回非零 createTime、非空 creator
- [ ] F2 Datasource creator 解析：调用 `/datasource/get/{id}` 返回 creator 用户名（非 user ID）
- [ ] F3 Field save 功能可用：POST `/datasetField/save` 可成功创建和更新字段
- [ ] F4 辅助端点可用：`/datasetField/getFunction` 和 `/datasetField/listByDsIds` 返回正确数据

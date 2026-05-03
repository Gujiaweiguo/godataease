## Why

DataFilling (数据填报) 是 Java 侧最大的未迁移功能模块，提供表单式数据录入、任务调度分发、Excel 导入导出和操作审计。它在 Java 中作为 xpack 企业版插件实现（37 个 API 端点、23 个 DTO、动态 DDL SPI），需要完整迁移到 Go 主线，使其成为社区版可用功能。

## What Changes

**分切片实施（6 个切片）**：

### Slice 1: Form CRUD + Domain（本次）
- 定义 DataFilling 领域模型（forms 表、树形结构、动态列定义 ExtTableField）
- 实现 Repository 层（MySQL CRUD + 树查询）
- 实现 Form CRUD Service（save/createTable、update/alterTable、delete/dropTable、rename、move、get、tree）
- 实现 HTTP Handler + 路由注册
- 数据源列表端点（listDatasourceList、listDatasourceListAll）

### Slice 2: Data CRUD + DDL Provider
- 实现 MySQL DDL Provider（createTableSql、alterTable、dropTable、insertData、updateData、deleteData、searchData）
- 实现表单数据 CRUD（tableData、saveRowData、deleteRowData、batchDelete、truncate）
- 搜索/过滤功能（DataFillFormTableDataSearchParam）
- 选项值查询（listColumnData、extraDetails）

### Slice 3: Task Management
- 任务定义 CRUD（TaskInfoVO：saveTask、info、taskPager）
- 任务调度集成（Go job scheduler + cron）
- 任务生命周期管理（startTask、stopTask、executeNow）
- 子任务管理（subTaskPager、batchDeleteSubTask）

### Slice 4: User Task & Fill
- 用户任务列表（listUserTask、countUserTodoList）
- 用户填报数据（saveFormRowData、appendFormRowData、userTaskDeleteRowData）
- 子任务用户列表（listSubTaskUser）
- 表单模板获取（getTemplateByUserTaskItemId）

### Slice 5: Excel Import/Export + Commit Log
- Excel 上传解析（excelUpload → DfExcelData）
- Excel 确认导入（confirmUpload）
- Excel 模板下载（excelTemplate）
- 数据导出（innerExport）
- 操作日志（logPager、clearLog）

### Slice 6: Frontend Integration + Polish
- 前端 API 模块（api/dataFilling.ts）
- 前端组件对接（如需重建 Vue 组件）
- 权限集成测试
- 端到端验证

## Capabilities

### New Capabilities
- `data-filling`: 数据填报表单管理、动态列定义、DDL 生成、数据 CRUD、任务调度、用户填报、Excel 导入导出、操作日志

### Modified Capabilities
- None（DataFilling 是全新模块，不修改现有 spec）

## Impact

- **Backend**: 新增 `domain/datafilling`、`repository/datafilling_repo.go`、`service/datafilling_service.go`、`service/datafilling_ddl_provider.go`、`handler/datafilling_handler.go`、路由注册
- **Database**: 新增 `data_filling_forms` 表；需执行 DDL 创建
- **Router**: `/api/data-filling` 路由组（37 个端点）+ `/de2api/data-filling` 兼容路由
- **Dependencies**: 复用现有 datasource 连接能力执行动态 DDL/DML
- **Tests**: 每个切片配套单元测试 + 集成测试

# Plan: Go API 兼容桥接

## 任务清单

### PHASE-0 基线盘点
- [x] 统计 Java 前缀与 Go 前缀差异（按 `@RequestMapping` 与 Go handler 路由组）
- [x] 冻结优先迁移前缀清单：`datasource`、`datasetTree`、`datasetData`、`chartData`、`user`、`org`、`msg-center`
- [x] 冻结响应兼容规则：`code/data/msg`、错误码映射、空数据语义（见 compatibility-rules.md）

### BRIDGE-001 路由别名兼容
- [x] 为已实现能力补齐 Java 兼容路径（旧前缀 + `/api` 别名）
- [x] 对历史非常用路径加兼容层路由转发（不改变业务语义）
- [x] 增加路由冲突检测，避免别名覆盖真实业务路由

### BRIDGE-002 契约一致性
- [x] 将关键桩接口替换为真实逻辑：`/datasource/perDelete/:id`、`/datasetData/enumValueObj`、`/datasetData/enumValueDs`、`/datasetData/enumValue`、`/chartData/getFieldData/:fieldId/:fieldType`、`/chartData/getDrillFieldData/:fieldId`、`/chart/save`、`/chart/listByDQ/:id/:chartId`、`/chart/copyField/:id/:chartId`、`/chart/deleteField/:id`、`/chart/deleteFieldByChart/:chartId`、`/datasource/latestUse`、`/datasource/showFinishPage`、`/datasource/setShowFinishPage`
- [x] 对齐成功响应结构（含空数组、空对象、null 的返回策略）
- [x] 对齐常见失败码与消息语义
- [x] 对齐分页字段命名（`list/total/current/size`）

### BRIDGE-003 验证与回归
- [x] 新增 chart 桥接回归用例：`/chart/save`、`/chart/listByDQ/:id/:chartId`、`/chart/copyField/:id/:chartId`、`/chart/deleteField/:id`、`/chart/deleteFieldByChart/:chartId` 及 `/api` 别名路径
- [x] 新增兼容性回归用例（按旧路径调用）
- [x] 新增 `/api` 别名回归用例
- [x] 回归前端关键页面接口（数据源、数据集、图表、用户组织、消息中心）

### INT-001 OpenSpec 与发布门禁
- [x] 更新迁移矩阵（Java -> Go）并记录已完成映射
- [x] 运行 OpenSpec 严格校验

## 里程碑

- [x] M1: 兼容清单冻结
- [x] M2: 高优先级路径别名完成
- [x] M3: 契约一致性完成
- [x] M4: 回归验证通过

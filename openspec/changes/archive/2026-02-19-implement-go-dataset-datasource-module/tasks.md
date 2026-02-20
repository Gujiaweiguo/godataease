# Plan: Go Dataset + Datasource 模块实现

## 任务清单

### PHASE-0 契约与边界确认
- [x] 盘点 Java 侧 ds/dataset 端点并冻结首批迁移 API 清单
- [x] 明确首批支持的数据源类型及连接参数映射
- [x] 形成字段类型映射与错误码映射基线

### DS-001 Domain/Repository
- [x] 完善 `internal/domain/datasource/datasource.go` 领域模型
- [x] 新建 `internal/repository/datasource_repo.go`
- [x] 实现数据源列表查询与连接校验的数据访问逻辑

### DS-002 Service/Handler
- [x] 新建 `internal/service/datasource_service.go`
- [x] 新建 `internal/transport/http/handler/datasource_handler.go`
- [x] 实现 `POST /api/ds/list`
- [x] 实现 `POST /api/ds/validate`

### DATASET-001 Domain/Repository
- [x] 完善 `internal/domain/dataset/dataset.go` 领域模型
- [x] 新建 `internal/repository/dataset_repo.go`
- [x] 实现数据集树、字段、预览查询的数据访问逻辑

### DATASET-002 Service/Handler
- [x] 新建 `internal/service/dataset_service.go`
- [x] 新建 `internal/transport/http/handler/dataset_handler.go`
- [x] 实现 `POST /api/dataset/tree`
- [x] 实现 `POST /api/dataset/fields`
- [x] 实现 `POST /api/dataset/preview`

### INT-001 集成与验证
- [x] 在 `internal/transport/http/router.go` 注册新路由
- [x] 对齐 Java 响应结构（`code/data/msg`）
- [x] 增加关键路径单元测试（连接校验、字段提取、预览查询）
- [x] 运行 OpenSpec 严格校验

## 里程碑

- [x] M1: 契约与边界确认完成
- [x] M2: Datasource 模块完成
- [x] M3: Dataset 模块完成
- [x] M4: 集成验证通过

# Change: 实现 Go 版本 Dataset + Datasource 核心模块

## Why

当前 Go 后端已完成系统管理侧模块（登录、用户、组织、角色、权限、菜单、地图、审计、嵌入），
但 BI 核心链路（Datasource -> Dataset -> Chart -> Visualization）仍未迁移完成。

其中 Datasource 与 Dataset 是上层图表和仪表板能力的前置依赖，需优先落地。

## What Changes

### 本次范围

- 实现 Go 版本 Datasource 核心能力（连接配置、列表查询、可用性校验）
- 实现 Go 版本 Dataset 核心能力（数据集树、字段元数据、基础预览查询）
- 定义并冻结与 Java 对齐的首批 API 契约（响应 code/data/msg 保持一致）
- 在 `backend-go/internal/transport/http/router.go` 注册对应路由

### 不包含

- Chart 模块迁移
- Visualization/Dashboard 模块迁移
- Export Center 模块迁移
- 第三方通信集成（Lark/WeCom/DingTalk）

## Impact

### 代码影响（计划）

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/datasource/datasource.go` | 更新 |
| `backend-go/internal/domain/dataset/dataset.go` | 更新 |
| `backend-go/internal/repository/datasource_repo.go` | 新增 |
| `backend-go/internal/repository/dataset_repo.go` | 新增 |
| `backend-go/internal/service/datasource_service.go` | 新增 |
| `backend-go/internal/service/dataset_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/datasource_handler.go` | 新增 |
| `backend-go/internal/transport/http/handler/dataset_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点（首批）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/ds/list | 查询数据源列表 |
| POST | /api/ds/validate | 校验数据源连接 |
| POST | /api/dataset/tree | 查询数据集树 |
| POST | /api/dataset/fields | 查询数据集字段 |
| POST | /api/dataset/preview | 预览数据集数据 |

## Risk

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| Java/Go SQL 行为差异 | 高 | 引入契约样例集，对关键 SQL 场景做回归对比 |
| 多数据源连接参数兼容 | 中 | 优先迁移主流数据源（MySQL/PostgreSQL），其余分批扩展 |
| 字段类型映射差异 | 中 | 建立统一字段类型映射表并加入集成测试 |

## Exit Criteria

- [ ] Datasource 与 Dataset 首批 API 在 Go 侧可用
- [ ] 与 Java 侧响应契约兼容（code/data/msg）
- [ ] OpenSpec 变更验证通过
- [ ] Go 模块可编译，并完成最小回归验证

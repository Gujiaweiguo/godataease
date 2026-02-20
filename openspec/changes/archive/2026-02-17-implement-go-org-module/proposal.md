# Change: 实现 Go 版本 Organization 模块

## Why

作为 Java 到 Go 渐进式迁移的第三个业务模块，Organization（组织管理）模块具有以下优势：
- 用户模块已实现，组织与用户有关联关系
- 组织管理是核心基础模块，权限和数据隔离依赖组织
- 数据库表结构已完整定义（sys_org 表）
- API 契约清晰，8 个核心端点易于验证

## What Changes

### 本次范围

- 实现 Go 版本的组织实体（SysOrg）
- 实现 GORM Repository 层（含树形结构查询）
- 实现 Service 层业务逻辑
- 实现 HTTP Handler 和 API 端点
- 与 Java 版本 API 保持兼容

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 组织成员管理（将在后续模块实现）
- 其他业务模块迁移

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/org/org.go` | 更新 |
| `backend-go/internal/repository/org_repo.go` | 新增 |
| `backend-go/internal/service/org_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/org_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/system/organization/create | 创建组织 |
| POST | /api/system/organization/update | 更新组织 |
| POST | /api/system/organization/delete/:orgId | 删除组织 |
| GET | /api/system/organization/list | 获取组织列表 |
| GET | /api/system/organization/info/:orgId | 获取组织详情 |
| GET | /api/system/organization/tree | 获取组织树 |
| GET | /api/system/organization/checkName | 检查组织名称 |
| POST | /api/system/organization/updateStatus | 更新组织状态 |
| GET | /api/system/organization/children/:parentId | 获取子组织 |

### 数据库表

| 表名 | 说明 |
|------|------|
| sys_org | 组织主表 |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| API 兼容性 | 中 | 保持与 Java 相同的请求/响应格式 |
| 树形结构查询 | 中 | 使用递归 CTE 或层级字段 |
| 数据一致性 | 低 | 使用相同的数据库表 |

## Exit Criteria

- [ ] 所有 API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容
- [ ] 树形结构查询正确
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过

# Change: 实现 Go 版本 Permission 模块

## Why

作为 Java 到 Go 渐进式迁移的第四个业务模块，Permission（权限配置）模块具有以下优势：
- 用户和组织模块已实现，权限与用户/组织有关联
- 权限是核心安全模块，控制菜单、资源、数据访问
- 数据库表结构已完整定义（sys_perm 表）
- API 契约清晰，4 个核心端点易于验证

## What Changes

### 本次范围

- 实现 Go 版本的权限实体（SysPerm）
- 实现 GORM Repository 层
- 实现 Service 层业务逻辑
- 实现 HTTP Handler 和 API 端点
- 与 Java 版本 API 保持兼容

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 行级/列级权限过滤逻辑（将在后续迭代实现）
- 其他业务模块迁移

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/permission/permission.go` | 更新 |
| `backend-go/internal/repository/perm_repo.go` | 新增 |
| `backend-go/internal/service/perm_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/perm_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/system/permission/list | 分页查询权限列表 |
| POST | /api/system/permission/create | 创建权限 |
| POST | /api/system/permission/update | 更新权限 |
| POST | /api/system/permission/delete/:permId | 删除权限 |

### 数据库表

| 表名 | 说明 |
|------|------|
| sys_perm | 权限主表 |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| API 兼容性 | 中 | 保持与 Java 相同的请求/响应格式 |
| 数据一致性 | 低 | 使用相同的数据库表 |

## Exit Criteria

- [ ] 所有 API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过

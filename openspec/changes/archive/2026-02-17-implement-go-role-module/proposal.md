# Change: 实现 Go 版本 Role（角色管理）模块

## Why

作为 Java 到 Go 渐进式迁移的第七个业务模块，Role（角色管理）模块是权限体系的核心组成部分：
- 与已实现的 User、Org、Permission 模块紧密关联
- 角色是用户权限分配的基础载体
- 实现核心 CRUD API，为后续复杂权限逻辑打下基础

## What Changes

### 本次范围

- 实现 Go 版本的角色实体（SysRole、SysRoleMenu、SysRolePerm）
- 实现 GORM Repository 层
- 实现 Service 层业务逻辑
- 实现 HTTP Handler 和核心 API 端点
- 与 Java 版本 API 保持兼容

### 不包含

- 前端代码修改
- 数据库 schema 变更
- 用户-角色绑定功能（后续迭代）
- 角色权限分配功能（后续迭代）

## Impact

### 代码影响

| 文件 | 变更类型 |
|------|----------|
| `backend-go/internal/domain/role/role.go` | 新增 |
| `backend-go/internal/repository/role_repo.go` | 新增 |
| `backend-go/internal/service/role_service.go` | 新增 |
| `backend-go/internal/transport/http/handler/role_handler.go` | 新增 |
| `backend-go/internal/transport/http/router.go` | 更新 |

### API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/role/query | 查询角色列表 |
| POST | /api/role/create | 创建角色 |
| POST | /api/role/edit | 编辑角色 |
| POST | /api/role/delete/:id | 删除角色 |
| GET | /api/role/detail/:id | 角色详情 |

### 数据库表

| 表名 | 说明 |
|------|------|
| sys_role | 角色主表 |

### 风险评估

| 风险 | 级别 | 缓解措施 |
|------|------|----------|
| API 兼容性 | 中 | 保持与 Java 相同的请求/响应格式 |
| 数据一致性 | 低 | 使用相同的数据库表 |

## Exit Criteria

- [ ] 所有 API 端点实现并通过测试
- [ ] 响应格式与 Java 版本兼容（code: 000000/500000）
- [ ] 代码通过 lint 检查
- [ ] OpenSpec 验证通过

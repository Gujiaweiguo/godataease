## Why

当前角色查询 API (`/role/query`) 返回全量数据，当角色数量增长时会导致前端性能问题。前端角色管理页面需要分页能力来支持大规模角色列表展示。

## What Changes

- 新增角色分页查询 API (`/role/page`)
- 支持按关键字、角色类型筛选
- 支持分页参数 (current, size)
- 返回分页结果 (list, total)

## Capabilities

### New Capabilities

(none - extending existing capability)

### Modified Capabilities

- `role-management`: 新增角色分页查询 API，扩展查询能力

## Impact

### 后端
- `apps/backend-go/internal/domain/role/role.go` - 新增 RolePageRequest, RolePageResult 结构
- `apps/backend-go/internal/repository/role_repo.go` - 新增 QueryWithPage 方法
- `apps/backend-go/internal/service/role_service.go` - 新增 QueryRolesPage 方法
- `apps/backend-go/internal/transport/http/handler/role_handler.go` - 新增 Page handler
- `apps/backend-go/internal/transport/http/handler/role_handler.go` - 注册 `/role/page` 路由

### 前端
- 前端已有分页 UI，只需调用新 API

### 测试
- 新增服务层单元测试
- 新增处理器单元测试

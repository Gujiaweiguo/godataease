## Why

当前菜单数据通过硬编码方式返回，无法在运行时动态配置菜单可见性和授权。角色管理页面缺少菜单授权配置能力，管理员无法灵活控制不同角色的菜单访问权限。此变更将消除硬编码菜单行为，引入数据驱动的菜单管理和角色菜单授权机制，同时保持迁移期间的兼容性。

## What Changes

- 新增 `sys_role_menu` 关联表，存储角色与菜单的授权映射
- 实现角色菜单仓储与服务层，支持幂等保存和授权查询
- 实现统一菜单组装服务，按角色过滤菜单树和路由
- 新增菜单管理 API（CRUD、排序、显隐控制）
- 新增角色-菜单授权 API（查询、保存）
- 改造兼容端点 `/api/roleRouter/query`、`/api/auth/menuResource` 为动态数据输出
- 移除或开关化硬编码菜单逻辑

## Capabilities

### New Capabilities
- `menu-management`: 菜单 CRUD、排序、显隐管理，支持动态菜单树维护

### Modified Capabilities
- `role-management`: 新增角色-菜单授权配置能力和授权查询/保存 API
- `api-compatibility-bridge`: 兼容端点从硬编码改为动态数据输出，保持前端路由解析兼容
- `permission-config`: 角色菜单授权与菜单可见性控制的集成

## Impact

- **数据库变更**:
  - 新增 `sys_role_menu` 表（role_id, menu_id 唯一约束）
  - Bootstrap 管理员菜单授权初始化数据
- **后端代码**:
  - `apps/backend-go/internal/domain/menu/` - 菜单领域模型扩展
  - `apps/backend-go/internal/domain/role/` - 角色菜单关系
  - `apps/backend-go/internal/transport/http/handler/` - 菜单管理、角色授权 API
  - `apps/backend-go/internal/transport/http/handler/frontend_compat_handler.go` - 兼容端点改造
- **前端集成**（验证目标）:
  - `apps/frontend/src/views/system/menu/` - 菜单管理页面
  - `apps/frontend/src/views/system/role/` - 角色授权配置
  - `apps/frontend/src/router/` - 动态路由消费
- **运维影响**: 减少菜单变更的代码发布需求，支持运行时配置

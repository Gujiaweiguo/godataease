## Why

当前仓库虽然已有权限、菜单和资源授权模块，但官方手册要求的是一个统一的权限配置中心，而不是分散的局部入口。如果在角色模型未稳定前就直接堆叠权限页面和授权规则，最终会同时出现重复赋权语义、菜单可见性漂移和资源/数据权限行为不一致的问题。

## What Changes

- 将权限管理对齐为统一权限配置中心，统一承载菜单权限、资源权限与行列权限三类工作流。
- 对齐菜单权限行为：菜单授权仅绑定角色，未授权访问必须返回明确的权限拒绝语义而不是误导性 404。
- 对齐资源权限行为：支持按用户/按资源双视角配置一致性，以及资源分组权限继承在新增资源上的自动生效。
- 对齐数据访问控制：将行权限和列权限作为统一权限中心中的一等工作流，而不是外围零散能力。
- 明确本 change 只消费稳定的组织/用户/角色模型，不重定义角色成员流程或登录身份基线。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `permission-config`: 扩展为统一权限配置中心，覆盖菜单权限、资源权限、行列权限及双视角一致性约束。
- `menu-access-governance`: 收紧菜单授权拒绝语义与角色授权后的菜单可见性一致性要求。
- `role-management`: 仅补充角色作为菜单/资源授权载体时的边界约束，避免在本 change 中展开角色成员流程。

## Impact

- Affected backend:
  - `apps/backend-go/internal/service/perm_service.go`
  - `apps/backend-go/internal/service/resource_perm_service.go`
  - `apps/backend-go/internal/service/role_menu_service.go`
  - `apps/backend-go/internal/transport/http/handler/perm_handler.go`
  - `apps/backend-go/internal/transport/http/handler/role_menu_handler.go`
  - `apps/backend-go/internal/transport/http/middleware/permission_middleware.go`
  - `apps/backend-go/internal/transport/http/middleware/menu_auth.go`
- Affected frontend:
  - `apps/frontend/src/views/system/permission/index.vue`
  - `apps/frontend/src/store/modules/permission.ts`
  - `apps/frontend/src/api/auth.ts`
  - `apps/frontend/src/api/menu.ts`
- Affected specs:
  - `openspec/specs/permission-config/spec.md`
  - `openspec/specs/menu-access-governance/spec.md`
  - `openspec/specs/role-management/spec.md`

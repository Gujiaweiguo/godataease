## Why

godataease 已经具备组织、用户、角色、权限管理的基础实现，但当前风险不在“有没有页面”，而在“语义是否与官方手册一致、组织作用域是否稳定、兼容桥是否掩盖了行为偏差”。如果继续按页面零散修补，会把组织隔离、角色继承、最后角色策略、菜单/资源/数据权限 enforcement 反复返工。

## What Changes

- 冻结用户提供的官方 v2 手册作为管理域基线，并把差异统一分为“已实现但不一致 / 部分实现 / 未实现”。
- 将默认 4-change 页面导向拆分重组为 3 个语义 change：IAM foundation、user-role lifecycle、permission center。
- 修正共享语义与兼容性错配：组织删除策略、最后角色策略、legacy route/compat API、菜单鉴权 stub、行权限 middleware stub。
- 补齐用户/角色生命周期闭环：组织选择、角色多选、导入/错误报告、权限边界验证、外部用户挂载。
- 补齐权限中心闭环：菜单/资源双视角一致性、target 授权通路、行/列权限、白名单/系统变量、P2 后置能力收敛。

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `rbac-docs-alignment-governance`: 将“文档对齐”从能力补点升级为“基线冻结 + 差异矩阵 + 3-change 依赖治理”。
- `organization-management`: 明确组织树、组织隔离、删除子组织前置与资源处置策略。
- `user-management`: 明确用户创建组织选择、启停、重置密码、导入/错误报告、第三方来源字段的对齐边界。
- `role-management`: 明确内置角色基线、自定义角色继承约束、最后角色策略、角色成员管理边界。
- `permission-config`: 明确按用户/按资源双视角一致性、菜单/资源/行/列权限、白名单与系统变量。
- `menu-access-governance`: 明确菜单权限 enforcement 与 role-bound menu access 的验证要求。
- `api-compatibility-bridge`: 明确 legacy frontend API、compat handler、canonical route 的过渡策略与退役顺序。

## Impact

- Affected backend:
  - `apps/backend-go/internal/service/org_service.go`
  - `apps/backend-go/internal/service/user_service.go`
  - `apps/backend-go/internal/service/user_import_service.go`
  - `apps/backend-go/internal/service/role_service.go`
  - `apps/backend-go/internal/service/resource_perm_service.go`
  - `apps/backend-go/internal/service/row_permission_service.go`
  - `apps/backend-go/internal/service/column_permission_service.go`
  - `apps/backend-go/internal/transport/http/middleware/menu_auth.go`
  - `apps/backend-go/internal/transport/http/middleware/permission.go`
  - `apps/backend-go/internal/transport/http/handler/permission_compat_handler.go`
  - `apps/backend-go/internal/transport/http/router.go`
- Affected frontend:
  - `apps/frontend/src/api/user.ts`
  - `apps/frontend/src/api/auth.ts`
  - `apps/frontend/src/views/system/org/index.vue`
  - `apps/frontend/src/views/system/user/index.vue`
  - `apps/frontend/src/views/system/user/RoleTab.vue`
  - `apps/frontend/src/views/system/permission/MenuPermission.vue`
  - `apps/frontend/src/views/system/permission/ResourcePermission.vue`
  - `apps/frontend/src/views/system/permission/DataPermission.vue`
- Affected specs:
  - `openspec/specs/rbac-docs-alignment-governance/spec.md`
  - `openspec/specs/organization-management/spec.md`
  - `openspec/specs/user-management/spec.md`
  - `openspec/specs/role-management/spec.md`
  - `openspec/specs/permission-config/spec.md`
  - `openspec/specs/menu-access-governance/spec.md`
  - `openspec/specs/api-compatibility-bridge/spec.md`

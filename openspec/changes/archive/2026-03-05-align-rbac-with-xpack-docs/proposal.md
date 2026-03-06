## Why

当前仓库的用户/角色/组织/权限模块已具备基础能力，但与 DataEase v2 X-Pack 文档仍存在关键能力缺口（例如用户批量导入、角色成员管理、角色继承约束的端到端闭环）。如果不先通过规格化重构补齐能力并统一行为边界，后续菜单动态化和权限演进会持续引入回归风险。

## What Changes

- 对齐用户管理能力：补齐 Excel 模板导入、导入校验与错误报告、重置密码流程。
- 对齐角色管理能力：补齐“添加组织用户/外部用户”“移除角色成员”“唯一角色保护”能力，并明确继承约束。
- 对齐权限配置能力：强化“按用户/按资源”双视角一致性与资源分组权限继承生效规则。
- 对齐组织管理语义：保持多级组织与隔离策略，并明确删除组织的资源处置策略（实现行为与文档语义一致）。
- 统一前后端契约：补齐缺失兼容接口，消除前端页面存在功能入口但后端能力不完整的情况。

## Capabilities

### New Capabilities

- `rbac-docs-alignment-governance`: 定义文档对齐的验收门禁、迁移顺序和回滚约束。

### Modified Capabilities

- `user-management`: 增补用户批量导入、错误报告、重置密码等规范性行为。
- `role-management`: 增补角色成员管理、继承约束与唯一角色保护规则。
- `permission-config`: 明确双视角配置一致性、资源分组继承与生效时机。
- `organization-management`: 明确组织删除与资源处置策略，保持隔离语义可验证。

## Impact

- Affected backend:
  - `apps/backend-go/internal/transport/http/handler/user_handler.go`
  - `apps/backend-go/internal/transport/http/handler/role_handler.go`
  - `apps/backend-go/internal/transport/http/handler/permission_compat_handler.go`
  - `apps/backend-go/internal/service/user_service.go`
  - `apps/backend-go/internal/service/role_service.go`
  - `apps/backend-go/internal/service/resource_perm_service.go`
- Affected frontend:
  - `apps/frontend/src/views/system/user/index.vue`
  - `apps/frontend/src/views/system/role/index.vue`
  - `apps/frontend/src/views/system/permission/index.vue`
  - `apps/frontend/src/api/user.ts`
  - `apps/frontend/src/api/auth.ts`
- Affected specs:
  - `openspec/specs/user-management/spec.md`
  - `openspec/specs/role-management/spec.md`
  - `openspec/specs/permission-config/spec.md`
  - `openspec/specs/organization-management/spec.md`

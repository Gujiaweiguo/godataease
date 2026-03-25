## Why

在 foundation 稳定之前，角色能力容易同时承担“主数据定义”和“授权运营”两类职责，导致角色边界不断膨胀。这个 change 的目标是把角色视为业务对象来治理其生命周期、成员关系和继承约束，让后续权限中心只消费稳定的角色模型，而不再回头重写角色规则。

## What Changes

- 对齐角色 CRUD 与详情/查询语义，确保组织内角色生命周期行为稳定可验证。
- 对齐角色成员管理：覆盖添加组织用户、添加外部用户、移除成员与幂等关系维护。
- 对齐唯一角色安全策略：明确移除用户最后一个角色时的确定性系统行为。
- 对齐自定义角色继承约束：要求自定义角色继承允许的内置角色，且不可越过父级权限边界。
- 将角色工作流在前端信息架构上明确落到用户管理页中的角色页签，避免运营入口分散。
- 显式声明非目标：本 change 不承载菜单/资源/行列权限统一配置中心，不扩展新的授权引擎。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `role-management`: 增补角色生命周期、成员管理、唯一角色保护、自定义角色继承边界与相关兼容接口要求。
- `role-in-user-management`: 明确角色工作流在用户管理页面中的托管入口、页签结构与保留行为。
- `user-management`: 仅补充与角色成员运营直接相关的交叉约束，避免在该 spec 中重新定义角色语义。

## Impact

- Affected backend:
  - `apps/backend-go/internal/service/role_service.go`
  - `apps/backend-go/internal/service/user_service.go`
  - `apps/backend-go/internal/transport/http/handler/role_handler.go`
  - `apps/backend-go/internal/transport/http/handler/user_handler.go`
  - `apps/backend-go/internal/repository/role_repo.go`
  - `apps/backend-go/internal/repository/user_repo.go`
- Affected frontend:
  - `apps/frontend/src/views/system/role/index.vue`
  - `apps/frontend/src/views/system/user/index.vue`
  - `apps/frontend/src/api/user.ts`
  - `apps/frontend/src/api/auth.ts`
- Affected specs:
  - `openspec/specs/role-management/spec.md`
  - `openspec/specs/role-in-user-management/spec.md`
  - `openspec/specs/user-management/spec.md`

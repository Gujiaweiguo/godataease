## Why

当前仓库已经具备用户、组织、角色、登录等基础模块，但这些能力对 X-Pack 官方手册中的“组织边界、用户归属、默认角色、组织切换、当前身份上下文”还没有形成统一且可复用的主数据基线。如果不先锁定 foundation，后续角色流程与权限中心会重复定义组织和身份规则，导致 spec 边界重叠、实现返工和回归风险持续放大。

## What Changes

- 对齐组织基础语义：明确多级组织树、叶子组织删除约束、资源处置策略与跨组织隔离规则。
- 对齐身份基础语义：明确当前用户资料、当前组织上下文、组织切换与登录引导返回中的身份信息契约。
- 对齐用户归属模型：明确用户与组织、成员关系、状态管理、重置密码和导入等基础行为的组织作用域。
- 对齐基础角色模型：明确系统级与组织级内置角色的基线语义，以及后续 role workflows / permission center 可复用的共享查询接口与边界。
- 显式声明非目标：本 change 不承载完整角色成员运营流程，不承载统一权限配置中心，也不展开资源级/行列级授权编排。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `organization-management`: 收紧组织树、删除约束、资源处置和组织切换前提，使其成为后续 IAM change 的组织边界基线。
- `user-management`: 明确用户归属组织、成员关系、组织作用域下的用户生命周期与管理员基础操作契约。
- `login-management`: 扩展登录后身份引导与当前用户上下文返回，确保前端能建立稳定的组织/身份状态。
- `role-management`: 仅补充基础角色模型、内置角色语义和共享查询边界，不在本 change 中展开完整角色工作流。

## Impact

- Affected backend:
  - `apps/backend-go/internal/service/org_service.go`
  - `apps/backend-go/internal/service/user_service.go`
  - `apps/backend-go/internal/service/auth_service.go`
  - `apps/backend-go/internal/service/role_service.go`
  - `apps/backend-go/internal/transport/http/handler/org_handler.go`
  - `apps/backend-go/internal/transport/http/handler/user_handler.go`
  - `apps/backend-go/internal/transport/http/handler/auth_handler.go`
- Affected frontend:
  - `apps/frontend/src/store/modules/user.ts`
  - `apps/frontend/src/views/system/org/index.vue`
  - `apps/frontend/src/views/system/user/index.vue`
  - `apps/frontend/src/api/org.ts`
  - `apps/frontend/src/api/user.ts`
  - `apps/frontend/src/api/auth.ts`
- Affected specs:
  - `openspec/specs/organization-management/spec.md`
  - `openspec/specs/user-management/spec.md`
  - `openspec/specs/login-management/spec.md`
  - `openspec/specs/role-management/spec.md`

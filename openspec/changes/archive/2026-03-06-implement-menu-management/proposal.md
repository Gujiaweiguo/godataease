## Why

仓库后端已具备 `core_menu` 菜单模型和 CRUD 接口，但前端缺少菜单管理页面，且顶部导航仍存在白名单/硬编码排序逻辑。这导致“菜单可配置”与“导航真实渲染”未形成闭环，影响权限可维护性和可观测性。

## What Changes

- 新增菜单管理控制台，支持菜单树展示及菜单元数据编辑（名称、路径、组件、图标、排序、隐藏）。
- 统一顶部菜单与左侧菜单的运行时数据源，移除前端硬编码过滤/排序逻辑。
- 保留并复用现有角色-菜单授权链路（`sys_role_menu` + `/roleMenu/auth`），确保授权模型不破坏。
- 增加菜单变更后的实时生效策略（刷新后可见、授权后可见、隐藏后不可见）。

## Capabilities

### New Capabilities

- `navigation-rendering`: 定义顶部/左侧导航由同一权限菜单树驱动的运行时规则。

### Modified Capabilities

- `menu-management`: 扩展为完整菜单管理（管理端 UI + 字段校验 + 变更生效）。
- `permission-config`: 明确菜单授权结果与导航可见性的绑定关系。

## Impact

- Affected backend:
  - `apps/backend-go/internal/transport/http/handler/menu_handler.go`
  - `apps/backend-go/internal/service/menu_service.go`
  - `apps/backend-go/internal/service/role_menu_service.go`
- Affected frontend:
  - `apps/frontend/src/layout/components/Header.vue`
  - `apps/frontend/src/layout/components/Menu.vue`
  - `apps/frontend/src/permission.ts`
  - `apps/frontend/src/api/menu.ts` (new)
  - `apps/frontend/src/views/system/menu/index.vue` (new)
- Affected specs:
  - `openspec/specs/menu-management/spec.md`
  - `openspec/specs/permission-config/spec.md`
  - `openspec/specs/navigation-rendering/spec.md` (new capability)

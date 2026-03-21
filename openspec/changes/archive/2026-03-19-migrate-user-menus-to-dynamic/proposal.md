# Proposal: Migrate User Menus to Dynamic

## Why

右上角用户菜单（头像下拉菜单）和"..."帮助菜单目前是硬编码的，不在菜单管理系统里。这导致：
1. 菜单配置分散在多个 Vue 组件中，难以统一管理
2. 无法通过菜单管理界面动态配置菜单项
3. 无法基于角色灵活控制菜单可见性
4. 菜单变更需要修改代码并重新部署

将硬编码菜单迁移到动态菜单管理系统，可以实现菜单的统一配置、动态授权和可追溯性，避免后期问题反复。

## What Changes

### 数据库层
- 扩展 `core_menu` 表结构，新增 `menu_location` 字段区分菜单位置（`sidebar` / `header` / `user-dropdown` / `more-menu`）
- 扩展 `menu_type` 字段支持更多类型（`route` / `event` / `external` / `group` / `component`）
- 新增 `action_config` JSON 字段存储动作配置（事件名、外部URL等）
- 插入用户菜单和帮助菜单的初始数据

### 后端层
- 扩展 `MenuVO` 结构，返回 `menu_location`、`menu_type`、`action_config` 字段
- 扩展 `/roleRouter/query` API，支持按 `menu_location` 过滤菜单
- 新增事件类型菜单的处理逻辑

### 前端层
- 重构 `AccountOperator.vue`，从动态菜单加载用户下拉菜单项
- 重构 `MoreMenu.vue`，从动态菜单加载帮助链接
- 新增菜单事件处理器映射（`open-about-dialog`、`user-logout` 等）
- 保留 `LangSelector` 和 `logout` 等核心功能的硬编码（或可选动态化）

### 迁移数据
- 将现有硬编码菜单项迁移到 `core_menu` 表
- 包括：关于、系统设置、修改密码、帮助文档、产品论坛、技术博客、企业版试用

## Capabilities

### New Capabilities

- `dynamic-user-menu`: 动态用户菜单管理，支持从数据库配置用户下拉菜单项，包括路由跳转、事件触发、外部链接等类型
- `dynamic-help-links`: 动态帮助链接管理，支持从数据库配置帮助菜单的外部链接

### Modified Capabilities

- `menu-access-governance`: 扩展现有菜单管理能力，支持 `menu_location` 字段区分不同位置的菜单（顶部导航、侧边栏、用户下拉、帮助菜单）

## Impact

### 受影响的代码

**后端（Go）**
- `internal/domain/menu/menu.go` — 扩展 `CoreMenu` 和 `MenuVO` 结构
- `internal/service/menu_service.go` — 扩展查询逻辑支持 `menu_location` 过滤
- `internal/transport/http/handler/frontend_compat_handler.go` — 扩展 `/roleRouter/query` 响应
- `migrations/mysql/` — 新增数据库迁移脚本

**前端（Vue 3）**
- `src/layout/components/AccountOperator.vue` — 重构为动态菜单加载
- `src/layout/components/MoreMenu.vue` — 重构帮助链接为动态加载
- `src/store/modules/permission.ts` — 扩展支持用户菜单状态
- `src/layout/components/menu-utils.ts` — 扩展菜单过滤逻辑
- `src/router/establish.ts` — 扩展支持 `event` 类型菜单
- `src/utils/menu-actions.ts` — **新增** 菜单事件处理器映射

**数据库**
- `core_menu` 表 — 新增字段和初始数据

### 兼容性

- **向后兼容**：现有顶部导航和侧边栏菜单不受影响
- **渐进迁移**：可以先迁移部分菜单，保留核心功能硬编码
- **回滚方案**：如果动态菜单加载失败，可以降级到硬编码模式

### 风险

1. **菜单加载失败**：需要确保动态菜单加载有降级方案
2. **权限配置错误**：需要确保管理员菜单不会被普通用户看到
3. **事件处理遗漏**：需要完整映射所有菜单事件处理器

## Success Criteria

1. 所有硬编码菜单项都可以在 `core_menu` 表中配置
2. 用户下拉菜单和帮助菜单从 API 动态加载
3. 菜单可见性可以通过角色权限控制
4. 现有功能不受影响（关于对话框、修改密码、帮助文档等正常工作）
5. 回归测试全部通过

## Timeline Estimate

- 设计（design.md）：1 天
- 后端实现：2-3 天
- 前端实现：3-4 天
- 测试和修复：1-2 天
- **总计：7-10 天**

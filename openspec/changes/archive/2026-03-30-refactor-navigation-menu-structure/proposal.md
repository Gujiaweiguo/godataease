## Why

当前一级菜单把"组织/权限管理"和"系统配置"混在同一个「系统管理」分组下，导致管理员侧边栏臃肿、职责边界模糊。同时，帮助文档入口占据右上角 More 菜单但与核心业务无关，工具箱被隐藏在二级弹出层里导致数据导出中心入口过深。需要重组信息架构，让菜单分组语义清晰、入口层级扁平。

## What Changes

- **拆分「系统管理」为两个一级菜单**：新建一级菜单「组织权限」（auth=1, 仅管理员可见），挂载用户管理、组织管理、角色管理、权限管理；原「系统设置」保留当前数据库基线中的受治理系统配置子项（本次验证环境下为菜单管理及审计相关菜单）
- **迁移「菜单管理」到「系统设置」**：菜单管理是系统级配置，从原系统管理（现为组织权限）移到系统设置下
- **升级「工具箱」为一级菜单**：取消 hidden 状态，设为一级侧边栏菜单（sort=6, 所有人可见），数据导出中心作为其二级子菜单，便于后续扩展更多工具项
- **移除右上角 More 菜单**：删除 MoreMenu.vue 组件及帮助文档（帮助文档、产品论坛、技术博客、企业版试用）的全部入口
- **精简用户菜单**：移除管理员头像下拉中的「系统设置」快捷入口（`/sys-setting/parameter`），仅保留关于、修改密码、语言、退出系统
- **重排 menu_sort**：工作台=1, 可视化=2, 数据管理=3, 组织权限=4, 系统设置=5, 工具箱=6

**BREAKING**：管理员侧边栏结构从一级 4 项变为一级 6 项；原 MoreMenu 帮助入口全部消失；管理员头像下拉不再有系统设置快捷入口。

## Capabilities

### New Capabilities

（无——本次变更是对已有导航分组和入口的重组，不引入新的业务能力）

### Modified Capabilities

- `top-menu-restructure`：一级导航从四项（工作台/可视化/数据管理/系统管理）扩展为六项（新增组织权限、系统设置拆分独立、工具箱升级为一级）
- `navigation-rendering`：侧边栏需支持 menu_type='event' 的二级菜单点击（工具箱下的数据导出中心触发事件而非路由跳转）
- `dynamic-help-links`：移除所有帮助链接入口和 MoreMenu 组件，此 capability 的运行时可见性将降为零
- `dynamic-user-menu`：移除管理员「系统设置」快捷入口，仅保留关于/修改密码/语言/退出
- `menu-management`：菜单管理页面从原系统管理（现组织权限）迁移到系统设置下，页面本身不变但父菜单路径变更
- `export-center-management`：数据导出中心入口从 MoreMenu 弹出层迁移到侧边栏工具箱二级菜单

## Impact

- **后端迁移**：1 个新 SQL 文件（`20260329_refactor_menu_final.sql`），涉及 core_menu INSERT/UPDATE/DELETE + sys_role_menu 授权
- **前端组件**：删除 `MoreMenu.vue`，修改 `Header.vue`（移除 MoreMenu 引用）、`AccountOperator.vue`（移除系统设置快捷入口）、`Menu.vue`/`menu-utils.ts`（支持 event 类型二级菜单点击）
- **前端国际化**：zh-CN/en/tw 添加 `commons.org_permission` 翻译
- **E2E 测试**：删除 `help-menu.spec.ts`，更新 `system-management-menu-smoke.spec.ts` 导航断言
- **不涉及**：所有 `views/system/` 页面组件、路由注册逻辑、权限中间件、菜单域模型和 service 层

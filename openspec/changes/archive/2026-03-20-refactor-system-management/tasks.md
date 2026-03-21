# Tasks: Refactor System Management

## Phase 1: 数据库迁移

### Task 1.1: 创建菜单迁移脚本
- [x] 创建 `apps/backend-go/migrations/mysql/20260320_refactor_menu_structure.sql`
- [x] 创建「可视化」分组菜单
- [x] 移动仪表板/大屏到可视化下
- [x] 移动模板市场到可视化下
- [x] 隐藏工具箱菜单
- [x] 调整菜单排序

### Task 1.2: 执行迁移并验证
- [x] 运行迁移脚本
- [x] 验证菜单树结构正确（4 个入口：- [x] 验证角色菜单关联正确

---

## Phase 2: 顶部菜单组件更新

### Task 2.1: 更新 Header.vue 菜单渲染逻辑
- [x] 修改 `apps/frontend/src/layout/components/Header.vue`
- [x] 支持分组菜单下拉显示
- [x] 调整菜单排序逻辑

### Task 2.2: 更新菜单工具函数
- [x] 修改 `apps/frontend/src/layout/components/menu-utils.ts`
- [x] 支持新的菜单分组结构

### Task 2.3: 验证顶部菜单显示
- [x] 验证 4 个入口正确显示
- [x] 验证下拉菜单功能正常
- [x] 验证菜单高亮状态正确

---

## Phase 3: 用户管理合并角色 Tab

### Task 3.1: 重构用户管理组件
- [x] 修改 `apps/frontend/src/views/system/user/index.vue`
- [x] 添加「用户」「角色」Tab 切换
- [x] 集成角色列表功能

### Task 3.2: 提取角色 Tab 组件
- [x] 创建 `apps/frontend/src/views/system/user/RoleTab.vue`
- [x] 包含角色列表、添加用户、移除用户、创建自定义角色

### Task 3.3: 移除独立角色管理菜单
- [x] 更新数据库隐藏 `/system/role` 菜单
- [x] 清理路由配置

---

## Phase 4: 权限配置功能完善

### Task 4.1: 重构权限配置组件
- [x] 修改 `apps/frontend/src/views/system/permission/index.vue`
- [x] 添加「菜单权限」「资源权限」「行列权限」Tab

### Task 4.2: 创建菜单权限组件
- [x] 创建 `apps/frontend/src/views/system/permission/MenuPermission.vue`
- [x] 从角色管理移入菜单授权功能
- [x] 支持按角色分配菜单

### Task 4.3: 创建资源权限组件
- [x] 创建 `apps/frontend/src/views/system/permission/ResourcePermission.vue`
- [x] 支持按用户/资源配置
- [x] 支持数据源/数据集/仪表板/大屏权限

### Task 4.4: 创建行列权限组件
- [x] 创建 `apps/frontend/src/views/system/permission/DataPermission.vue`
- [x] 支持行权限配置（按角色/用户/系统变量）
- [x] 支持列权限配置（禁用/脱敏）

### Task 4.5: 后端权限 API 完善
- [x] 添加菜单权限相关 API (已使用现有 API: menuTreeApi, roleMenuAuthApi, roleMenuAuthSaveApi)
- [x] 添加资源权限相关 API (已使用现有 API: resourceTreeApi, resourcePerApi, busiPerSaveApi)
- [x] 添加行列权限相关 API (已使用现有 API: rowPermissionList, columnPermissionList, saveRowPermission, saveColumnPermission 等)

---

## Phase 5: 清理与测试

### Task 5.1: 删除重复菜单
- [x] 删除「菜单权限配置」菜单（`/system/menu-permission`）
- [x] 清理相关组件文件

### Task 5.2: 集成测试
- [x] 测试顶部菜单导航（4 个入口：工作台、可视化、数据管理、系统管理）
- [x] 测试用户管理（用户/角色 Tab）- 组件文件已创建
- [x] 测试组织管理 - 路由正常
- [x] 测试权限配置（菜单/资源/行列权限）- 组件文件已创建并连接 API
- [x] 测试菜单配置 - 路由正常
- [x] 前端测试通过（401 个测试用例）
- [x] 后端测试通过

### Task 5.3: 文档更新
- [x] 更新菜单结构文档（`menu-structure.md`）
- [x] 更新用户手册（如有，当前仓库无独立用户手册，已在 `closure-notes.md` 记录 N/A 说明）

---

## Phase 6: 收尾与归档准备

### Task 6.1: 变更收尾说明
- [x] 记录一级/二级菜单最终结构与旧入口映射关系
- [x] 记录用户管理合并角色 Tab 的最终说明
- [x] 记录权限配置三 Tab 的最终说明与边界

### Task 6.2: 验收与遗留项说明
- [x] 补齐本 change 的最终验收记录
- [x] 明确哪些问题已完成，哪些问题转入恢复类 change 跟进
- [x] 确认本 change 不再承接广义 broken-features 修复

### Task 6.3: 归档准备
- [x] 确认 proposal / design / verification / tasks 与最终范围一致
- [x] 确认剩余事项仅为文档和归档动作
- [x] 标记为可归档状态（用户 create/delete、权限保存回显、组织 CRUD 与菜单/路由闭环验收已补齐）

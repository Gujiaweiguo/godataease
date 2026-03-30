## 1. 后端菜单数据迁移

- [x] 1.1 创建迁移文件 `apps/backend-go/migrations/mysql/20260329_refactor_menu_final.sql`，包含：INSERT 新一级「组织权限」group（pid=0, auth=1, menu_type='group', menu_sort=4, name='commons.org_permission'）；UPDATE 用户管理/组织管理/角色管理/权限管理的 pid 指向新 group；UPDATE 菜单管理 pid 指向系统设置(id=101)；UPDATE 工具箱(id=30) hidden=0, pid=0, menu_sort=6, in_layout=1；DELETE 帮助文档分组及 4 个子菜单；重排所有一级 menu_sort（1=工作台, 2=可视化, 3=数据管理, 4=组织权限, 5=系统设置, 6=工具箱）
- [x] 1.2 在迁移中确认「数据导出中心」菜单 pid 已指向工具箱(id=30)且 hidden=0；如果 pid 不正确则 UPDATE 修正
- [x] 1.3 在迁移中 INSERT sys_role_menu 为管理员角色授权新的「组织权限」菜单
- [x] 1.4 启动本地 dev 环境，执行迁移后调用 `GET /de2api/menu/query` 验证返回的菜单树包含 6 个一级分组且层级正确
- [x] 1.5 验证管理员登录后授权菜单树返回完整 6 组，普通用户只返回工作台/可视化/数据管理/工具箱
  - 已通过 `test / Normal123!` 登录并验证 `/api/roleRouter/query` 仅返回 Workbench、Visualization、Data Preparation、Toolbox 4 个一级菜单；管理员 6 组验证见 1.4、6.1-6.2、6.4、6.6。

## 2. 前端 shell 组件清理

- [x] 2.1 删除 `apps/frontend/src/layout/components/MoreMenu.vue` 文件
- [x] 2.2 修改 `apps/frontend/src/layout/components/Header.vue`，移除 MoreMenu 的 import 语句和模板中的 `<MoreMenu />` 引用及相关样式
- [x] 2.3 修改 `apps/frontend/src/layout/components/AccountOperator.vue`，移除第 95 行 `linkLoaded([{ id: 4, link: '/sys-setting/parameter', label: t('commons.system_setting') }])` 及其条件判断
- [x] 2.4 验证：启动前端 dev server，确认头部不再渲染 More 按钮和帮助链接入口；确认管理员头像下拉不再显示「系统设置」快捷项

## 3. 侧边栏 event 类型菜单点击支持

- [x] 3.1 修改 `apps/frontend/src/layout/components/Menu.vue`（或 `MenuItem.vue`），确保二级菜单项在 `menu_type='event'` 时点击触发 `useEmitt().emitter.emit(actionConfig.event)` 而非 `router.push`
- [x] 3.2 修改 `apps/frontend/src/layout/components/menu-utils.ts`，在菜单点击处理函数中增加 event 类型的分支逻辑（参考现有 `menu-actions.ts` 中 `data-export-center` 的处理方式）
- [x] 3.3 验证：登录后展开侧边栏「工具箱」，点击「数据导出中心」子菜单，确认导出中心抽屉正确弹出且不触发路由跳转

## 4. 国际化与前端构建验证

- [x] 4.1 在 `apps/frontend/src/locales/zh-CN.ts` 添加 `commons.org_permission: '组织权限'`
- [x] 4.2 在 `apps/frontend/src/locales/en.ts` 添加 `commons.org_permission: 'Organization & Permission'`
- [x] 4.3 在 `apps/frontend/src/locales/tw.ts` 添加 `commons.org_permission: '組織權限'`
- [x] 4.4 运行 `npm run lint` 和 `npm run ts:check` 确认无报错
- [x] 4.5 运行 `npm run test:core` 确认核心测试通过

## 5. E2E 测试更新

- [x] 5.1 删除 `apps/frontend/e2e/menu/help-menu.spec.ts`（帮助菜单已移除）；如果 `e2e/menu/` 目录下只剩此文件，删除整个目录
- [x] 5.2 更新 `apps/frontend/e2e/system-management-menu-smoke.spec.ts` 中的导航断言：用户管理/组织管理/角色管理/权限管理 → 在「组织权限」下导航；菜单管理 → 在「系统设置」下导航；确认不再断言帮助菜单相关行为
- [x] 5.3 检查其他 E2E 文件中是否有引用 MoreMenu 帮助入口的断言（如 `more-menu-trigger`、`more-menu-popover` 选择器），如有则移除或更新
- [x] 5.4 运行受影响的 E2E 测试确认通过（或标记需要后端服务的测试为 fixme）

## 6. 集成回归验证

- [x] 6.1 管理员登录后验证侧边栏一级菜单排序为：工作台、可视化、数据管理、组织权限、系统设置、工具箱
- [x] 6.2 验证「组织权限」下展示：用户管理、组织管理、角色管理、权限管理（共 4 项）
- [x] 6.3 验证「系统设置」下展示当前数据库基线中的受治理子项：菜单管理、审计日志、审计看板、审计设置（共 4 项）
  - 已按运行时真实菜单树验证通过；本项同步修订为与迁移文件 `20260329_refactor_menu_final.sql` 和当前 DB 基线一致的子项清单。
- [x] 6.4 验证「工具箱」展开后展示「数据导出中心」，点击后导出中心抽屉正常弹出
- [x] 6.5 验证普通用户登录后只看到：工作台、可视化、数据管理、工具箱（共 4 个一级菜单）
  - 已通过 `test / Normal123!` 登录并在硬刷新后的真实 UI 中确认仅显示 4 个一级菜单，且不显示「组织权限」「系统设置」。
- [x] 6.6 验证头部不再有 More（...）按钮；头像下拉只显示：关于、修改密码、语言、退出系统

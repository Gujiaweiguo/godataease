## 1. API and Data Contract Alignment

- [x] 1.1 新增前端 `api/menu.ts` 并对接 `/api/menu/query|create|update|delete|updateSort|updateHidden|detail`
- [x] 1.2 增加菜单字段校验规则（path/component/icon/sort/hidden）并统一错误提示

## 2. Menu Admin Console

- [x] 2.1 新增 `views/system/menu/index.vue` 菜单管理页（树形列表）
- [x] 2.2 实现菜单新增/编辑/删除弹窗与表单校验
- [x] 2.3 实现菜单排序和隐藏状态更新交互

## 3. Dynamic Navigation Refactor

- [x] 3.1 移除 `permission.ts` 中硬编码过滤逻辑（含 `system` 特判）
- [x] 3.2 改造 `Header.vue`：顶栏菜单改为后端数据驱动排序
- [x] 3.3 改造 `Menu.vue`：侧栏菜单改为同源子树渲染

## 4. Authorization Integration

- [x] 4.1 验证角色-菜单授权变更后导航可见性同步生效
- [x] 4.2 验证未授权路由直接访问拦截行为不回退

## 5. Verification and Rollout

- [x] 5.1 执行前端 `npm run lint` 与 `npm run ts:check`
- [x] 5.2 执行菜单核心路径 E2E（登录->授权->导航渲染）
- [x] 5.3 以 feature flag 灰度发布并验证回滚

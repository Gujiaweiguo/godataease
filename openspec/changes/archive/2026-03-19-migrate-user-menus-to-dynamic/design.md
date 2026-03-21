# Design: Migrate User Menus to Dynamic

## Overview

本文档描述如何将右上角硬编码菜单（用户下拉菜单和帮助菜单）迁移到动态菜单管理系统。

## Data Model Changes

### 1. 扩展 core_menu 表

**新增字段**：

```sql
ALTER table core_menu add column menu_location varchar(50) default 'sidebar';
alter table core_menu add column menu_type varchar(20) default 'route';
alter table core_menu add column action_config json default null;
```

**字段说明**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `menu_location` | varchar(50) | 菜单位置：`top-nav`（顶部导航）、`sidebar`（侧边栏）、`user-dropdown`（用户下拉）、`help-menu`（帮助菜单） |
| `menu_type` | varchar(20) | 菜单类型： `route`（路由跳转）、`event`（触发事件）、`external`（外部链接）、`group`（分组） |
| `action_config` | json | 动作配置： `{"event": "open-about-dialog"}` 或 `{"url": "https://..."}` |

### 2. 初始菜单数据

```sql
-- 用户下拉菜单
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_location, menu_type, action_config) VALUES
(0, 0, 'common.about', '', 1, 'question-circle', '/mine/about', 0, 1, 0, 'user-dropdown', 'event', '{"event": "open-about-dialog"}'),
(0, 0, 'user.change_password', 'system/modify-pwd/index', 2, 'lock', '/modify-pwd/index', 0, 1, 0, 'user-dropdown', 'route', NULL),
(0, 0, 'commons.system_setting', 'sys-setting/parameter/index', 3, 'setting', '/sys-setting/parameter', 0, 1, 0, 'user-dropdown', 'route', NULL);

-- 帮助菜单
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_location, menu_type, action_config) VALUES
(0, 0, 'api_pagination.help_documentation', '', 1, 'book', '/help/doc', 0, 0, 0, 'help-menu', 'external', '{"url": "https://dataease.io/docs/v2/"}'),
(0, 0, 'api_pagination.product_forum', '', 2, 'chat-dot-round', '/help/forum', 0, 0, 0, 'help-menu', 'external', '{"url": "https://bbs.fit2cloud.com/c/de/6"}'),
(0, 0, 'api_pagination.technical_blog', '', 3, 'article', '/help/blog', 0, 0, 0, 'help-menu', 'external', '{"url": "https://blog.fit2cloud.com/categories/dataease"}'),
(0, 0, 'api_pagination.enterprise_edition_trial', '', 4, 'star', '/help/trial', 0, 0, 0, 'help-menu', 'external', '{"url": "https://jinshuju.net/f/TK5TTd"}');
```

## API Changes
### 1. 扩展 /roleRouter/query 响应

**当前响应**：
```json
{
  "path": "/mine/about",
  "name": "mine-about",
  "component": "about/page",
  "hidden": false,
  "meta": { "title": "关于" }
}
```

**扩展后响应**：
```json
{
  "path": "/mine/about",
  "name": "mine-about",
  "component": "about/page",
  "hidden": false,
  "menuLocation": "user-dropdown",
  "menuType": "event",
  "actionConfig": { "event": "open-about-dialog" },
  "meta": { "title": "关于" }
}
```

### 2. 新增 API: /api/menus/by-location

**用途**：按位置获取菜单列表

**请求**：
```json
GET /api/menus/by-location?location=user-dropdown
```

**响应**：
```json
{
  "code": "000000",
  "data": [
    {
      "id": 1,
      "name": "关于",
      "path": "/mine/about",
      "menuType": "event",
      "actionConfig": { "event": "open-about-dialog" },
      "menuSort": 1
    }
  ]
}
```

## Frontend Architecture
### 1. 菜单状态管理

**扩展现有 permissionStore**：

```typescript
// src/store/modules/permission.ts
interface PermissionState {
  routers: AppCustomRouteRecordRaw[]
  addRouters: AppCustomRouteRecordRaw[]
  userMenus: UserMenuItem[]      // 新增：用户下拉菜单
  helpMenus: UserMenuItem[]      // 新增：帮助菜单
  isAddRouters: boolean
}

interface UserMenuItem {
  id: number
  name: string
  path: string
  icon?: string
  menuType: 'route' | 'event' | 'external' | 'group'
  actionConfig?: {
    event?: string
    url?: string
  }
  menuSort: number
}
```

### 2. 菜单事件处理器

**新建文件**： `src/utils/menu-actions.ts`

```typescript
// 事件处理器映射
const eventHandlers: Record<string, () => void> = {
  'open-about-dialog': () => {
    useEmitt().emitter.emit('open-about-dialog')
  },
  'user-logout': () => {
    performLogout()
  },
  // 可扩展更多事件
}

export function executeMenuAction(menuType: string, actionConfig?: any) {
  if (menuType === 'event' && actionConfig?.event) {
    const handler = eventHandlers[actionConfig.event]
    if (handler) handler()
  } else if (menuType === 'external' && actionConfig?.url) {
    window.open(actionConfig.url, '_blank')
  }
}
```

### 3. 组件改造

**AccountOperator.vue 改造要点**：

```vue
<script setup lang="ts">
import { usePermissionStore } from '@/store/modules/permission'
import { executeMenuAction } from '@/utils/menu-actions'

const permissionStore = usePermissionStore()
const userMenus = computed(() => permissionStore.userMenus)
const handleMenuClick = (menu: UserMenuItem) => {
  if (menu.menuType === 'route') {
    router.push(menu.path)
  } else {
    executeMenuAction(menu.menuType, menu.actionConfig)
  }
}
</script>

<template>
  <el-dropdown-item 
    v-for="menu in userMenus" 
    :key="menu.id"
    @click="handleMenuClick(menu)"
  >
    {{ menu.name }}
  </el-dropdown-item>
</template>
```

## Migration Strategy
### Phase 1: 后端扩展（2-3天）

1. **数据库迁移**
   - 创建迁移脚本 `20260319_menu_location_extension.sql`
   - 添加 `menu_location`、`menu_type`、`action_config` 字段
   - 插入初始菜单数据

2. **后端代码修改**
   - 扩展 `CoreMenu` 和 `MenuVO` 结构
   - 扩展 `MenuService` 支持按位置过滤
   - 扩展 `FrontendCompatHandler` 响应

3. **单元测试**
   - 测试菜单按位置查询
   - 测试菜单类型解析

### Phase 2: 前端改造（3-4天）

1. **Store 扩展**
   - 扩展 `permissionStore` 支持用户菜单和帮助菜单
   - 实现菜单加载逻辑

2. **事件处理器**
   - 创建 `menu-actions.ts`
   - 实现事件映射

3. **组件重构**
   - 重构 `AccountOperator.vue`
   - 重构 `MoreMenu.vue`

4. **单元测试**
   - 测试事件处理器
   - 测试菜单加载

### Phase 3: 集成测试（1-2天）

1. **E2E 测试**
   - 测试用户菜单显示和交互
   - 测试帮助菜单链接
   - 测试角色权限过滤

2. **回归测试**
   - 确保现有功能不受影响
   - 验证菜单管理界面正常

## Risk Mitigation
### 1. 数据迁移风险

**风险**：菜单数据迁移失败导致菜单消失
**缓解**：
- 迂移脚本先在测试环境验证
- 提供回滚脚本
- 保留硬编码作为 fallback

### 2. 权限控制风险

**风险**：菜单权限配置不当导致敏感功能暴露
**缓解**：
- 管理员菜单（系统设置、修改密码）默认只分配给 admin 角色
- 前端同时检查 `uid === '1'` 作为双重保护

### 3. 兼容性风险

**风险**：旧版本 API 响应不包含新字段
**缓解**：
- 新字段使用默认值，确保旧数据兼容
- 前端优雅处理缺失字段

## Testing Strategy
### 1. 单元测试

- **后端**：
  - `TestMenuService_QueryByLocation` — 测试按位置查询
  - `TestMenuService_MenuTypeParsing` — 测试菜单类型解析
  - `TestFrontendCompatHandler_ExtendedResponse` — 测试扩展响应

- **前端**：
  - `menu-actions.test.ts` — 测试事件处理器
  - `permission-store.test.ts` — 测试菜单状态管理

### 2. E2E 测试

- `user-menu.spec.ts`：
  - 普通用户看到"关于"、"退出"
  - 管理员用户看到"系统设置"、"修改密码"
  - 点击"关于"打开对话框
  - 点击"退出"执行登出

- `help-menu.spec.ts`：
  - 点击帮助文档打开正确 URL
  - 点击产品论坛打开正确 URL

### 3. 回归测试

- 现有菜单管理功能正常
- 现有顶部导航正常
- 现有侧边栏菜单正常

## Success Metrics
- 所有硬编码菜单项成功迁移到数据库
- 菜单管理界面可以管理用户菜单和帮助菜单
- 角色权限可以控制菜单可见性
- 回归测试 100% 通过
- 代码覆盖率不低于 80%

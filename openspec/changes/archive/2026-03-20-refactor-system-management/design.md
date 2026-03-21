# Design: Refactor System Management

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                              顶部菜单结构                                    │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  工作台   │  │  📊 可视化 ▼ │  │ 📁 数据管理 ▼ │  │ ⚙️ 系统管理 ▼ │          │
│  └──────────┘  └──────────────┘  └──────────────┘  └──────────────┘          │
│                                                                              │
└────────────────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────────────────┐
│                              系统管理子菜单                                  │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  👤 用户管理 ────────────────────────────────────────────────────────────────  │
│  ├── 【用户】Tab                                                              │
│  │   ├── 用户列表                                                             │
│  │   ├── 添加用户、批量导入、编辑、重置密码、删除、查询                         │
│  │   └── 字段：账号、姓名、邮箱、手机、角色、组织、启用状态                     │
│  └── 【角色】Tab                                                              │
│      ├── 角色列表（系统级/组织级/自定义）                                     │
│      ├── 添加用户到角色（组织用户/外部用户）                                  │
│      ├── 移除用户                                                             │
│      └── 创建自定义角色（继承内置角色）                                       │
│                                                                              │
│  🏢 组织管理 ────────────────────────────────────────────────────────────────  │
│  ├── 组织树（多级结构）                                                       │
│  ├── 新建/编辑/删除组织                                                      │
│  └── 字段：组织名称、描述、父组织、层级、状态                                  │
│                                                                              │
│  🔐 权限配置 ────────────────────────────────────────────────────────────────  │
│  ├── 【菜单权限】Tab                                                          │
│  │   └── 给角色分配功能模块（工作台、仪表板、大屏、数据集、数据源、系统管理）    │
│  ├── 【资源权限】Tab                                                          │
│  │   ├── 按用户配置 / 按资源配置（视角切换）                                  │
│  │   ├── 数据源权限（查看/管理/授权）                                         │
│  │   ├── 数据集权限（查看/管理/授权/导出）                                    │
│  │   ├── 仪表板权限（查看/管理/授权/导出）                                    │
│  │   └── 大屏权限（查看/管理/授权/导出）                                      │
│  └── 【行列权限】Tab                                                          │
│      ├── 行权限（按角色/用户/系统变量过滤数据）                               │
│      │   └── 白名单（不受规则限制）                                           │
│      └── 列权限（字段禁用/脱敏）                                              │
│                                                                              │
│  📋 菜单配置 ────────────────────────────────────────────────────────────────  │
│  └── 菜单项 CRUD（后台管理）                                                 │
│                                                                              │
│  ⚙️ 系统参数 ────────────────────────────────────────────────────────────────  │
│  └── 基础设置、地图设置、引擎设置                                            │
│                                                                              │
│  🔤 字体管理 ────────────────────────────────────────────────────────────────  │
│  └── 字体上传、默认字体设置                                                 │
│                                                                              │
└────────────────────────────────────────────────────────────────────────────────┘
```

## Component Design

### 1. 用户管理组件重构

**文件**: `src/views/system/user/index.vue`

```vue
<template>
  <div class="user-management">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="用户" name="user">
        <UserList />
      </el-tab-pane>
      <el-tab-pane label="角色" name="role">
        <RoleList />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
```

**改动点**:
- 添加 Tab 切换组件
- 将原 `views/system/role/index.vue` 重构为 Tab 组件
- 移除角色管理的菜单授权、权限设置功能（移到权限配置）

### 2. 权限配置组件重构

**文件**: `src/views/system/permission/index.vue`

```vue
<template>
  <div class="permission-config">
    <el-tabs v-model="activeTab">
      <el-tab-pane label="菜单权限" name="menu">
        <MenuPermission />
      </el-tab-pane>
      <el-tab-pane label="资源权限" name="resource">
        <ResourcePermission />
      </el-tab-pane>
      <el-tab-pane label="行列权限" name="data">
        <DataPermission />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
```

**改动点**:
- 添加三个 Tab：菜单权限、资源权限、行列权限
- 菜单权限：从角色管理移入，给角色分配菜单
- 资源权限：扩展现有功能，支持按用户/按资源配置视角
- 行列权限：新增功能，支持行权限和列权限配置

### 3. 顶部菜单组件

**文件**: `src/layout/components/Header.vue`

**改动点**:
- 合并仪表板 + 大屏 → 可视化下拉菜单
- 移除工具箱独立菜单
- 移除帮助菜单（移至右上角）

## Data Model Changes

### core_menu 表调整

```sql
-- 1. 创建"可视化"分组菜单
INSERT INTO core_menu (pid, type, name, component, menu_sort, icon, path, hidden, in_layout, auth, menu_type)
VALUES (0, 1, 'commons.visualization', NULL, 2, 'chart', '/visualization', 0, 1, 0, 'group');

-- 2. 移动仪表板、大屏、模板管理到可视化下
UPDATE core_menu SET pid = (SELECT id FROM core_menu WHERE path = '/visualization') 
WHERE path IN ('/panel', '/screen');

-- 3. 移动模板设置到可视化
UPDATE core_menu SET pid = (SELECT id FROM core_menu WHERE path = '/visualization') 
WHERE path = '/template-setting';

-- 4. 隐藏工具箱菜单
UPDATE core_menu SET hidden = 1 WHERE path = '/toolbox';

-- 5. 删除重复的菜单权限配置菜单
DELETE FROM core_menu WHERE path = '/system/menu-permission';

-- 6. 调整菜单排序
UPDATE core_menu SET menu_sort = 1 WHERE path = '/workbranch';
UPDATE core_menu SET menu_sort = 2 WHERE path = '/visualization';
UPDATE core_menu SET menu_sort = 3 WHERE path = '/data';
UPDATE core_menu SET menu_sort = 4 WHERE path = '/system';
```

## API Changes

### 新增 API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/system/permission/menu-tree` | 获取菜单权限树 |
| POST | `/system/permission/menu-auth` | 保存角色菜单授权 |
| GET | `/system/permission/resource-tree` | 获取资源权限树 |
| POST | `/system/permission/resource-auth` | 保存资源权限 |
| GET | `/system/permission/row-column-list` | 获取行列权限列表 |
| POST | `/system/permission/row-column-save` | 保存行列权限 |

### 修改 API

| 方法 | 路径 | 变更 |
|------|------|------|
| GET | `/menu/query` | 返回调整后的菜单树 |

## Migration Path

1. **Phase 1**: 数据库迁移（菜单结构调整）
2. **Phase 2**: 前端菜单组件更新
3. **Phase 3**: 用户管理合并角色 Tab
4. **Phase 4**: 权限配置功能完善
5. **Phase 5**: 集成测试

## Backwards Compatibility

- 保留旧菜单路径重定向（`/panel` → `/visualization/panel`）
- 保留现有角色 API，前端调用不变
- 权限数据迁移保证授权一致

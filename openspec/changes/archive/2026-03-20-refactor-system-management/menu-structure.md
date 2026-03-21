# Menu Structure Summary: Refactor System Management

## 1. Governed Primary Navigation

This change governs the primary functional information architecture as four first-level entries:

1. 工作台 (`/workbranch`)
2. 可视化 (`/visualization`)
3. 数据管理 (`/data`)
4. 系统管理 (`/system`)

Utility entries such as help and personal-center style menus remain outside this primary functional grouping.

## 2. Second-Level Menu Structure

### 可视化
- 仪表板 (`/panel`)
- 数据大屏 (`/screen`)
- 模板市场 (`/template-market`)

### 数据管理
- 数据集 (`/dataset`)
- 数据源 (`/datasource`)
- 数据导出中心 (`/export-center`)

### 系统管理
- 用户管理 (`/system/user`)
- 组织管理 (`/system/org`)
- 权限配置 (`/system/permission`)
- 菜单配置 (`/system/menu`)

Additional system pages such as system parameters and font management remain part of the broader system-management capability surface, but this change's verified restructuring work is centered on the menu grouping above.

## 3. Legacy-to-New Mapping

| Legacy entry | New location | Outcome |
|---|---|---|
| 独立仪表板一级入口 | 可视化 / 仪表板 | Moved into grouped visualization navigation |
| 独立大屏一级入口 | 可视化 / 数据大屏 | Moved into grouped visualization navigation |
| 模板相关入口 | 可视化分组内 | Consolidated under visualization |
| 数据导出中心 | 数据管理 / 数据导出中心 | Moved into data-management grouping |
| 独立角色管理菜单 | 系统管理 / 用户管理 / 角色 Tab | Standalone role menu hidden |
| 菜单权限配置 | 系统管理 / 权限配置 | Duplicate entry removed |
| 工具箱 | 不再作为主导航入口 | Hidden from primary navigation |

## 4. Scope Boundary Notes

- This document describes the **menu-structure outcomes governed by this change**.
- Broader reachability, RBAC recovery, and unrelated broken-feature stabilization are tracked in separate changes.
- This document is intended as the archive-preparation summary for the information-architecture changes made here.

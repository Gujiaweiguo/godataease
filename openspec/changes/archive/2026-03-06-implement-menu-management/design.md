## Context

后端已实现 `core_menu` 持久化与菜单 CRUD、角色菜单授权，但前端仍存在导航渲染硬编码（顶部白名单、桌面端过滤 `system` 等），且缺少菜单管理 UI，导致“配置-授权-渲染”链路不完整。

## Goals / Non-Goals

**Goals:**
- 新增菜单管理控制台，支持菜单树可视化与字段编辑。
- 顶部/左侧导航统一由后端权限菜单树驱动。
- 保持角色-菜单授权模型不变，最小化破坏性改动。

**Non-Goals:**
- 不在本 change 中改造 RBAC 核心表结构。
- 不实现角色/用户管理能力补齐（由 `align-rbac-with-xpack-docs` 负责）。

## Decisions

1. 以 `core_menu` 为菜单单一真源（SSOT）。
   - Why: 已有 schema 与服务完备，改造成本最低。
   - Alternative: 前端静态路由为主；被拒绝，无法治理权限一致性。

2. 顶部与侧栏均基于同一菜单树渲染。
   - Why: 消除顶侧不一致和重复配置。
   - Alternative: 顶侧分别维护；被拒绝，持续产生偏差。

3. 导航渲染遵循后端 `menu_sort/hidden/type`，移除前端硬编码过滤。
   - Why: 菜单管理改动后可立即生效，减少前端发布依赖。
   - Alternative: 前端继续白名单排序；被拒绝，不可维护。

## Risks / Trade-offs

- [菜单配置错误导致路由不可达] -> 后端增加 path/component 校验与冲突提示。
- [菜单授权与导航显示不一致] -> 统一以 `/roleRouter/query` 结果驱动渲染。
- [历史菜单数据质量参差] -> 增加迁移脚本检查（空 path、重复 sort、孤儿节点）。

## Migration Plan

1. 新增前端菜单管理页面与 API 封装。
2. 增量改造 Header/Menu/permission.ts，保留 feature flag 回滚开关（`localStorage['feature.dynamic-navigation']`，默认开启；置为 `false/0/off/disabled` 可回滚旧逻辑）。
3. 完成角色授权联调与 E2E 回归。
4. 灰度发布并监控 403、404 与菜单渲染异常。

## Resolved Decisions

- 菜单删除策略采用“禁止级联删除”：当菜单存在子节点时拒绝删除，管理员必须先清理子节点后再删除父节点。
  - Rationale: 与当前后端 `menu has children, cannot delete` 约束一致，避免误删造成授权与导航结构瞬时断裂。
- 顶部菜单判定规则采用 `pid=0` 根节点语义，不新增显式 `top` 持久化字段。
  - Rationale: 复用现有 `core_menu` 树结构与服务输出，降低数据模型变更风险并保持顶侧导航同源渲染。

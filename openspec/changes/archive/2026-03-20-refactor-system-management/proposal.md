# Refactor System Management

## Motivation

当前系统管理模块存在以下问题：

1. **顶部菜单层级混乱**：7 项平铺菜单，用户心智负担重，缺乏功能域分组
2. **角色管理位置不合理**：角色管理与用户管理分离，但角色本质是用户在组织下的身份
3. **权限配置功能不完整**：缺少菜单权限、资源权限、行列权限的统一入口
4. **存在重复菜单**：「菜单权限配置」与「角色管理 > 菜单授权」功能重复
5. **工具箱入口分散**：模板管理独立于可视化模块

## Capabilities

### New Capabilities

- `top-menu-restructure`: 顶部菜单重组为 4 个入口（工作台、可视化、数据管理、系统管理）
- `unified-permission-config`: 统一权限配置入口（菜单权限 + 资源权限 + 行列权限）
- `role-in-user-management`: 角色管理合并到用户管理模块

### Modified Capabilities

- `user-management`: 增加「角色」Tab，- `permission-config`: 从权限定义 CRUD 扩展为完整的权限配置中心

## Impact

### 前端影响

- `src/layout/components/Header.vue`: 顶部菜单渲染逻辑
- `src/views/system/user/index.vue`: 增加角色 Tab
- `src/views/system/role/index.vue`: 简化为 Tab 组件或移除
- `src/views/system/permission/index.vue`: 扩展为完整权限配置
- `src/views/system/menu/index.vue`: 保留为后台菜单配置

### 后端影响

- 菜单相关 API: `/menu/*`
- 权限相关 API: `/system/permission/*`
- 角色相关 API: `/system/role/*`

### 数据库影响

- `core_menu` 表: 菜单结构调整
- `sys_role_menu` 表: 角色菜单关联调整

## Scope

### In Scope

1. 顶部菜单结构重组
2. 角色管理合并到用户管理
3. 权限配置功能完善
4. 删除重复菜单
5. 工具箱整合到可视化

### Out of Scope

1. 权限底层架构重构（保持现有 RBAC 模型）
2. 数据权限引擎优化
3. 审计日志功能新增

## Success Criteria

1. 顶部菜单显示为 4 个入口，结构清晰
2. 用户管理包含「用户」和「角色」两个 Tab
3. 权限配置包含菜单权限、资源权限、行列权限三个功能区
4. 无重复菜单项
5. 所有现有功能正常工作

## Risks

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 菜单路径变更导致书签失效 | 用户访问 404 | 保留旧路径重定向 |
| 权限配置变更影响现有授权 | 用户权限丢失 | 数据迁移保证授权一致 |
| 前端路由变更 | 开发中功能异常 | 充分测试回归 |

## Timeline

- **Phase 1**: 顶部菜单重组（1 day）
- **Phase 2**: 权限配置完善（1 day）
- **Phase 3**: 角色管理合并（0.5 day）
- **Phase 4**: 集成测试（0.5 day）

**Total**: 3 days

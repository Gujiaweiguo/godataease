## Context

角色管理页面已经实现了前端分页 UI，但后端 `/role/query` 接口只返回全量数据。需要为后端添加分页支持以匹配前端行为。

当前状态：
- 查询： `QueryRoles` 返回 `[]*RoleVO`
- 无分页参数

## Goals / Non-Goals

**Goals:**
- 添加角色分页查询接口
- 支持关键字和角色类型筛选
- 返回分页元数据 (列表、 总数、 当前页、 每页大小)

**Non-Goals:**
- 修改现有 `/role/query` 接口 (保持兼容性)
- 复杂的高级筛选条件

## Decisions
1. **新增独立接口**: 添加 `/role/page` 端点，保持 `/role/query` 不变
2. **仓储层分页**: 通过 `QueryWithPage` 方法在仓储层完成分页与筛选
3. **服务层封装**: `QueryRolesPage` 方法
4. **复用现有模式**: 参考用户管理等其他模块的分页实现

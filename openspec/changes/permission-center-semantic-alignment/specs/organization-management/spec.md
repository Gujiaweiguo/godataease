## ADDED Requirements

### Requirement: Org-Scoped Permission Query Guards

系统 SHALL 为管理员权限查询视图提供组织范围守卫，确保查询结果仅包含属于当前组织的资源。

#### Scenario: Admin permission query filters by org
- **WHEN** 管理员查询权限配置中的资源列表
- **THEN** 系统 SHALL 仅返回属于当前组织范围的资源
- **AND** 其他组织的资源 SHALL 不可见

#### Scenario: Global admin sees all orgs
- **WHEN** 全局管理员（非组织管理员）查询权限配置
- **THEN** 系统 SHALL 返回所有组织的资源（向后兼容）
- **AND** 结果 SHALL 标注每个资源所属的组织ID

#### Scenario: Org admin permission query with invalid org context
- **WHEN** 组织管理员查询权限但组织上下文无效或缺失
- **THEN** 系统 SHALL 返回显式错误
- **AND** 系统 SHALL 不返回任何资源数据

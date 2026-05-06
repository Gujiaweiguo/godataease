## MODIFIED Requirements

### Requirement: Role-based menu authorization remains available from the unified center

系统 SHALL 让管理员从权限配置中心管理角色的菜单授权。角色-权限绑定操作 SHALL 遵守组织范围约束。

#### Scenario: Menu permission tab authorizes governed menu visibility
- **WHEN** 用户打开菜单权限标签页并保存角色-菜单授权变更
- **THEN** 系统 SHALL 通过受治理的权限工作流持久化变更
- **AND** 系统 SHALL 验证目标角色属于当前组织范围
- **AND** 生成的菜单可见性 SHALL 与保存的角色授权状态一致

#### Scenario: Role permission binding rejects cross-org target
- **WHEN** 角色-权限绑定操作尝试绑定属于不同组织的角色和资源
- **THEN** 系统 SHALL 拒绝该操作
- **AND** 系统 SHALL 返回显式组织范围错误

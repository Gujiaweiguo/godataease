# Spec: Dynamic Help Links
## Capability
动态帮助链接管理，支持从数据库配置帮助菜单的外部链接（如帮助文档、产品论坛、技术博客等）。
## Requirements
### Requirement: 导出中心和工具箱保持现有逻辑
导出中心和工具箱功能 SHALL 保持现有业务语义，但其入口位置必须迁移到受治理的一级 Toolbox 菜单及其二级子项中。

#### Scenario: 点击数据导出中心子菜单
- **WHEN** 用户在一级 Toolbox 菜单下点击“数据导出中心”二级菜单
- **THEN** 系统 MUST 触发 `data-export-center` 事件
- **AND** 导出中心的业务行为 MUST 与变更前保持一致

#### Scenario: Toolbox 作为一级导航分组可扩展
- **WHEN** 用户查看受治理一级导航结构
- **THEN** Toolbox MUST 作为独立一级分组存在
- **AND** 后续新增工具项 MUST 通过其子菜单扩展，而不是重新引入头部 More 菜单

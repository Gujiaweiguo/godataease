## REMOVED Requirements

### Requirement: 帮助链接从后端动态加载
**Reason**: 本次导航重构明确删除右上角 More 菜单及其帮助链接入口，帮助链接不再属于当前 shell 的受治理导航面。
**Migration**: 现有帮助文档、产品论坛、技术博客、企业版试用入口应从头部组件与 `core_menu` 配置中移除，后续如需恢复必须通过新的 change 重新定义入口形态。

### Requirement: 帮助链接支持外部 URL
**Reason**: 帮助链接入口整体删除后，不再需要在当前导航中定义外部 URL 打开行为。
**Migration**: 删除对应菜单记录与前端打开逻辑；任何仍需保留的外部链接必须迁移到本 change 范围外的其他载体。

### Requirement: 帮助链接可通过外观配置自定义
**Reason**: 帮助链接入口已不再存在，继续保留 appearance 层自定义要求会制造无效配置面。
**Migration**: 当前 `showDoc` 与帮助文档 URL 配置不再控制头部导航行为；实施时应移除其对 shell 导航的直接影响。

### Requirement: 帮助菜单可见性可通过外观配置控制
**Reason**: More 菜单整体删除后，不存在帮助菜单显示/隐藏语义。
**Migration**: 将帮助菜单可见性逻辑从头部渲染路径中清理，避免隐藏逻辑残留为无用分支。

## MODIFIED Requirements

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

# Spec: Dynamic User Menu

## Capability

动态用户菜单管理，支持从数据库配置用户下拉菜单项，包括路由跳转、事件触发、外部链接等类型。

## Added Requirements
### Requirement: 用户菜单从后端动态加载
用户下拉菜单项应从后端 `/roleRouter/query` API 动态加载，而不是硬编码在前端组件中；但 shell 级语言选择与退出登录仍可保留为受治理的显式账户操作。
#### Scenario: 普通用户查看菜单
- **When** 用户点击右上角头像
- **Then** 系统显示该用户可见的菜单项（如"关于"、"语言"、"退出"）
- **And** 不显示需要管理员权限或已被本次导航重构移除的快捷项
#### Scenario: 管理员用户查看菜单
- **When** 管理员用户点击右上角头像
- **Then** 系统显示本次 change 保留的管理员可见菜单项（包括“修改密码”）
- **And** 不再显示“系统设置”快捷入口，因为系统设置已成为稳定的一级导航入口
### Requirement: 支持多种菜单类型
菜单项支持多种类型：路由跳转、事件触发、外部链接；但本次保留在用户菜单中的受治理项只需要覆盖已有的关于事件和修改密码路由，不再要求系统设置快捷路由。
#### Scenario: 点击路由类型菜单
- **When** 用户点击"修改密码"
- **Then** 系统跳转到 `/modify-pwd/index` 页面
#### Scenario: 点击事件类型菜单
- **When** 用户点击"关于"
- **Then** 系统触发 `open-about-dialog` 事件，打开关于对话框
#### Scenario: 用户菜单不再暴露系统设置快捷路由
- **WHEN** 管理员打开右上角头像菜单
- **THEN** 菜单中 MUST NOT 出现 `/sys-setting/parameter` 对应的系统设置快捷入口
- **AND** 系统设置相关导航 MUST 通过受治理的侧边栏一级菜单访问
### Requirement: 菜单数据包含动作配置
菜单数据应包含 `menuType` 和 `actionConfig` 字段，用于指定菜单的行为。
#### Scenario: 解析菜单动作配置
- **When** 前端接收到菜单数据
- **Then** 系统根据 `menuType` 字段判断菜单类型
- **And** 根据 `actionConfig` 字段执行相应动作（路由跳转、事件触发、打开链接）
### Requirement: 语言选择保持硬编码
语言选择功能由于涉及复杂的 UI 交互，保持硬编码在 `LangSelector` 组件中。
#### Scenario: 用户切换语言
- **When** 用户点击语言选择器
- **Then** 系统显示语言列表
- **And** 用户选择语言后，系统切换语言
### Requirement: 退出登录保持硬编码
退出登录功能由于涉及敏感操作，保持硬编码在 `AccountOperator` 组件中。
#### Scenario: 用户点击退出
- **When** 用户点击"退出系统"
- **Then** 系统调用 `performLogout()` 函数
- **And** 清除本地状态并跳转到登录页

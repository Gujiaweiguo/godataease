# Spec: Dynamic User Menu

## Capability

动态用户菜单管理，支持从数据库配置用户下拉菜单项，包括路由跳转、事件触发、外部链接等类型。

## Added Requirements
### Requirement: 用户菜单从后端动态加载
用户下拉菜单项应从后端 `/roleRouter/query` API 动态加载，而不是硬编码在前端组件中。
#### Scenario: 普通用户查看菜单
- **When** 用户点击右上角头像
- **Then** 系统显示该用户可见的菜单项（如"关于"、"语言"、"退出"）
- **And** 不显示需要管理员权限的菜单项（如"系统设置"、"修改密码"）
#### Scenario: 管理员用户查看菜单
- **When** 管理员用户点击右上角头像
- **Then** 系统显示所有菜单项（包括"系统设置"、"修改密码"）
### Requirement: 支持多种菜单类型
菜单项支持多种类型：路由跳转、事件触发、外部链接。
#### Scenario: 点击路由类型菜单
- **When** 用户点击"修改密码"
- **Then** 系统跳转到 `/modify-pwd/index` 页面
#### Scenario: 点击事件类型菜单
- **When** 用户点击"关于"
- **Then** 系统触发 `open-about-dialog` 事件，打开关于对话框
#### Scenario: 点击外部链接类型菜单
- **When** 用户点击某个外部链接菜单项
- **Then** 系统在新标签页打开指定 URL
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

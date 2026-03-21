# Spec: Dynamic Help Links
## Capability
动态帮助链接管理，支持从数据库配置帮助菜单的外部链接（如帮助文档、产品论坛、技术博客等）。
## Added Requirements
### Requirement: 帮助链接从后端动态加载
帮助菜单的外部链接应从后端 `/roleRouter/query` API 动态加载，而不是硬编码在前端组件中。
#### Scenario: 普通用户查看帮助菜单
- **When** 用户点击右上角"..."按钮
- **Then** 系统显示帮助菜单（包含帮助文档、产品论坛、技术博客、企业版试用等）
- **And** 每个链接都是可以配置的
### Requirement: 帮助链接支持外部 URL
帮助菜单项都是外部链接类型，点击后在新标签页打开指定 URL。
#### Scenario: 点击帮助文档链接
- **When** 用户点击"帮助文档"
- **Then** 系统在新标签页打开 `https://dataease.io/docs/v2/`（或配置的自定义 URL）
#### Scenario: 点击产品论坛链接
- **When** 用户点击"产品论坛"
- **Then** 系统在新标签页打开 `https://bbs.fit2cloud.com/c/de/6`（或配置的自定义 URL）
### Requirement: 帮助链接可通过外观配置自定义
系统管理员可以通过外观配置（Appearance Store）自定义帮助文档的 URL。
#### Scenario: 使用自定义帮助文档 URL
- **When** 系统管理员配置了自定义帮助文档 URL
- **Then** "帮助文档"菜单项使用自定义 URL
- **And** 其他帮助链接保持默认值
### Requirement: 帮助菜单可见性可通过外观配置控制
帮助菜单的显示/隐藏可以通过外观配置（`showDoc` 字段）控制。
#### Scenario: 隐藏帮助菜单
- **When** 系统管理员设置了 `showDoc: false`
- **Then** 帮助菜单不显示
- **And** "..."按钮只显示导出中心和工具箱（如果有权限）
### Requirement: 导出中心和工具箱保持现有逻辑
导出中心和工具箱功能不在本次改动范围内，保持现有实现。
#### Scenario: 点击导出中心
- **When** 用户点击"导出中心"
- **Then** 系统触发 `data-export-center` 事件（保持现有逻辑）
#### Scenario: 点击工具箱
- **When** 用户点击工具箱菜单项
- **Then** 系统跳转到对应工具箱页面（保持现有逻辑）

## MODIFIED Requirements

### Requirement: User API Compatibility

系统 SHALL 保持用户管理 REST API 的完全兼容，并在用户资料引导响应中返回规范化后的语言值供前端建立 locale 状态。

#### Scenario: API response format
- **WHEN** 调用用户管理 API
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的 JSON 响应格式和字段

#### Scenario: Error response format
- **WHEN** 用户操作失败
- **THEN** Go 实现 SHALL 返回与 Java 版本相同的错误码和错误消息格式

#### Scenario: Current user language bootstrap
- **WHEN** 客户端调用当前用户资料接口（例如 `GET /user/info`）以初始化前端状态
- **THEN** 响应中的 `language` 字段 SHALL 返回受支持的规范化 locale 值
- **AND** Go 实现 SHALL NOT 为所有用户返回固定的硬编码默认语言

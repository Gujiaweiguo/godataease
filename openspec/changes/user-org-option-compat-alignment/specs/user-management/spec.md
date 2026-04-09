## MODIFIED Requirements

### Requirement: User API Compatibility

系统 SHALL 保持用户管理 REST API 的完全兼容，并在用户资料引导响应中返回规范化后的语言值供前端建立 locale 状态。

#### Scenario: Legacy `/user/org/option` endpoint returns user-option payload semantics
- **WHEN** a client invokes compatibility endpoint `/user/org/option`
- **THEN** the response MUST preserve user-option payload semantics expected by legacy consumers
- **AND** the endpoint MUST remain aligned with user-option behavior instead of returning organization-list payload semantics

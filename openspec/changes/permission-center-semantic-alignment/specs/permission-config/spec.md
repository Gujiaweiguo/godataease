## MODIFIED Requirements

### Requirement: Permission Service Layer

系统 SHALL 使用 Service 层封装业务逻辑。所有受治理的权限写操作 SHALL 消费组织上下文，拒绝跨组织的权限变更。

#### Scenario: Permission service interface
- **WHEN** 实现权限业务逻辑
- **THEN** 系统 SHALL 提供 PermService 接口，包含 CreatePerm、UpdatePerm、DeletePerm、ListPerms 方法

#### Scenario: Permission key uniqueness
- **WHEN** 创建或更新权限
- **THEN** Service 层 SHALL 验证权限标识（perm_key）唯一性

#### Scenario: Org-scoped permission mutation rejects cross-org target
- **WHEN** 受治理的权限变更操作收到目标资源属于不同组织的上下文
- **THEN** 系统 SHALL 拒绝该操作并返回显式错误
- **AND** 系统 SHALL 记录审计条目记录拒绝原因

#### Scenario: Org-scoped permission mutation with valid org context
- **WHEN** 受治理的权限变更操作收到有效的组织上下文
- **THEN** 系统 SHALL 在该组织范围内执行变更
- **AND** 系统 SHALL 记录审计条目记录操作类型、执行者、目标、组织ID、时间戳

### Requirement: Permission Dual-Perspective Consistency

系统 SHALL 提供按用户和按资源两种视角的权限配置，两种视角 SHALL 持久化到同一个授权模型并产生等效的最终结果。`CheckPermissionConsistency` SHALL 在组织范围内执行交叉验证。

#### Scenario: Consistency check detects no divergence within org scope
- **WHEN** `CheckPermissionConsistency` 在某个组织内调用，且该组织内用户视角和资源视角权限一致
- **THEN** 方法 SHALL 返回 `Consistent: true` 并包含该组织范围内的 `UserCount` 和 `ResourceCount`
- **AND** `Inconsistencies` SHALL 为空

#### Scenario: Consistency check detects divergence within org scope
- **WHEN** `CheckPermissionConsistency` 在某个组织内调用，且存在 `(userID, permKey)` 对在一个视角中存在但另一个视角中不存在
- **THEN** 方法 SHALL 返回 `Consistent: false`
- **AND** `Inconsistencies` SHALL 包含至少一个条目描述差异

#### Scenario: Consistency check only scans within org boundary
- **WHEN** `CheckPermissionConsistency` 收到 orgID 参数
- **THEN** 方法 SHALL 仅扫描属于该组织的用户和资源
- **AND** 其他组织的权限状态 SHALL 不影响结果

#### Scenario: Consistency check on empty org
- **WHEN** `CheckPermissionConsistency` 在一个没有用户、资源或权限的组织内调用
- **THEN** 方法 SHALL 返回 `Consistent: true` 并包含 `UserCount: 0` 和 `ResourceCount: 0`

#### Scenario: Admin bypasses org-scoped consistency check
- **WHEN** 管理员调用 `CheckPermissionConsistency` 不带 orgID 参数
- **THEN** 方法 SHALL 执行全局一致性检查（向后兼容）

## ADDED Requirements

### Requirement: Deferred Permission Dimension Registry

系统 SHALL 提供集中化的延迟权限维度注册表，跟踪所有已标记为延迟的权限维度及其稳定错误码。

#### Scenario: Registry lists all deferred dimensions
- **WHEN** 系统初始化延迟维度注册表
- **THEN** 注册表 SHALL 包含以下条目：`sysParams`（系统变量权限目标）、`whiteList`（行权限白名单）、`dept`（部门授权目标）
- **AND** 每个条目 SHALL 包含维度名称、稳定错误码、人类可读消息

#### Scenario: Service queries registry for deferred dimension rejection
- **WHEN** 权限服务收到涉及延迟维度的请求
- **THEN** 服务 SHALL 从注册表查询稳定错误码和消息
- **AND** 系统 SHALL 返回注册表中定义的错误码和消息（非硬编码字符串）

#### Scenario: Registry returns stable error code for sysParams
- **WHEN** 客户端提交包含 sysParams 目标的权限请求
- **THEN** 系统 SHALL 返回错误码 `DEFERRED_DIMENSION_SYS_PARAMS` 及对应的稳定消息

#### Scenario: Registry returns stable error code for whiteList
- **WHEN** 客户端提交包含 whiteList 的行权限保存请求
- **THEN** 系统 SHALL 返回错误码 `DEFERRED_DIMENSION_WHITELIST` 及对应的稳定消息

#### Scenario: Registry returns stable error code for dept target
- **WHEN** 客户端提交包含 dept 授权目标的权限请求
- **THEN** 系统 SHALL 返回错误码 `DEFERRED_DIMENSION_DEPT` 及对应的稳定消息

### Requirement: Permission Mutation Audit Trail

系统 SHALL 为所有受治理的权限变更操作记录审计条目，使用与 C2 角色生命周期审计一致的模式。

#### Scenario: Permission grant produces audit entry
- **WHEN** 管理员授予用户或角色资源权限
- **THEN** 系统 SHALL 记录审计条目包含：操作类型（GRANT）、执行者ID、目标用户/角色ID、资源ID、组织ID、时间戳

#### Scenario: Permission revocation produces audit entry
- **WHEN** 管理员撤销用户或角色的资源权限
- **THEN** 系统 SHALL 记录审计条目包含：操作类型（REVOKE）、执行者ID、目标用户/角色ID、资源ID、组织ID、时间戳

#### Scenario: Row/column permission save produces audit entry
- **WHEN** 管理员保存行权限或列权限规则
- **THEN** 系统 SHALL 记录审计条目包含：操作类型（SAVE_ROW_PERM 或 SAVE_COLUMN_PERM）、执行者ID、目标数据集ID、组织ID、时间戳

#### Scenario: Audit entries are queryable by org admins
- **WHEN** 组织管理员查询审计日志
- **THEN** 系统 SHALL 返回该组织范围内的所有权限变更审计条目
- **AND** 其他组织的审计条目 SHALL 不可见（除非全局管理员）

### Requirement: Permission Middleware Org-Context Integration

权限中间件 SHALL 从请求中解析组织上下文，验证用户组织成员资格，并在组织上下文不可用时拒绝请求。

#### Scenario: Middleware resolves org context from JWT claims
- **WHEN** 权限中间件处理带有有效 JWT 的请求
- **THEN** 中间件 SHALL 从 JWT claims 中提取组织ID
- **AND** 中间件 SHALL 验证用户属于该组织

#### Scenario: Middleware fails closed when org context unavailable
- **WHEN** 权限中间件无法从请求中解析组织上下文
- **THEN** 中间件 SHALL 返回 403 Forbidden
- **AND** 系统 SHALL 记录审计条目记录失败原因

#### Scenario: Middleware allows admin bypass without org check
- **WHEN** 权限中间件处理管理员用户的请求
- **THEN** 中间件 SHALL 跳过组织成员资格验证
- **AND** 管理员 SHALL 保持跨组织访问能力

#### Scenario: Row permission middleware validates org membership
- **WHEN** `RowPermissionMiddleware` 处理带有数据集ID的请求
- **THEN** 中间件 SHALL 验证数据集属于当前组织
- **AND** 如果数据集不属于当前组织，请求 SHALL 被拒绝

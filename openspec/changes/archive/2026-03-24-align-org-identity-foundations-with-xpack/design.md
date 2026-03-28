## Context

当前仓库已经具备组织、用户、角色、登录等基础模块，也已经有对应的 OpenSpec capability，但这些规范还没有被整理成一个统一的 IAM foundation。官方手册要求后续角色流程和权限配置都建立在稳定的组织边界、成员关系、默认角色和当前身份上下文之上，因此这个 change 需要先把共享主数据和共享查询契约固定下来。

## Goals / Non-Goals

**Goals:**
- 让组织边界、用户归属、当前组织上下文和基础角色模型成为后续 IAM change 的唯一主数据来源。
- 明确登录引导、当前用户资料和组织切换时必须返回的身份上下文字段。
- 保持现有 Go/Vue 模块可复用，优先通过补充契约和服务语义对齐官方手册。

**Non-Goals:**
- 不在本 change 中实现完整角色成员运营流程。
- 不在本 change 中承载统一权限配置中心或资源/行列授权编排。
- 不引入新的外部鉴权框架或整体重写现有表结构。

## Decisions

1. foundation 负责组织、用户、成员关系、默认角色与共享查询接口的唯一落点。
   - Why: Oracle 评审指出，如果这些规则散落到后续 change，角色流程和权限中心会重复定义主数据边界。
   - Alternative: 仅在 foundation 中定义组织和用户，把默认角色延后到 role workflows；被拒绝，因为默认角色决定了后续角色/权限语义的上限。

2. 登录引导与当前用户资料响应必须返回稳定的组织身份上下文。
   - Why: 前端需要在登录后立即建立 current-user、current-org、available-orgs 等状态，否则后续角色页与权限页会在 bootstrap 阶段出现伪 404/伪无权限问题。
   - Alternative: 由前端再拼装组织上下文；被拒绝，因为会造成多接口拼接和语义漂移。

3. 优先保留现有核心实体与服务分层，通过契约对齐补齐行为而不是先做大规模 schema 重构。
   - Why: 仓库已有 `org_service`、`user_service`、`auth_service`、`role_service` 及相应前端模块，增量收敛风险更低。
   - Alternative: 先重新建模 IAM 再迁移；被拒绝，因为会扩大 change 体积并拖慢后续三段式推进。

4. 将“组织切换”和“组织删除后的处置语义”都纳入 foundation 的输入/输出契约。
   - Why: 这两处会被角色流程和权限中心共同消费，属于共享语义而不是页面级细节。
   - Alternative: 在具体页面实现中各自处理；被拒绝，因为无法形成一致的跨 change 验收点。

## Risks / Trade-offs

- [foundation 定义过薄，后续 change 继续补主数据规则] → 在 specs 中显式写出输入依赖、输出契约和共享字段要求。
- [组织上下文字段与现有前端状态模型不一致] → 在 migration plan 中要求先冻结 current-user/bootstrap 契约，再推进页面联动。
- [默认角色模型与现有角色数据冲突] → 先把内置角色定义成行为契约，再在实现阶段增加兼容性迁移与审计验证。
- [组织切换与登录 bootstrap 责任不清] → 在 `login-management` 与 `organization-management` spec delta 中分别写清“身份建立”和“切换行为”。

## Migration Plan

1. 冻结现有组织、用户、登录、基础角色相关接口的响应基线。
2. 先更新 OpenSpec delta，明确 foundation 的共享字段、共享查询和边界语义。
3. 实现阶段优先补后端服务与 handler 契约，再调整前端 store/bootstrap 流程。
4. 在 role workflows 开始前，执行一次 foundation 依赖验收，确认组织上下文和默认角色可被后续复用。
5. 如需回滚，保留旧响应字段兼容层，并以 feature flag 或兼容 handler 切回旧 bootstrap 行为。

## Open Questions

- “移除最后一个角色后的用户处置”是删除、禁用还是禁止操作，最终策略会在 role workflows 中锁定，但 foundation 需要先预留查询语义。
- 组织切换是否要求前端在一次切换后刷新全部授权缓存，还是仅刷新组织上下文与核心菜单。
- 内置角色命名是否完全沿用官方文档，还是保留当前仓库历史命名并通过兼容映射对齐。

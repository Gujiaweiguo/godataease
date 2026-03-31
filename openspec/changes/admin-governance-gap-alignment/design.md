## Context

当前 godataease 的组织、用户、角色、权限相关前后端链路已经基本存在，但存在两类高风险问题：一类是共享语义未锁定，例如组织隔离、最后角色策略、组织删除资源处置策略；另一类是兼容桥与 canonical route 并存，导致“有页面/有接口”不等于“语义已对齐”。本 change 的目标不是重新发明 IAM 模型，而是在保留现有 Go + Vue 架构与主数据模型的前提下，把管理域行为收敛到官方 v2 手册所定义的可验证边界。

该 change 采用 umbrella 方式承载一个单一 Plan v1，并把后续执行组织为三个语义子流：
1. `iam-foundation-semantics`
2. `user-role-lifecycle-alignment`
3. `permission-center-semantic-alignment`

## Goals / Non-Goals

**Goals:**
- 冻结官方基线、差异矩阵、deviation policy 与 compat policy。
- 先修正共享语义和兼容性错配，再推进用户/角色生命周期与权限中心对齐。
- 让后续 OpenSpec 执行任务直接复用 Plan v1 的任务 ID、依赖、验收标准、回滚方案。
- 通过 delta specs 修改已有 capability，而不是并行维护第二套独立规划。

**Non-Goals:**
- 不在本 change 中直接实现所有代码修复。
- 不重构为新的权限引擎或新的租户模型。
- 不复制官方品牌元素、Logo、商业素材或专有文案。
- 不把 UI 重绘、无关重构、体验美化混入该 change。

## Decisions

1. 采用单一 umbrella change 承载 Plan v1，而不是直接创建 3 个并行 change。
   - Why: 用户明确要求基于最新 Plan v1 先创建实际 OpenSpec change 并落 proposal/tasks/specs，且后续统一以 `<ACTUAL_CHANGE_ID>` 引用。
   - Alternative: 直接把 3 个语义 change 各自落盘；被拒绝，因为会导致计划与 change 双重拆分、增加维护面。

2. capability 以“修改已有 specs”为主，不新增独立 capability。
   - Why: 主 specs 已经存在 `organization-management`、`user-management`、`role-management`、`permission-config`、`menu-access-governance`、`api-compatibility-bridge`、`rbac-docs-alignment-governance`。
   - Alternative: 新建 capability；被拒绝，因为会把同一治理问题拆成并行规范，削弱 traceability。

3. 兼容性错配与业务语义偏差分开治理。
   - Why: legacy frontend API、compat handler、canonical route 的问题如果与角色/权限语义混在一起，会让回归验证失真。
   - Alternative: 所有问题一起修；被拒绝，因为无法稳定评估“是 contract 问题还是 semantic 问题”。

4. 保持 Plan v1 的任务 ID、依赖、验收、回滚原样下沉到 tasks.md。
   - Why: 用户要求以 `.sisyphus/plans/*.md` 为唯一规划原稿，不并行维护另一套独立计划。
   - Alternative: 根据 OpenSpec 模板重写任务；被拒绝，因为会引入漂移。

## Risks / Trade-offs

- [文档语义与当前实现存在冲突，如最后角色策略] → 先在 foundation 子流锁定 policy，再进入实现。
- [compat route 过多导致范围膨胀] → 先定义 compat policy，优先 canonical route，shim 最小保留。
- [权限中心涉及菜单/资源/行/列多层语义] → 先菜单/资源，后行/列/P2，防止 P2 阻塞 P0/P1。
- [现有主 spec 与新 delta 可能重复] → delta 仅补充治理要求与行为对齐要求，不复制整个主 spec。

## Migration Plan

1. 先用本 change 固化 proposal、design、spec deltas、tasks。
2. 执行 T1-T4：锁定 baseline、差异矩阵、共享不变量、compat policy。
3. 执行 T5-T6：先收敛 IAM foundation 与 user-role lifecycle。
4. 执行 T7-T8：最后收敛 permission center，并明确 P2 后置边界。
5. 在 F1-F4 通过前，不归档该 change，也不把 deferred/P2 伪装成已完成。

## Open Questions

- “移除最后角色后删除用户”与“禁止移除最后角色”在当前 fork 中最终采用哪一种 policy；本 change 先要求必须显式记录。
- 组织删除时资源处置策略是软删、同步级联还是异步回收；本 change 先要求策略可验证。
- 第三方用户来源字段是否全部落地，还是先保留最小字段与 deferred 说明；本 change 先要求边界明确，不强制一步到位。

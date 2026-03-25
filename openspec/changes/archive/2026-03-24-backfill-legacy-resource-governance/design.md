## Context

`align-permission-center-with-xpack` 已经把统一权限中心收口为菜单、资源、行列权限的共同治理入口，并且完成了“新建且位于已纳管父级下的资源自动继承权限”这条主链路。

但当前治理模型仍然保留一个明确边界：大量历史 datasource、dataset、dashboard、screen 资源并未完整注册进 `sys_resource` / `sys_resource_perm`，因此它们在统一权限中心、按用户/按资源查询、以及运行时鉴权中的身份并不稳定。结果是同一类资源会同时存在“新资源按统一治理模型生效、老资源仍依赖旧语义或类型级语义”的双轨状态。

这说明剩余工作已经不是权限中心页面本身的问题，而是一个资源身份与治理补齐问题：系统需要为历史资源补建稳定、可验证、可审计的治理身份，并让它们与新资源落到同一套有效授权模型上。

## Goals / Non-Goals

**Goals:**
- 为历史 datasource / dataset / dashboard / screen 建立稳定的资源身份补纳管方案。
- 让历史资源在统一权限中心、用户视角、资源视角和运行时鉴权中共享同一有效授权语义。
- 明确历史资源权限继承的补算规则，保证补纳管后行为与新资源一致。
- 为补纳管流程建立幂等、审计、异常跳过和回滚边界，避免一次性迁移破坏已有授权。

**Non-Goals:**
- 不在本 change 中扩展 direct-user grants 的编辑能力。
- 不重新定义角色生命周期、成员关系或组织/身份 bootstrap 语义。
- 不要求一次性修复所有历史脏数据；无法安全归属的异常资源允许跳过并记录。
- 不引入新的权限模型；本 change 只把历史资源补齐到现有统一治理模型。

## Decisions

1. 历史资源补纳管必须复用现有统一资源身份模型，而不是为“老资源”单独发明兼容层。
   - Why: 如果历史资源继续走单独语义，统一权限中心只会变成“部分统一”，后续所有双视角和运行时鉴权都会持续分叉。
   - Alternative: 保留旧逻辑，仅在前端展示层做兼容；被拒绝，因为这会把真实治理差异隐藏成 UI 假象。

2. 补纳管应以“资源身份注册 + 继承补算”两阶段执行，而不是只做查询时的临时映射。
   - Why: 查询时映射只能解决“看得见”，无法解决保存、继承、审计和运行时判定一致性。
   - Alternative: 仅在 `/auth/*` 查询路径中动态推断历史资源；被拒绝，因为无法成为持久治理来源。

3. 无法确定父级归属或治理边界的历史资源允许跳过，但必须产生可追踪记录。
   - Why: 历史数据可能存在缺失父节点、跨组织漂移、孤儿资源等问题，强行补纳管会引入错误授权。
   - Alternative: 对所有历史资源强制补注册；被拒绝，因为错误归属的风险高于短期覆盖率收益。

4. 补纳管结果必须幂等。
   - Why: 该流程很可能需要多次执行、分批执行或在不同环境重复执行；重复运行不能产生重复资源记录或重复授权关系。
   - Alternative: 只支持一次性迁移脚本；被拒绝，因为不利于灰度验证和故障恢复。

5. 继承补算只补齐“当前应生效的治理状态”，不伪造未定义的历史直接授权。
   - Why: 新 change 的目标是对齐统一治理模型，而不是推测历史每一次人工操作意图。
   - Alternative: 根据现存运行结果反推所有直授和继承来源；被拒绝，因为可解释性和可信度不足。


## Resolved Baseline

The following baseline decisions have been frozen and documented in `legacy-resource-governance-baseline.md`:

1. **Natural Key Strategy**: Historical governed identity is frozen as `(resource_type, business_primary_key)` using existing business tables only.

2. **Parent Ownership Validation**: Automatic backfill is frozen to resources with a governable parent path; historical resources without a positive governable parent remain skip/manual-remediation cases.

3. **Organization Boundary Handling**: Only visualization resources expose a safe org boundary (`org_id`) for org-scoped rollout in this change; datasource and dataset do not claim an equivalent automatic org derivation rule.

4. **Anomaly Classification Matrix**: The baseline explicitly covers missing parent, parent-not-governed, cross-org drift, and orphan/remediation-only scenarios.

5. **Auditable Surface**: The minimum auditable surface is the backfill report/output contract returned by the current implementation; this change does not assume a separate persistent migration log.

See `legacy-resource-governance-baseline.md` for the complete frozen baseline.

## Risks / Trade-offs

- [历史资源父级不完整] → 需要显式定义“可自动纳管 / 需跳过记录 / 需人工修复”的分类标准。
- [补纳管后有效权限结果变化] → 必须把 by-user、by-resource 和运行时访问结果纳入同一回归矩阵。
- [重复执行产生脏数据] → 资源注册、权限补算和审计记录都必须按自然键或稳定标识实现幂等。
- [历史直接授权与继承授权混合] → 本 change 只补齐当前统一治理模型可解释的有效状态，不承诺完整追溯所有历史来源。
- [一次性全量迁移风险过高] → 支持按资源类型、组织或批次分段执行，降低回归半径。

## Migration Plan

1. 先冻结历史资源补纳管所需的“稳定身份规则”：资源类型、组织归属、父级归属、自然键或映射规则。
2. 在 `datasource-management`、`dataset-management`、`visualization-management` 与 `permission-config` 中更新 delta specs，明确历史资源进入统一治理后的预期语义。
3. 后端先实现资源注册与继承补算能力，再对齐资源查询和运行时鉴权消费同一套补纳管结果。
4. 前端保持统一权限中心入口不变，只验证历史资源在资源视角、用户视角中的可见性和有效授权一致性。
5. rollout 采用分批策略：先以单一资源类型或受控组织做验证，再扩到全量历史资源。
6. 如需回滚，允许停止补纳管流程并回退新生成的补注册/补算记录，但不得破坏现有已纳管新资源的治理语义。

## Open Questions

- 对于已存在权限效果但缺失 `sys_resource` 记录的历史资源，是否需要保留“来源不可追溯但当前有效”的标记。
- 补纳管审计是否需要升级为独立持久化迁移日志，而不只是当前 backfill report 输出。
- 首轮 rollout 的默认运营策略应优先按资源类型推进，还是在 visualization 内优先按组织维度推进。
- 对于孤儿资源、跨组织异常资源、以及 datasource/dataset 的 org-sensitive 异常，是否需要后续独立 change 提供最小人工修复工具。

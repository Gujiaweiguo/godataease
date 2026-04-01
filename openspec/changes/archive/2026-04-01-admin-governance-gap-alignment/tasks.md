## 1. Research and Gap Analysis

- [x] 1.1 T1 Freeze baseline and gap classification rules
  - Inputs: 用户提供的 4 个官方链接；Oracle/Metis 结论。
  - Outputs: 标准化 gap-audit header（baseline、bucket 定义、证据格式、偏离政策）。
  - Acceptance: 形成 gap matrix 列定义；明确定义三类差异；明确何时允许有意偏离。
  - Rollback: 若 baseline 锁定错误，仅回滚计划规则段落，不进入实现任务。
  - Dependencies: None
  - Risk: High

- [x] 1.2 T2 Produce the admin-domain gap matrix
  - Inputs: T1 输出；官方手册摘要；当前实现映射。
  - Outputs: 可执行 gap matrix，作为后续 change scope 的唯一来源。
  - Acceptance: 覆盖组织/用户/角色/权限四域；每项差异带 manual reference、current evidence、risk、planned action。
  - Rollback: 若归类错误，仅回滚矩阵条目，不推进依赖该条目的 change 切分。
  - Dependencies: T1
  - Risk: High

- [x] 1.3 T3 Define shared invariants and final change split
  - Inputs: T1, T2。
  - Outputs: 共享不变量清单；最终 3-change 结构；对 4-change 的否定理由。
  - Acceptance: 明确最终 3-change 结构；每个 change 的范围、边界、依赖、回滚已定义。
  - Rollback: 若拆分不合理，仅回滚 split/ordering 段落，保留差异矩阵。
  - Dependencies: T1, T2
  - Risk: High

- [x] 1.4 T4 Converge compatibility policy before implementation
  - Inputs: T1, T2。
  - Outputs: compat policy（shim 保留、前端迁移、双支持过渡的明确清单）。
  - Acceptance: 覆盖 user legacy routes、permission compat routes、menu target routes；compat 与 semantic 明确分离。
  - Rollback: 如兼容策略引发范围扩大，回滚为“前端迁移优先，shim 最小保留”的保守方案。
  - Dependencies: T1, T2
  - Risk: High

## 2. IAM Foundation Alignment

- [x] 2.1 T5 Correct IAM foundation semantics
  - Inputs: T2, T3, T4。
  - Outputs: C1 执行清单与验证顺序。
  - Acceptance: 至少包含 org isolation、last-role policy、org delete policy 三条 foundation invariant；每条 invariant 映射到后端测试与至少一条前端/接口验证路径。
  - Rollback: 若 foundation 语义无法一次收敛，拆分为 policy lock 与 enforcement 两步，先保留 policy + failing tests。
  - Dependencies: T2, T3, T4
  - Risk: High

## 3. User and Role Lifecycle Alignment

- [x] 3.1 T6 Align user-role lifecycle
  - Inputs: T2, T4, T5。
  - Outputs: C2 执行清单与验证顺序。
  - Acceptance: 用户组织选择、角色多选、启停、重置密码、导入/错误报告都有明确执行顺序；明确区分修正现有实现与新增 P2 字段/来源支持；`role_service.go:506` 的权限边界 TODO 被纳入 scope。
  - Rollback: 若 P2 来源字段范围过大，回滚为“保留字段占位 + 文档化 deferred implementation”，不阻塞 P0/P1 生命周期对齐。
  - Dependencies: T2, T4, T5
  - Risk: High

## 4. Permission Center Alignment

- [x] 4.1 T7 Align menu and resource permission semantics
  - Inputs: T2, T5, T6。
  - Outputs: C3 前半段（菜单/资源）执行清单与验证顺序。
  - Acceptance: 菜单权限角色绑定、资源权限层级、双视角一致性均明确；`menuTargetPerApi` / `saveMenuTargetPerApi` 被归类并纳入处理；menu auth enforcement stub 被列为修正优先项。
  - Rollback: 若 target 通路补齐风险过大，先回滚到 canonical-only 路线，并将 compat UI 标记为后置。
  - Dependencies: T2, T5, T6
  - Risk: High

- [x] 4.2 T8 Align row/column semantics and contain P2 expansion
  - Inputs: T2, T5, T6, T7。
  - Outputs: C3 后半段（行/列/P2）执行清单、分批次策略与回滚方案。
  - Acceptance: 行权限 stub 与白名单/系统变量差距被纳入处理；列权限规则覆盖差距被纳入处理；P2 项被标记为后置子流且不阻塞 P0/P1。
  - Rollback: 若 P2 子流超出范围，回滚到“仅输出 deferred list + compatibility note”，保留 P0/P1 已完成部分。
  - Dependencies: T2, T5, T6, T7
  - Risk: High

## 5. Final Verification and Scope Fidelity

- [x] 5.1 F1 Plan compliance audit
  - Inputs: 全部执行结果、gap matrix、3 个 change 产出。
  - Outputs: 对照 Plan v1 的合规审计结论。
  - Acceptance: 所有已执行项均能映射回本任务清单或相应 change；不存在未记录的范围扩张。
  - Rollback: 若发现偏离，回滚对应偏离提交并回到最近通过的 change 边界。
  - Dependencies: T8
  - Risk: Medium

- [x] 5.2 F2 Code quality review
  - Inputs: 变更代码、测试结果、lint/typecheck/integration 输出。
  - Outputs: 质量审查结论与缺陷列表。
  - Acceptance: Frontend `npm run lint` 与 `npm run ts:check` 通过；Backend `make test` 通过；涉及持久化时 `make test-integration` 通过。
  - Rollback: 若质量门禁失败，回滚最近一批语义修正，不进入 F3/F4 通过态。
  - Dependencies: T8
  - Risk: Medium

- [x] 5.3 F3 Real manual QA via executable regression flows
  - Inputs: 可运行前后端环境、Playwright 用例、关键管理流程。
  - Outputs: 关键路径验证记录。
  - Acceptance: `e2e/system-user-init.spec.ts`、`e2e/system-role-tab.spec.ts`、`e2e/permission-menu-echo.spec.ts`、`e2e/permission-resource-echo.spec.ts`、`e2e/permission-data-smoke.spec.ts` 均被执行并通过。
  - Rollback: 若关键流程失败，仅回滚对应子流，不整体回滚 foundation。
  - Dependencies: T8
  - Risk: Medium

- [x] 5.4 F4 Scope fidelity check
  - Inputs: 最终交付物、计划范围、deviation policy。
  - Outputs: 范围忠实性结论（IN/OUT、deferred、intentional deviations）。
  - Acceptance: 所有 deferred/P2 项均有明确记录；所有 intentional deviations 均被显式标注。
  - Rollback: 若出现越界实现，回滚越界部分并恢复到计划定义的 IN/OUT 边界。
  - Dependencies: T8
  - Risk: Medium

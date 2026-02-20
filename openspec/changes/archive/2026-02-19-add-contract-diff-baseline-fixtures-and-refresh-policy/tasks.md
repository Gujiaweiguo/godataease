# Plan v1: Baseline Fixtures & Refresh Policy（唯一执行计划）

> 执行基线：本文件是 `add-contract-diff-baseline-fixtures-and-refresh-policy` 的唯一执行计划。任何执行系统仅可依据本文件推进，不允许额外独立计划。

## Dependency Graph

`BASELINE-001 -> BASELINE-002 -> BASELINE-003 -> BASELINE-004 -> {BASELINE-005, BASELINE-006} -> BASELINE-007`

## Parallel Execution Waves

- **Wave 1（规范冻结）**: BASELINE-001
- **Wave 2（目录与命名）**: BASELINE-002
- **Wave 3（增量刷新流程）**: BASELINE-003
- **Wave 4（审阅与门禁）**: BASELINE-004
- **Wave 5（并行完善）**: BASELINE-005 + BASELINE-006
- **Wave 6（联调验收）**: BASELINE-007

## Task List

- [x] **BASELINE-001** Baseline 输入输出契约冻结
  - **Risk**: Medium
  - **Depends On**: None
  - **Input**:
    - `backend-go/testdata/contract-diff/output-schema.md`
    - `backend-go/testdata/contract-diff/critical-whitelist.yaml`
    - `backend-go/testdata/contract-diff/baseline-policy.md`
  - **Output**:
    - baseline fixture 最小字段契约（接口标识、请求上下文、标准响应、版本元数据）
    - 与 runtime engine 的 baseline 消费契约映射
  - **Acceptance Criteria**:
    - 明确 fixture 必填字段与路径规范
    - 明确 baseline 与 diff 报告字段的对应关系
  - **Rollback Plan**:
    - 回退到上版契约定义并暂停后续任务

- [x] **BASELINE-002** 按接口基线目录与命名策略落地
  - **Risk**: Medium
  - **Depends On**: BASELINE-001
  - **Input**:
    - 契约冻结结果
    - whitelist 接口清单
  - **Output**:
    - baseline 目录结构（按 capability/endpoint/method 组织）
    - baseline 命名规则（含接口、版本、时间戳/提交信息）
  - **Acceptance Criteria**:
    - 任意 baseline 文件可反向定位到唯一接口
    - 新增接口 baseline 路径可由规则自动推导
  - **Rollback Plan**:
    - 回退到旧目录布局并提供映射表

- [x] **BASELINE-003** 增量刷新流程实现（dry-run + apply）
  - **Risk**: High
  - **Depends On**: BASELINE-002
  - **Input**:
    - baseline 目录与命名策略
    - runtime engine 输出
  - **Output**:
    - 增量刷新流程：`dry-run` 预览、差异确认、`apply` 落盘
    - 仅刷新受影响接口的策略与过滤规则
  - **Acceptance Criteria**:
    - dry-run 不改文件，仅产出差异清单
    - apply 仅更新目标接口 baseline，不污染无关接口
  - **Rollback Plan**:
    - 切换为只读模式（仅预览不落盘）

- [x] **BASELINE-004** 审阅规则与准入门禁固化
  - **Risk**: High
  - **Depends On**: BASELINE-003
  - **Input**:
    - 增量刷新流程
    - ownership 与签收角色
  - **Output**:
    - baseline 变更审阅清单（owner、tech lead、QA）
    - No-Go 条件（未审批 baseline 不得进入主干）
  - **Acceptance Criteria**:
    - baseline 变更 PR 可识别审阅责任人
    - 未通过审阅时 gate 不能以 baseline 更新为由绕过失败
  - **Rollback Plan**:
    - 恢复为人工审批告警模式并保留审计日志

- [x] **BASELINE-005** 回滚策略与应急手册落地
  - **Risk**: Medium
  - **Depends On**: BASELINE-004
  - **Input**:
    - baseline 版本记录
    - 失败样本与误报样本
  - **Output**:
    - 快速回滚步骤（指定版本回切 + 验证命令）
    - 回滚后验证与证据归档要求
  - **Acceptance Criteria**:
    - 可在一次回滚操作内恢复到上一个稳定 baseline
    - 回滚记录可追溯到触发原因、操作人、验证结果
  - **Rollback Plan**:
    - 若回滚失败，切换到固定稳定 tag 的基线快照

- [x] **BASELINE-006** 漂移监测与误报/漏报防护规则
  - **Risk**: Medium
  - **Depends On**: BASELINE-004
  - **Input**:
    - 增量刷新差异样本
    - 失败 taxonomy
  - **Output**:
    - 漂移检测规则（异常刷新频率、异常字段变化）
    - 误报/漏报判定准则与处置路径
  - **Acceptance Criteria**:
    - 可识别高风险漂移模式并触发人工复核
    - 误报/漏报案例有统一归档和复盘入口
  - **Rollback Plan**:
    - 降级为仅记录漂移告警，不阻断主流程

- [x] **BASELINE-007** CI 联调验收与策略生效确认
  - **Risk**: Medium
  - **Depends On**: BASELINE-005, BASELINE-006
  - **Input**:
    - baseline 管理机制全链路
    - 至少一组刷新成功样本 + 一组回滚样本
  - **Output**:
    - 联调验收记录（刷新、审阅、回滚、归档）
    - 生效确认清单（门禁与流程一致性）
  - **Acceptance Criteria**:
    - 增量刷新、审阅、回滚均可在 CI 中复现
    - 漂移规则可有效减少误报/漏报
    - `openspec validate add-contract-diff-baseline-fixtures-and-refresh-policy --strict --no-interactive` 通过
  - **Rollback Plan**:
    - 暂停 baseline 自动刷新，仅保留稳定基线与手动审批

## Execution Notes

- BASELINE-003/004 是关键路径，失败将阻断后续所有任务。
- baseline 变更必须与审阅结果绑定，禁止“先改后补审”。
- 若刷新流程或命名规则调整，需同步更新相关消费脚本与文档。

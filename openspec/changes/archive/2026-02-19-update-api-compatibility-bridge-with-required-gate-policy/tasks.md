# Plan v1: Required Gate Policy for API Compatibility Bridge（唯一执行计划）

> 执行基线：本文件是 `update-api-compatibility-bridge-with-required-gate-policy` 的唯一执行计划。任何执行系统仅可依据本文件推进，不允许额外独立计划。

## Dependency Graph

`POLICY-001 -> POLICY-002 -> POLICY-003 -> POLICY-004 -> {POLICY-005, POLICY-006} -> POLICY-007`

## Parallel Execution Waves

- **Wave 1（范围冻结）**: POLICY-001
- **Wave 2（强制门禁条款）**: POLICY-002
- **Wave 3（例外审批机制）**: POLICY-003
- **Wave 4（豁免时限机制）**: POLICY-004
- **Wave 5（并行收敛）**: POLICY-005 + POLICY-006
- **Wave 6（验收与生效）**: POLICY-007

## Task List

- [x] **POLICY-001** 必过 Gate 接口范围与来源冻结
  - **Risk**: High
  - **Depends On**: None
  - **Input**:
    - 现有 whitelist / critical API 清单
    - 已落地 gate 能力范围
  - **Output**:
    - 必过接口定义（范围、优先级、变更准入）
    - 接口来源与责任归属规则
  - **Acceptance Criteria**:
    - 必过接口范围可追溯且可审计
    - 范围变更有明确审批入口
  - **Rollback Plan**:
    - 回退到上版必过接口清单并恢复旧策略

- [x] **POLICY-002** 发布前门禁强制条款写入 capability spec
  - **Risk**: High
  - **Depends On**: POLICY-001
  - **Input**:
    - 必过接口范围定义
    - 发布流程门禁现状
  - **Output**:
    - 发布前 gate 必过条款（失败即阻断）
    - 不满足条件禁止发布的规范表述
  - **Acceptance Criteria**:
    - spec 明确“未过 gate 不得发布”
    - 条款可直接映射到 CI/发布流程检查点
  - **Rollback Plan**:
    - 回退为告警策略并保留阻断建议

- [x] **POLICY-003** 例外审批机制固化
  - **Risk**: High
  - **Depends On**: POLICY-002
  - **Input**:
    - 发布门禁强制条款
    - 组织审批角色（owner/tech lead/release）
  - **Output**:
    - 例外申请条件、审批链路、证据要求
    - 未审批例外不可生效的强制约束
  - **Acceptance Criteria**:
    - 所有例外必须绑定审批记录与业务理由
    - 例外流程不可绕开必需审批角色
  - **Rollback Plan**:
    - 回退为仅记录例外，不允许豁免执行

- [x] **POLICY-004** 豁免时限与到期失效策略
  - **Risk**: High
  - **Depends On**: POLICY-003
  - **Input**:
    - 例外审批机制
  - **Output**:
    - 豁免生效时长上限、自动到期规则、续期条件
    - 到期未续期自动恢复阻断的条款
  - **Acceptance Criteria**:
    - 豁免必须具备明确失效时间
    - 到期行为（失效/重审）可机读判定
  - **Rollback Plan**:
    - 回退为最短有效期策略并强制人工复审

- [x] **POLICY-005** 审计与追踪要求补全
  - **Risk**: Medium
  - **Depends On**: POLICY-004
  - **Input**:
    - 例外/豁免流程定义
  - **Output**:
    - 审计最小字段（接口、版本、审批人、原因、有效期）
    - 证据归档与查询要求
  - **Acceptance Criteria**:
    - 任意例外可追溯到完整审批与生效记录
    - 审计字段可支持发布后追责与复盘
  - **Rollback Plan**:
    - 降级为最小审计字段但保留关键追踪信息

- [x] **POLICY-006** 防绕过条款（No-Go）固化
  - **Risk**: Medium
  - **Depends On**: POLICY-004
  - **Input**:
    - 门禁、例外、豁免规则
  - **Output**:
    - 明确绕过判定（未审批豁免、超期豁免、缺证据发布）
    - No-Go 阻断条件清单
  - **Acceptance Criteria**:
    - 任何绕过条件命中均可触发阻断
    - No-Go 条款与发布门禁一致
  - **Rollback Plan**:
    - 回退为观察模式并保留 No-Go 报告

- [x] **POLICY-007** 策略验收与生效确认
  - **Risk**: Medium
  - **Depends On**: POLICY-005, POLICY-006
  - **Input**:
    - 完整策略条款
  - **Output**:
    - 生效确认记录（正常发布样本 + 例外样本 + 到期样本）
    - 规则一致性结论（spec 与流程映射）
  - **Acceptance Criteria**:
    - 发布前门禁、例外审批、豁免时限三者闭环成立
    - 策略不可绕过且可审计
    - `openspec validate update-api-compatibility-bridge-with-required-gate-policy --strict --no-interactive` 通过
  - **Rollback Plan**:
    - 暂停策略生效，仅保留评估报告

## Execution Notes

- POLICY-002/003/004 为核心路径，任一失败将阻断后续任务。
- “例外可用”不等于“永久豁免”，必须受时限约束与到期机制控制。
- 如发布流程实现细节变化，必须保持与 capability spec 条款一致。

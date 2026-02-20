# Change: 将必过 Gate 接口策略固化到 API Compatibility Bridge

## Plan Version
Plan v1（唯一执行计划，执行以本变更 `tasks.md` 为准）

## Why
当前“哪些接口必须过 gate”主要分散在执行习惯与临时清单中，缺少 capability spec 强制条款。发布前若无制度化约束，容易出现绕过门禁、例外长期不收敛、责任不清晰的问题。

## What Changes
- 在 `api-compatibility-bridge` capability 中增加“必过 gate 接口策略”强制要求。
- 固化发布前门禁：未满足必过接口 gate 结果时禁止进入发布流程。
- 增加例外审批机制：例外必须审批、必须记录原因、必须限定有效期。
- 增加豁免时限与到期策略：过期自动失效，触发重新评审或恢复阻断。

## Deliverables
- Required Gate Policy
  - 明确必过接口范围、来源、变更流程
- Pre-Release Gate Rule
  - 发布前强制校验与阻断条件
- Exception & Waiver Governance
  - 例外审批、证据记录、豁免时限与续期规则

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected process/config (expected):
  - `openspec/changes/update-api-compatibility-bridge-with-required-gate-policy/specs/api-compatibility-bridge/spec.md`
  - CI 发布流程策略配置（如需与 spec 对齐）

## Execution Policy
本变更只允许一个执行计划：`openspec/changes/update-api-compatibility-bridge-with-required-gate-policy/tasks.md`。
不生成、不维护任何独立计划文档。

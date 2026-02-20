# Change: 实现 Go Contract Diff Runtime Engine

## Plan Version
Plan v1（唯一执行计划，执行以本变更 `tasks.md` 为准）

## Why
当前 `backend-go/scripts/contract-diff/run_contract_diff.sh` 仍是占位实现，无法稳定完成 Java/Go 契约比对与失败分级判定。CI 已接入 gate 触发链路，但缺少可用于真实阻断的运行时引擎，导致门禁具备形式而不具备回归拦截能力。

## What Changes
- 将 contract diff 脚本从 stub 升级为可执行的运行时引擎，支持 whitelist 驱动比对。
- 实现并发请求执行、超时控制、稳定重试（含退避）与确定性失败输出。
- 输出结构化 diff（`status/code/msg/payload schema/payload value`）与失败分级结果。
- 保持与现有 CI gate、阈值配置、报告归档协议兼容，确保可直接用于阻断。

## Deliverables
- Runtime Engine
  - `run_contract_diff.sh` 具备真实比对执行能力（非占位逻辑）
- Whitelist-driven Execution
  - 基于 `critical-whitelist.yaml` 读取并筛选待比对接口集
- Structured Diff
  - 机器可读 JSON + 人类可读 Markdown，包含差异分类与证据
- Stable Failure Semantics
  - 明确退出码语义，支持 CI 稳定阻断与重试后定性

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected code (expected):
  - `backend-go/scripts/contract-diff/run_contract_diff.sh`
  - `backend-go/testdata/contract-diff/output-schema.md`（必要时补充字段说明）
  - `backend-go/testdata/contract-diff/failure-taxonomy.md`（必要时补充分类映射）
  - `.github/workflows/go-contract-diff-gate.yml`（仅当参数/输出路径需对齐时）

## Execution Policy
本变更只允许一个执行计划：`openspec/changes/implement-go-contract-diff-runtime-engine/tasks.md`。
不生成、不维护任何独立计划文档。

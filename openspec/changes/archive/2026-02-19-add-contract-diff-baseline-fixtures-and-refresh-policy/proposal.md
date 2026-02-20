# Change: 建立 Contract Diff Baseline Fixtures 与增量刷新策略

## Plan Version
Plan v1（唯一执行计划，执行以本变更 `tasks.md` 为准）

## Why
当前 contract diff 对比缺少稳定、可审计、可回滚的 baseline 机制。基准样本随环境和接口演进漂移后，容易出现误报或漏报，导致 CI gate 可信度下降、回归定位成本上升。

## What Changes
- 建立按接口维度的 baseline fixtures 目录规范与命名策略（可追溯到接口与版本）。
- 定义增量刷新流程：仅刷新变更接口，支持 dry-run、差异预览、审批后落盘。
- 建立审阅规则：owner 责任、必审项、禁止绕过的门禁条件。
- 建立回滚策略：快速回退到上一版稳定 baseline 并保留审计证据。
- 与现有 `go-contract-diff-gate` / runtime engine 对齐，避免破坏既有门禁链路。

## Deliverables
- Baseline Fixtures Governance
  - baseline 文件结构、命名、元数据规范
- Incremental Refresh Workflow
  - 刷新命令、预览报告、审批落盘策略
- Review Policy
  - 变更准入、签收责任、No-Go 条件
- Rollback Policy
  - 回滚触发条件、回滚步骤、验证与证据保留

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected code/config (expected):
  - `backend-go/testdata/contract-diff/baselines/**`
  - `backend-go/testdata/contract-diff/baseline-policy.md`
  - `backend-go/testdata/contract-diff/report-archive-policy.md`（如需对齐刷新报告保留策略）
  - `backend-go/scripts/contract-diff/*`（如需新增 baseline 刷新脚本）

## Execution Policy
本变更只允许一个执行计划：`openspec/changes/add-contract-diff-baseline-fixtures-and-refresh-policy/tasks.md`。
不生成、不维护任何独立计划文档。

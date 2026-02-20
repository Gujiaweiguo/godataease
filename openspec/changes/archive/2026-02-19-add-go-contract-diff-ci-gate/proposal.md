# Change: 将 Java/Go 关键接口 contract diff 接入 CI Gate

## Plan Version
Plan v1（唯一执行计划，执行以本变更 `tasks.md` 为准）

## Why
当前 Java/Go 接口契约对齐主要依赖阶段性回归与人工检查，缺少稳定的 CI 阻断机制。随着 Go 迁移范围扩大，关键接口的细微契约偏差（状态码、`code/msg/data` 语义、字段缺失）可能在合并后才暴露，导致回归成本和上线风险增加。

## What Changes
- 新增 Java/Go contract diff CI job，在 PR 与主干合并前自动执行关键接口契约比对。
- 建立并版本化关键接口白名单（whitelist），明确纳入 gate 的接口范围、归属与变更流程。
- 定义失败阈值策略（整体通过率、阻断级差异、必过接口清单）并接入 CI 失败判定。
- 统一产出并归档 contract diff 报告（机器可读 + 人类可读），用于审计、回溯和趋势对比。

## Deliverables
- CI Gate Job
  - 在 GitHub Actions 中新增/扩展 contract diff gate job（PR + protected branch）
- 白名单接口集
  - 提供版本化 whitelist（含 `path/method/owner/priority/blockingLevel`）
- 失败阈值策略
  - 默认阈值建议：`overallParity >= 99%`、`requiredApiPassRate = 100%`、`blockingLevel=critical` 零容忍
- 报告归档
  - 每次执行产出 `JSON + Markdown` 报告，按 `commit/PR/timestamp` 可追溯命名并归档为 artifact

## Impact
- Affected specs:
  - `specs/api-compatibility-bridge/spec.md`
- Affected code (expected):
  - `.github/workflows/go-backend.yml` 或新增独立 workflow（如 `.github/workflows/go-contract-diff-gate.yml`）
  - `backend-go` 下 contract diff 执行脚本与阈值配置（新增）
  - 关键接口白名单与报告目录（新增）
- Operational impact:
  - CI 时长增加（可通过白名单分层与并发策略控制）
  - PR 质量门禁增强，阻断高风险契约回归

## Execution Policy
本变更只允许一个执行计划：`openspec/changes/add-go-contract-diff-ci-gate/tasks.md`。
不生成、不维护任何独立计划文档。

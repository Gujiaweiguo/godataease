# 前端周质量基线

本文档用于统一记录前端每周质量基线，便于趋势跟踪与回归预警。

## 目标

- 用固定字段记录每周前端质量信号。
- 在 PR/CI 阶段尽早发现回归，避免问题进入 `main`。

## 数据来源

- `apps/frontend/scripts/quality-report.cjs` 生成的 CI Summary。
- 工作流：`.github/workflows/frontend.yml`（`quality` job 的 Summary 步骤）。

## 固定字段

- Date (UTC)
- Branch
- Commit
- TypeScript
- ESLint
- Tests (core)
- Build
- Datasource benchmark
- Datasource benchmark compare
- Affected target
- Affected tests
- Workflow run

## 周记录模板

每周追加一行：

| Week | Date (UTC) | Branch | Commit | TS | ESLint | Tests(core) | Build | DS Bench | DS Bench Compare | Affected target | Affected tests | Workflow run | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-W10 | 2026-03-02 | main | 7170ddc | pass | pass | pass | pass | n/a | n/a | test:core | pass | https://github.com/Gujiaweiguo/godataease/actions/runs/<run-id> | 基线初始化 |

## 执行步骤

1. 选择当周一个代表性前端 PR（或 main 合并）并确认 CI 全绿。
2. 打开该次 workflow 的 Summary，定位 `Weekly Baseline Snapshot`。
3. 将字段值复制到上表并追加新行。
4. 若字段不为 `pass`，在 `Notes` 记录原因与修复动作。

## 记录节奏

- 常规：每周至少记录一次。
- 额外：发生大规模前端重构或 CI 规则变更时，补充记录一次。

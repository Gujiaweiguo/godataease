# 前端数据源性能门禁 Runbook

本文档用于维护数据源树性能门禁的日常操作，包括 PR 阻断、nightly 观测、告警配置和阈值复盘。

## 1. 当前门禁策略

- PR 流程：`.github/workflows/frontend.yml`
  - 变更命中 `datasource/shared` 时执行：
    1) `bench:datasource-tree`
    2) `benchmark-compare.cjs`
  - 阻断规则：`threshold=10`，`ignore-below-ms=0.5`，`fail-on-regression=true`
- Nightly 流程：`.github/workflows/frontend-benchmark-nightly.yml`
  - 每日 UTC 18:00 执行，默认只观测不阻断
  - 产物路径：`apps/frontend/benchmark-results/nightly/<STAMP>/`

## 2. 告警配置（Nightly 失败通知）

Nightly 失败时支持两种可选告警：

- Slack: `FRONTEND_BENCHMARK_SLACK_WEBHOOK_URL`
- 企微机器人: `FRONTEND_BENCHMARK_WECOM_WEBHOOK_URL`

配置方法：

1. 打开仓库 Settings -> Secrets and variables -> Actions
2. 新增以上 Secret（至少配置一个）
3. 手动触发 `Frontend Datasource Benchmark Nightly` 验证告警链路

## 3. 基线更新流程

当 nightly 连续 1-2 周稳定后，可按以下流程更新基线文件：

1. 从 nightly artifact 中挑选稳定样本（避免异常值当日）
2. 更新 `apps/frontend/tests/perf/baseline/datasource-tree-baseline.json`
3. 在 PR 中附上对比说明（更新前后趋势与原因）
4. 合并后观察下一周是否出现误报

## 4. 1-2 周观测与阈值复盘模板

每次 nightly 追加一行：

| Date (UTC) | Run URL | Search Trend | Expand Trend | Prune Trend | Regressed Count | Action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-03-05 | <run-url> | stable | stable | improved | 0 | keep |

阈值复盘建议：

- 若 7 天内 `Regressed Count` 全为 0，可考虑将阈值从 10 收紧到 8
- 若误报频繁且集中在小耗时指标，优先提高 `ignore-below-ms`（例如 0.7）
- 若仅某一指标波动大，改为分项阈值（search/expand/prune 各自阈值）

## 5. Search 路径下一轮优化计划

目标：提升 `searchOptimizedMedianMs` 的稳定收益，避免仅持平。

优先级顺序：

1. 增加 `name` 预索引构建时机控制（仅树变更时重建）
2. 对关键字前缀搜索增加短路策略（空结果快速返回）
3. 对大树场景评估分层过滤（先目录再叶子）

验收标准：

- 5k 节点 `searchOptimizedMedianMs` 相对基线提升 >= 5%
- PR compare 连续 3 次稳定 `regressed=0`
- 不引入 lint/ts/test/build 回归

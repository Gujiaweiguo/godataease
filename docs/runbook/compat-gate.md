# 兼容门禁与回滚 Runbook

本文档用于 DataEase Go 主线的兼容门禁巡检、发布后回归与应急回滚。

## 适用范围

- 兼容接口：`/api/*` 与 `/de2api/*`
- 关键链路：`auth/role/dataVisualization/tree`
- 关键脚本：
  - `infra/scripts/check-compat-baseline.sh`
  - `apps/backend-go/scripts/compat-checks/run_auth_visualization_compat.sh`

## 日常巡检（开发/预发/生产）

在仓库根目录执行：

```bash
bash infra/scripts/check-compat-baseline.sh
cd apps/backend-go && ./scripts/compat-checks/run_auth_visualization_compat.sh
```

通过标准：

- 基线脚本：`PASS: 6` 且 `FAIL: 0`
- 严格脚本：`Compatibility checks passed: 18/18`
- 严格脚本中关键接口必须返回 `code=000000`

## 发布前最小检查

```bash
docker compose -f infra/compose/docker-compose.yml up -d
cd apps/backend-go && make build && make test && make drift-check
bash infra/scripts/check-compat-baseline.sh
cd apps/backend-go && ./scripts/compat-checks/run_auth_visualization_compat.sh
```

通过标准：所有命令退出码为 `0`。

## CI 门禁现状

- 独立工作流：`/.github/workflows/strict-compat-gate.yml`
- 触发范围：`apps/backend-go/**`
- 阻断策略：严格脚本失败即 CI 失败

## 故障分级与处置

### P0（核心链路大面积失败）

现象：登录、权限、可视化树大范围异常，或严格脚本多条失败。

动作：优先执行“全量回滚”。

### P1（局部链路失败）

现象：仅 `dataVisualization/tree` 或个别兼容端点失败。

动作：优先执行“局部回滚”。

## 回滚方案

### A. 全量回滚（回退整个合并）

```bash
git checkout main
git pull --ff-only origin main
git revert -m 1 <merge-commit-sha> --no-edit
git push origin main
```

### B. 局部回滚（仅可视化树严格校验）

```bash
git checkout -b rollback/visualization-tree-validation origin/main
git restore --source=<target-sha>^ -- apps/backend-go/internal/transport/http/handler/visualization_handler.go
git commit -m "revert(visualization): rollback strict tree validation guard"
git push -u origin rollback/visualization-tree-validation
```

### C. CI 行为回滚（不改生产功能）

```bash
git checkout -b rollback/strict-compat-ci origin/main
git restore --source=<target-sha>^ -- .github/workflows/strict-compat-gate.yml .github/workflows/e2e.yml
git commit -m "revert(ci): rollback strict compatibility workflow changes"
git push -u origin rollback/strict-compat-ci
```

## 回滚后验收

```bash
cd apps/backend-go && make build && make test
bash infra/scripts/check-compat-baseline.sh
cd apps/backend-go && ./scripts/compat-checks/run_auth_visualization_compat.sh
```

通过标准：构建/测试通过，两个脚本均通过。

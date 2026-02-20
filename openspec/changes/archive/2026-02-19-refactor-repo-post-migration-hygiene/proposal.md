# Change: 迁移后仓库整洁治理（激进清理）

## Why

目录迁移已完成，但仓库仍存在历史残留目录、旧路径文档、过时构建入口和资产归属不清的问题，持续增加维护成本并导致误用风险。

## What Changes

- 执行激进清理：清理根目录运行时残留与无效历史目录
- 统一构建与部署入口到新目录拓扑（`apps/`、`legacy/`、`infra/`）
- 清理或迁移大体量历史资产目录（`mapFiles/`、`drivers/`、`staticResource/`、`de-xpack`）
- 全量修正文档路径与开发指南，移除 `core/*`、`backend-go/*` 旧路径叙述
- 修复旧路径扫描门禁的误报规则，增加边界匹配

## Impact

- Affected specs:
  - `backend-go-architecture`
- Affected code and assets:
  - `Dockerfile`
  - `.gitignore`, `.dockerignore`, `.gitmodules`
  - `README.md`, `development_guide.md`, `docs/**`, `AGENTS.md`
  - `infra/compose/**`, `infra/scripts/**`
  - `mapFiles/**`, `drivers/**`, `staticResource/**`, `de-xpack`
  - 根目录运行时残留目录（`BOOT-INF/`, `dataease-data/`, `logs/`, `opt/`）

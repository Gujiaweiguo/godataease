# Change: 统一仓库目录拓扑（Go 后端主线 + Java 只读备份）

## Why

当前仓库同时存在 `backend-go/`、`core/core-frontend/`、`core/core-backend/`，目录语义与迁移状态不一致，导致开发入口、CI 路径和运维脚本认知成本高，且容易产生旧路径残留。

## What Changes

- 建立统一目录拓扑：`apps/`、`legacy/`、`infra/`、`docs/`
- 执行一次性路径切换（不保留旧路径兼容层）
- 明确 Java 后端为长期只读备份区（仅安全补丁/应急修复/迁移对照）
- 同批次重构 `infra/docs/scripts` 与 CI/CD 路径引用
- 采用 `tests-after` 回归策略作为迁移验收门禁
- 在 OpenSpec `tasks.md` 维护唯一执行计划（Plan v2）

## Impact

- Affected specs:
  - `backend-go-architecture`
- Affected code and assets:
  - `backend-go/**` -> `apps/backend-go/**`
  - `core/core-frontend/**` -> `apps/frontend/**`
  - `core/core-backend/**` -> `legacy/backend-java/**`
  - `.github/workflows/*.yml`
  - `docker-compose.yml`
  - `scripts/**`
  - `README.md`, `development_guide.md`, `AGENTS.md`

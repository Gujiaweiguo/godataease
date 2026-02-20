## 1. Inventory and Decision Freeze

- [x] 1.1 生成冗余目录与文档旧路径基线清单（含体积、引用数、保留决策）
- [x] 1.2 冻结资产处置决策：`mapFiles/`、`drivers/`、`staticResource/`、`de-xpack`（迁移/删除）

## 2. Root Workspace Cleanup (Aggressive)

- [x] 2.1 清理根目录运行时残留：`BOOT-INF/`、`dataease-data/`、`logs/`、`opt/`
- [x] 2.2 清理本地构建残留策略：`apps/frontend/node_modules/`、`apps/frontend/dist/`（保留为本地产物，不入仓）
- [x] 2.3 修订 `.gitignore`/`.dockerignore` 以匹配新目录拓扑并移除旧路径规则

## 3. Asset Relocation or Removal

- [x] 3.1 迁移或删除 `mapFiles/`，并在目标目录建立明确归属说明
- [x] 3.2 迁移或删除 `drivers/`，如仅 Java 使用则并入 `legacy/backend-java/` 语义域
- [x] 3.3 迁移或删除 `staticResource/`，统一到前端或运行时资产目录
- [x] 3.4 处理空子模块 `de-xpack`（保留则补说明，删除则清理 `.gitmodules`）

## 4. Build/Deploy Entry Consistency

- [x] 4.1 修正 `Dockerfile` 的构建输入路径与运行时策略，确保不引用旧路径
- [x] 4.2 统一 Compose 与脚本入口到 `infra/compose`、`infra/scripts`
- [x] 4.3 处理 `redis-dataease-compose.yml`（迁移到 `infra/compose` 或删除）并去除硬编码敏感值

## 5. Documentation and Governance Cleanup

- [x] 5.1 全量修正文档旧路径：`README.md`、`development_guide.md`、`docs/**`、`AGENTS.md`
- [x] 5.2 修复 `infra/scripts/scan-old-paths.sh` 子串误报，改为边界匹配并更新 allowlist
- [x] 5.3 更新迁移后目录规范说明，明确"可写区/只读区/运行时目录"

## 6. Verification and Gate

- [x] 6.1 执行路径残留扫描并输出报告（阻塞级=0）
- [x] 6.2 执行 tests-after 回归（Go 构建测试、前端检查、关键脚本 smoke）
- [x] 6.3 完成 OpenSpec 验证：`openspec validate refactor-repo-post-migration-hygiene --strict --no-interactive`

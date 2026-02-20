## ADDED Requirements

### Requirement: Post-migration Repository Hygiene

系统 SHALL 在目录迁移完成后执行仓库整洁治理，清理冗余目录、统一资产归属并移除旧路径歧义。

#### Scenario: Root redundancy cleanup
- **WHEN** 执行迁移后整洁治理
- **THEN** 系统 SHALL 清理根目录运行时残留与历史临时目录，避免与源代码并存

#### Scenario: Legacy path elimination
- **WHEN** 执行文档和脚本治理
- **THEN** 系统 SHALL 消除 `core/*` 与旧 `backend-go/*` 等过时路径引用（归档记录除外）

### Requirement: Large Asset Governance

系统 SHALL 对大体量历史资产目录建立明确处置策略（迁移或删除），并以引用验证为前置条件。

#### Scenario: Asset relocation with reference validation
- **WHEN** 处置 `mapFiles/`、`drivers/`、`staticResource/`、`de-xpack`
- **THEN** 系统 SHALL 在迁移或删除前完成引用扫描，并在处置后通过命令级验证确认无阻塞回归

### Requirement: Canonical Build and Infra Entry

系统 SHALL 统一构建与部署入口到迁移后的目录拓扑，禁止旧入口继续作为事实来源。

#### Scenario: Canonical entry enforcement
- **WHEN** 开发或运维执行构建/部署
- **THEN** 系统 SHALL 通过 `infra/` 与新目录入口执行，且 `Dockerfile`/Compose/脚本不再依赖旧目录结构

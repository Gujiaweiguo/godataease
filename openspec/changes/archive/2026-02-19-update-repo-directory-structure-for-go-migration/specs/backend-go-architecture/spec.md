## MODIFIED Requirements

### Requirement: Single Authoritative Execution Plan

系统 SHALL 在 OpenSpec 中维护唯一执行计划，作为 `backend-go-architecture` 相关变更的执行事实来源。

#### Scenario: Plan authority for repository restructuring
- **WHEN** 执行仓库目录统一重构
- **THEN** 执行系统 SHALL 仅依据 `openspec/changes/update-repo-directory-structure-for-go-migration/tasks.md` 中 Plan v2 执行

#### Scenario: Unplanned task rejection
- **WHEN** 存在未在 Plan v2 中声明的执行项
- **THEN** 执行系统 SHALL 拒绝执行该任务，直到 Plan v2 更新并通过评审

#### Scenario: Stale plan rejection
- **WHEN** 存在与 Plan v2 冲突的外部计划文件或口头任务
- **THEN** 执行系统 SHALL 以 Plan v2 为唯一依据并阻断冲突执行

### Requirement: Task Metadata Completeness

系统 SHALL 要求 Plan v2 中每个任务包含完整执行元数据。

#### Scenario: Required task fields
- **WHEN** 定义或更新任务
- **THEN** 每个任务 SHALL 包含任务ID、输入、输出、验收标准、回滚方案、依赖关系和风险等级

#### Scenario: Dependency and risk traceability
- **WHEN** 查询执行计划
- **THEN** 系统 SHALL 能够明确展示任务依赖顺序和风险等级分布

#### Scenario: Command-level verifiability
- **WHEN** 任务进入验收阶段
- **THEN** 任务 SHALL 提供可执行的命令级验证方法，避免仅人工主观确认

## ADDED Requirements

### Requirement: Repository Directory Topology Governance

系统 SHALL 采用统一目录拓扑治理迁移后仓库结构：`apps/`、`legacy/`、`infra/`、`docs/`。

#### Scenario: Canonical directory mapping
- **WHEN** 执行目录重构
- **THEN** 系统 SHALL 将 Go 后端映射至 `apps/backend-go/`，前端映射至 `apps/frontend/`，Java 后端备份映射至 `legacy/backend-java/`

#### Scenario: Legacy Java read-only governance
- **WHEN** 迁移完成后维护 Java 备份
- **THEN** 系统 SHALL 将 `legacy/backend-java/` 视为只读区域，仅允许安全补丁、应急修复和迁移对照类改动

### Requirement: Immediate Path Cutover Governance

系统 SHALL 执行一次性路径切换，不保留旧路径兼容层。

#### Scenario: One-shot path migration
- **WHEN** 冻结窗口内执行切换
- **THEN** 系统 SHALL 在同一批次完成 CI、脚本、部署与文档路径改写

#### Scenario: Residual old-path detection
- **WHEN** 切换完成后执行仓库扫描
- **THEN** 系统 SHALL 发现并阻断阻塞级旧路径残留（允许名单除外）

### Requirement: Tests-after Regression Gate for Directory Migration

系统 SHALL 在目录切换后执行 tests-after 回归门禁，确认新路径下构建、检查与关键脚本可用。

#### Scenario: Go backend post-migration verification
- **WHEN** 目录切换完成
- **THEN** 系统 SHALL 在 `apps/backend-go/` 成功执行构建与测试命令

#### Scenario: Frontend post-migration verification
- **WHEN** 目录切换完成
- **THEN** 系统 SHALL 在 `apps/frontend/` 成功执行类型检查与 lint 命令

#### Scenario: Cutover acceptance gate
- **WHEN** tests-after 结果存在阻塞级失败
- **THEN** 系统 SHALL 阻断解冻并触发回滚流程

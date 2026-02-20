## Context

- 迁移主线已切换为 Go 后端，但仓库仍保留 Java/前端历史路径。
- 用户已确认策略：
  - Java 后端长期只读备份
  - 立即切换（不保留旧路径）
  - 主体目录 + `infra/docs/scripts` 一次到位深度重构
  - 冻结窗口切换
  - 测试策略为 `tests-after`

## Goals / Non-Goals

- Goals:
  - 建立可长期维护的目录语义边界
  - 一次性消除旧路径歧义
  - 将执行计划唯一化并固化在 OpenSpec `tasks.md`
  - 在冻结窗口内完成切换与回归，具备可执行回滚
- Non-Goals:
  - 不在本变更中实现新业务功能
  - 不调整 API 业务语义
  - 不引入 Java 与 Go 双写/双活机制

## Decisions

- Decision: 采用目标拓扑 `apps/`, `legacy/`, `infra/`, `docs/`
  - Why: 与运行态、历史态、运维资产、文档资产职责一致
- Decision: Java 目录落位 `legacy/backend-java`，并声明只读
  - Why: 保留迁移对照与应急兜底，同时阻断常规业务回流
- Decision: 立即切换，不保留兼容软链接
  - Why: 快速收敛路径事实源，避免长期双路径维护成本
- Decision: tests-after
  - Why: 当前前端/Java 测试基础设施不完整，优先保证切换窗口内可控落地

## Risks / Trade-offs

- 风险: 一次性切换导致遗漏路径引用
  - Mitigation: 增加旧路径残留扫描门禁（CI + 本地命令）
- 风险: 构建/部署脚本同时改动面大
  - Mitigation: 按波次执行，先移动后改引用，最后统一回归
- 风险: Java 只读策略执行不一致
  - Mitigation: 在约定文件中显式声明 + CODEOWNERS/流程门禁（实现任务中落地）

## Migration Plan

1. 冻结窗口开始，建立 baseline 快照与回滚锚点
2. 创建目标目录并完成三大主体迁移（Go/前端/Java 备份）
3. 同批次改写 CI、脚本、compose、文档路径
4. 执行 tests-after 回归与旧路径扫描
5. 通过门禁后解冻；失败则按回滚步骤恢复

## Rollback Plan

- 触发条件：任一 P0 验证失败（构建、关键 CI、关键脚本、旧路径残留）
- 回滚步骤：
  1. 回退到冻结窗口前 tag/commit
  2. 恢复旧目录树与关键配置文件
  3. 重跑最小可用验证（Go 构建、前端类型检查、关键 CI 工作流语法检查）
- 成功标准：主分支恢复可构建、可发布、可追踪

## Open Questions

- 无阻塞性开放问题；进入执行前审阅阶段。

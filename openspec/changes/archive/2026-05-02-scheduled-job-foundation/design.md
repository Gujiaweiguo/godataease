## Context

当前 Go 侧已经有一个最小调度封装：`apps/backend-go/internal/job/scheduler.go` 基于 `robfig/cron` 提供 `AddFunc`、`AddJob` 和 `AddDistributedFunc`，并通过 Redis `SETNX` 做 30 秒分布式锁；但它仍停留在“库封装”阶段，而不是“任务平台”阶段。当前代码没有统一任务注册面、没有稳定 job metadata、没有样例任务、没有结构化的执行结果语义，`apps/backend-go/internal/job/jobs/job.go` 仍为空文件，现有测试也只覆盖了 cron 基本触发与无 Redis 情况下的执行。

这带来两个直接约束。第一，后续 Report、Threshold、Data Filling 等能力虽然都需要调度，但现在没有统一位置去声明“有哪些任务、什么时候启用、拿不到锁算什么结果”。第二，`backend-go-architecture` 现有 Scheduled Tasks requirement 只要求“启动时注册定时任务”和“通过 Redis 分布式锁保证单节点执行”，没有把跳过、失败、观测、回滚这些运行时行为治理成可验证契约。

因此，这个 change 需要在不引入具体业务任务复杂度的前提下，把现有 scheduler 封装提升为可复用、可诊断、可回滚的统一底座。

## Goals / Non-Goals

**Goals:**
- 为 Go 后端建立统一的 scheduled job registry，让任务定义集中化而不是散落在各模块启动逻辑中。
- 冻结基础执行语义：任务执行结果至少区分 `success`、`skipped`、`failed`，其中分布式锁竞争归类为 `skipped`。
- 为任务提供稳定 metadata（job key、cron spec、description、enabled state），使启动、日志与运维回滚可追踪。
- 在不引入业务任务复杂度的情况下注册至少一个低风险样例任务，证明调度路径不是空壳。
- 将这些运行时行为反映到 OpenSpec 能力层，补强 `backend-go-architecture` 对 Scheduled Tasks 的要求。

**Non-Goals:**
- 不在本 change 中迁移 Report、Threshold、Data Filling 等业务任务。
- 不在本 change 中引入统一任务管理 UI 或外部 HTTP 管理 API。
- 不强制定义所有业务任务的自动重试策略；首阶段只治理记录和可诊断性。
- 不改变现有 Redis / cron 选型，也不引入新的调度外部依赖。

## Decisions

### 1. 引入 registry-first 模式，而不是继续直接调用 `AddFunc`
**Decision:** 增加统一 job registry，任务必须以结构化 metadata 注册，再由启动流程统一装配进 scheduler。

**Why:** 当前 `AddFunc`/`AddDistributedFunc` 只接收 `spec` 和执行函数，无法表达任务的业务标识、启停状态、说明信息或后续观测维度。registry-first 可以把“声明任务”和“启动任务”分开，为后续模块接入留下单一入口。

**Alternatives considered:**
- 继续让业务模块在各自初始化位置直接调用 `AddFunc`：实现快，但会把启停、日志、回滚语义分散到每个模块。
- 先引入数据库驱动的动态任务配置：能力更强，但超出本 change 的基础设施范围。

### 2. 保留现有 cron + Redis 组合，但把锁竞争显式治理为 `skipped`
**Decision:** 继续使用 `robfig/cron` 和 Redis `SETNX` 锁；当任务未拿到锁时，不记为失败，而是记录为 `skipped`。

**Why:** 当前 scheduler 已经采用该组合，变更基础选型没有必要。更重要的是，锁竞争本质上是部署拓扑下的去重结果，不应与业务失败混淆，否则后续任务统计和告警会失真。

**Alternatives considered:**
- 将锁竞争视为失败：会放大误报，并使多实例部署下的正常行为看起来像异常。
- 为每个任务单独定制锁策略：灵活但会破坏平台一致性，留待后续特定业务任务再扩展。

### 3. 平台首阶段只统一结果记录，不统一自动重试
**Decision:** 本 change 只规定统一执行包装和结果记录，不提供平台级自动重试编排。

**Why:** 不同任务的幂等性、成本和外部副作用差异很大。现在先统一 `success/skipped/failed` 结果、耗时和错误信息，已经足够支撑后续业务任务落地；自动重试应由后续 change 在理解具体任务语义后再加。

**Alternatives considered:**
- 直接加全局自动重试：会把平台和业务耦合过深，并可能掩盖真正的执行失败。

### 4. 样例任务必须低风险且可验证
**Decision:** 注册一个只做低风险观测或 housekeeping 的样例任务，用于验证注册、锁竞争、成功/失败/跳过日志路径。

**Why:** 当前 `jobs/job.go` 为空，说明“框架存在”还没有经过真实任务验证。样例任务必须证明平台可用，但不能修改核心业务数据或引入误伤风险。

**Alternatives considered:**
- 直接拿未来 Report/Threshold 任务做首个样例：会把业务复杂度带进基础设施 change。
- 使用纯 no-op 任务：验证价值太低，无法证明真实执行包装、日志和锁语义。

### 5. 启停与回滚优先走配置和注册面，而不是删除代码
**Decision:** 任务的开关和回滚优先通过 registry/配置禁用完成，而不是靠移除任务代码。

**Why:** 这样能把回滚路径保持在运行时层面，出现问题时可以快速退回“无任务启用”状态，同时保留代码与测试资产。

**Alternatives considered:**
- 出现问题时直接回退代码：更重，也不利于任务平台作为长期基础设施稳定演进。

## Risks / Trade-offs

- **[Risk] 现有 `scheduler.go` 只暴露简单函数注册，补 registry 后可能出现双入口并存** → **Mitigation:** 将 registry 定义为平台标准入口，并在 design/spec 中明确后续任务接入必须通过 registry。
- **[Risk] Redis 锁 TTL 固定 30 秒，不适配未来长任务** → **Mitigation:** 本 change 只冻结“单节点执行”基础语义，不承诺长任务续锁；后续业务任务如超过该窗口，再通过独立 change 扩展续锁或心跳策略。
- **[Risk] 样例任务选得太轻，无法证明平台真实可用** → **Mitigation:** 要求样例任务至少经过 scheduler 触发、锁包装和结果记录全路径，而不是纯 no-op。
- **[Risk] 先不做平台级自动重试会让部分任务恢复能力不足** → **Mitigation:** 明确这是有意收紧范围；首阶段优先确保结果可观测与安全回滚，再由业务域定义重试。

## Migration Plan

1. 在 `internal/job/` 中引入 registry 和任务元数据结构，不破坏现有 `Scheduler` 封装。
2. 将应用启动过程接到 registry 装配层，由该层统一决定注册哪些任务。
3. 为 scheduler 增加统一执行包装，输出 `success/skipped/failed` 结果与基础诊断信息。
4. 接入 1 个低风险样例任务，验证单实例与多实例锁竞争语义。
5. 补 runbook 和配置开关，确保可以在不删除代码的情况下回退到“无任务启用”状态。

回滚策略：保留 registry 与执行包装代码，但将样例任务和后续任务注册全部禁用，让运行时回到当前“框架存在但无任务”的安全基线。

## Open Questions

- registry 的 `enabled` 状态首阶段是静态配置驱动，还是允许从数据库/系统参数读取？
- 样例任务应该优先选择观测类、清理类，还是纯内部状态校验类？
- 统一执行日志首阶段落在普通结构化日志即可，还是需要同步抽象成持久化任务结果记录？

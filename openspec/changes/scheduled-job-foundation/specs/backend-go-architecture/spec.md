## MODIFIED Requirements

### Requirement: Scheduled Tasks

系统 SHALL 使用 `robfig/cron` 实现定时任务，并通过统一注册面治理任务声明、启停和运行时结果语义。

#### Scenario: Task Scheduling
- **WHEN** 服务启动
- **THEN** 系统 SHALL 通过集中式任务注册表装配并注册所有启用的定时任务

#### Scenario: Cron Expression
- **WHEN** 定义定时任务
- **THEN** 系统 SHALL 支持标准 cron 表达式语法
- **AND** 每个任务 SHALL 具有稳定的任务标识、调度表达式、描述信息和启用状态

#### Scenario: Distributed Lock
- **WHEN** 定时任务执行
- **THEN** 系统 SHALL 通过 Redis 分布式锁保证单节点执行
- **AND** 未获取到锁的执行尝试 MUST be classified as a skipped run instead of a failed run

#### Scenario: Runtime outcome diagnostics
- **WHEN** 定时任务完成一次执行尝试
- **THEN** 系统 SHALL 以可诊断的方式区分 `success`、`skipped` 和 `failed` 结果

#### Scenario: Registration rollback
- **WHEN** 运维需要回退定时任务启用状态
- **THEN** 系统 SHALL 支持通过禁用任务注册回到无任务启用的安全运行态，而不是要求删除任务代码

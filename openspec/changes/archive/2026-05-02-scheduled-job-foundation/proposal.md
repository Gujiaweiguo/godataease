## Why

当前 Go 后端已经具备基于 `robfig/cron` 和 Redis 分布式锁的调度基础设施，但没有任何真实任务注册与统一执行治理，导致 legacy Java 中依赖调度的能力在 Go 主线中没有可运行的承载面。现在先补齐统一任务平台，可以把后续 Report、Threshold、Data Filling 等 P0 迁移建立在同一执行底座上，避免各业务域重复定义锁、失败、跳过与回滚语义。

## What Changes

- Introduce a unified scheduled-job foundation for the Go backend, including centralized job registration, stable job metadata, and startup wiring.
- Define platform-level execution semantics for scheduled jobs, including `success`, `skipped`, and `failed` outcomes, plus deterministic handling for distributed-lock contention.
- Add baseline observability and rollback controls for scheduled jobs so operators can enable, disable, diagnose, and safely back out task registration.
- Register at least one low-risk sample job to prove the scheduler path is functional instead of remaining framework-only.
- Tighten the existing backend architecture contract so scheduled-task behavior is governed as a runtime platform capability rather than an implementation detail.

## Capabilities

### New Capabilities
- `job-scheduling-foundation`: Unified scheduled-job registration, execution, locking, and diagnostics contract for the Go backend.

### Modified Capabilities
- `backend-go-architecture`: Refine scheduled-task requirements so service startup, job registration, and distributed single-node execution have verifiable runtime behavior.

## Impact

- **Affected code**: `apps/backend-go/internal/job/`, backend startup/bootstrap wiring, Redis lock integration, logging/diagnostic paths, and related configuration surfaces.
- **Affected systems**: Go runtime task execution, Redis-backed distributed locking, and operator runbooks for task enable/disable and rollback.
- **APIs**: No immediate external HTTP API changes are required in this change; the focus is the execution foundation that later task-driven modules will depend on.
- **Dependencies**: Continues to rely on `robfig/cron` and Redis; later P0 migrations such as Report, Threshold, and Data Filling will consume this foundation.
- **Breaking changes**: None planned. Rollback remains available by disabling task registration and returning the runtime to a no-job state.

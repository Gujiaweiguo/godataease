## Context

The current Go backend already contains sync-adjacent pieces, including datasource compatibility routes such as `/datasource/syncApiTable`, `/datasource/syncApiDs`, and `/datasource/listSyncRecord`, plus an existing `seatunnel-sync-integration` spec. However, the frontend sync center is built around a separate module surface: sync datasource management, sync task authoring/execution, sync task log inspection, and sync summary dashboards. Those `/sync/*` routes are currently absent from Go mainline, so frontend parity stops at the datasource compatibility bridge instead of reaching the operational sync workflows.

This change is cross-cutting because it touches HTTP contracts, orchestration semantics, persistence, and test coverage. It also depends on deterministic status mapping so that the frontend can treat Go and legacy behavior consistently.

## Goals / Non-Goals

**Goals:**
- Provide a first-class Go route surface for the frontend sync center.
- Reuse existing SeaTunnel-oriented sync orchestration where possible instead of inventing a parallel execution path.
- Standardize sync task lifecycle semantics for execute/start/stop/status/log flows.
- Make sync datasource, task, log, and summary APIs verifiable through backend tests and compatibility checks.

**Non-Goals:**
- Rebuild template market, AI, or other unrelated parity gaps.
- Change XPack/commercial feature boundaries.
- Replace existing `/datasource/syncApi*` compatibility routes during this change.
- Introduce a new non-SeaTunnel sync engine.

## Decisions

### 1. Introduce a dedicated sync transport surface instead of extending only datasource compatibility routes
The frontend already organizes sync flows under `/sync/datasource`, `/sync/task`, `/sync/task/log`, and `/sync/summary`. Mirroring that shape in Go keeps the frontend contract readable and avoids pushing task-management concerns into datasource handlers.

**Alternatives considered:**
- Keep expanding `/datasource/syncApi*` only: rejected because it does not cover task list, task log, and summary workflows cleanly.
- Proxy `/sync/*` to legacy Java services: rejected because the goal is Go-mainline parity, not a split runtime.

### 2. Reuse existing SeaTunnel orchestration and sync persistence as the backend execution backbone
The archived SeaTunnel change and current spec already establish SeaTunnel as the orchestration path for migration-critical sync workflows. The new sync module should sit on top of those primitives rather than re-implement task execution semantics.

**Alternatives considered:**
- Implement frontend-visible task state without persistent backend records: rejected because pager/log/summary endpoints require durable task metadata.
- Build a separate in-memory task manager: rejected because it would diverge from existing integration direction and be fragile across restarts.

### 3. Treat task logs and summary endpoints as first-class parity requirements, not optional follow-ons
The frontend sync center depends on log paging/detail/clear/termination and resource-count summary endpoints. These are operational APIs, not incidental extras, so they belong in the same change as task CRUD and lifecycle operations.

**Alternatives considered:**
- Ship datasource/task CRUD first and defer logs/summary: rejected because the frontend module would still be incomplete and hard to validate end-to-end.

### 4. Preserve existing datasource-side sync entrypoints while aligning their semantics with the new sync module
Existing compatibility routes such as `/datasource/syncApiTable`, `/datasource/syncApiDs`, and `/datasource/listSyncRecord` should continue to work, but they should not become the only public orchestration surface. The new sync module should share the same status and record model so behavior stays consistent.

**Alternatives considered:**
- Remove datasource-side compatibility routes immediately: rejected because that would expand migration risk and break existing callers.

## Risks / Trade-offs

- **[Risk] Sync module scope is broader than a single endpoint family** → Mitigation: keep the change focused on the four frontend route groups only, and exclude unrelated parity gaps.
- **[Risk] SeaTunnel availability and retry behavior can make task state nondeterministic** → Mitigation: define explicit status mapping and failure semantics in specs before implementation.
- **[Risk] Existing archived compatibility assumptions may be stale** → Mitigation: validate against current frontend API files and Go route registrations rather than relying only on archived matrices.
- **[Risk] Summary/log endpoints may require additional persistence not yet exposed cleanly** → Mitigation: make persistence and observability part of the change scope rather than post-implementation cleanup.

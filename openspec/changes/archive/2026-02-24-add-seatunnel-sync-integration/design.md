## Context
Java migration includes API datasource sync workflows that depend on SeaTunnel orchestration. Go handlers currently return placeholder statuses, without real task lifecycle management.

## Goals / Non-Goals
- Goals:
  - Integrate SeaTunnel task submission, status polling, and cancellation.
  - Provide stable sync record persistence for operational visibility.
  - Ensure compatibility endpoints return deterministic lifecycle semantics.
- Non-Goals:
  - Build a new sync engine in Go.
  - Change SeaTunnel cluster architecture.

## Decisions
- Decision: Keep SeaTunnel as external orchestrator and use gRPC client adapter.
- Decision: Persist sync task metadata locally for pagination and audit.
- Decision: Expose bounded polling behavior and deterministic state transitions.

## Alternatives Considered
- Fire-and-forget integration without record persistence: rejected due to observability and UX gaps.
- In-memory status tracking: rejected due to restart data loss.

## Risks / Trade-offs
- SeaTunnel availability directly affects sync start latency.
- Polling and persistence add operational complexity.

## Migration Plan
1. Implement gRPC submit/status/cancel methods.
2. Add sync record model/repository/service integration.
3. Replace compatibility placeholder endpoints with real orchestration.
4. Validate with contract tests and task lifecycle tests.

## Rollback
- Disable SeaTunnel-backed execution and return explicit unavailable status while preserving record reads.

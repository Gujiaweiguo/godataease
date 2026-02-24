## Context
Java-era SQL parsing and validation rely on Calcite behavior. The current Go client exists but method bodies are placeholders, causing migration parity gaps.

## Goals / Non-Goals
- Goals:
  - Provide deterministic Calcite parse/validate behavior in Go.
  - Ensure dataset SQL preview uses Calcite validation before execution.
  - Surface actionable, structured errors to callers.
- Non-Goals:
  - Re-implement Calcite engine in Go.
  - Expand SQL dialect features beyond current Java compatibility scope.

## Decisions
- Decision: Keep Calcite as external gRPC dependency and implement typed request/response adapters.
- Decision: Enforce bounded timeout and limited retries for transient network errors.
- Decision: Map upstream errors to stable Go compatibility error semantics.

## Alternatives Considered
- Local SQL parser in Go: rejected due to parity drift risk.
- No retry policy: rejected due to flaky upstream/network behavior.

## Risks / Trade-offs
- Upstream Calcite instability can affect preview availability.
- Retry policy increases tail latency but improves resilience.

## Migration Plan
1. Implement gRPC calls and adapters.
2. Integrate validation in service and compatibility paths.
3. Add tests and contract-diff evidence.
4. Roll out with observability and fallback behavior.

## Rollback
- Feature flag or configuration switch to bypass Calcite-dependent path and return explicit degraded-mode errors.

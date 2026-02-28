# Change: Add SeaTunnel Sync Integration

## Why
Go migration still lacks real SeaTunnel task orchestration. Current sync methods are placeholders, so API datasource sync cannot achieve Java parity.

## What Changes
- Implement SeaTunnel gRPC client methods for task submit, status query, and cancel.
- Define task lifecycle and status mapping semantics for compatibility endpoints.
- Persist sync records for pagination and operational observability.

## Impact
- Affected specs: `seatunnel-sync-integration`, `datasource-management`
- Affected code:
  - `apps/backend-go/internal/integration/seatunnel/seatunnel.go`
  - `apps/backend-go/internal/service/datasource_service.go`
  - `apps/backend-go/internal/transport/http/handler/compatibility_bridge_handler.go`
